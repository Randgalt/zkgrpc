package zookeeper

import (
	"math"
	"sync"
	"sync/atomic"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
)

type ClientCnxn struct {
	conn ZooKeeperServiceClient
	stream ZooKeeperService_ProcessClient
	waitc chan struct{}
	xid int32
	mutex sync.Mutex
        lastSeenZxid int64 // would be used it reconnect is needed
}

func NewClientCnxn() (client *ClientCnxn) {
	client = new(ClientCnxn)
	client.xid = 1
        client.lastSeenZxid = 0
	return
}

func (client *ClientCnxn) GetXid() (result int32) {
        client.mutex.Lock()
        defer client.mutex.Unlock()
        // Avoid negative cxid values.  In particular, cxid values of -4, -2, and -1 are special and
        // must not be used for requests -- see SendThread.readResponse.
        // Skip from MAX to 1.

        if client.xid == math.MaxInt32 {
                client.xid = 1
        }

        result = client.xid
        client.xid++
        return
}

func (client *ClientCnxn) Connect() {
	// standard gRPC methods to connect
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:2181", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client.conn = NewZooKeeperServiceClient(conn)

	client.stream, err = client.conn.Process(context.Background())

	client.waitc = make(chan struct{})

	var wg sync.WaitGroup		// semaphore to wait for successful connect
	wg.Add(1)


	go func() {		// go channel for receiving responses
		for {
			in, err := client.stream.Recv()
			if err == io.EOF {
				// read done.
				close(client.waitc)
				return
			}
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			if in.Header != nil {
				client.SetLastSeenZxid(in.Header.Zxid)
			}

			if in.GetConnect() != nil {
				wg.Done()
			}

			// just print out responses for now
			log.Printf("Header: %v - Response: %v", in.Header, in.Detail)
		}
	}()

	// send connect request
	var request = &RpcRequest{
		Header: &RpcRequestHeader{
			Xid: client.GetXid(),
		},
		Detail: &RpcRequest_Connect{Connect: &RpcConnectRequest{
			TimeOut: 100000,
		}},
	}
	err = client.stream.Send(request)

	wg.Wait()	// wait for connection success

	// send a ZNode create for /test
	request = &RpcRequest{
		Header: &RpcRequestHeader{
			Xid: client.GetXid(),
		},
		Detail: &RpcRequest_Create{Create: &RpcCreateRequest{
			Path: "/test",
			Acl:  OpenAllUnsafe,
		}},
	}
	client.stream.Send(request)

	// send a ZNode check-exists for /test with watcher
	request = &RpcRequest{
		Header: &RpcRequestHeader{
			Xid: client.GetXid(),
		},
		Detail: &RpcRequest_Exists{Exists: &RpcExistsRequest{
			Path: "/test",
			Watch:  true,
		}},
	}
	client.stream.Send(request)

	// send a ZNode delete for /test
	request = &RpcRequest{
		Header: &RpcRequestHeader{
			Xid: client.GetXid(),
		},
		Detail: &RpcRequest_Delete{Delete: &RpcDeleteRequest{
			Path:    "/test",
			Version: -1,
		}},
	}
	client.stream.Send(request)

	<-client.waitc 
}


// corresponds to ZooDefs.Ids.OPEN_ACL_UNSAFE
var OpenAllUnsafe = []*RpcAcl{
	&RpcAcl{
		Perms: []RpcPerms{RpcPerms_All},
		Id: &RpcId{
			Scheme: "world",
			Id:     "anyone",
		},
	},
}


func (client *ClientCnxn) SetLastSeenZxid(zkid int64) {
	atomic.StoreInt64(&client.lastSeenZxid, zkid)
}

func (client *ClientCnxn) GetLastSeenZxid() (lastSeenZxid int64) {
	atomic.LoadInt64(&client.lastSeenZxid)
	return

}
