package zookeeper

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
	"sync"
)

func Test() {
	// standard gRPC methods to connect
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial("localhost:2181", opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := NewZooKeeperServiceClient(conn)
	stream, err := client.Process(context.Background())
	waitc := make(chan struct{})

	var wg sync.WaitGroup		// semaphore to wait for successful connect
	wg.Add(1)

	go func() {		// go channel for receiving responses
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			if in.Header != nil {
				SetLastSeenZxid(in.Header.Zxid)
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
			Xid: GetXid(),
		},
		Detail: &RpcRequest_Connect{Connect: &RpcConnectRequest{
			TimeOut: 100000,
		}},
	}
	err = stream.Send(request)

	wg.Wait()	// wait for connection success

	// send a ZNode create for /test
	request = &RpcRequest{
		Header: &RpcRequestHeader{
			Xid: GetXid(),
		},
		Detail: &RpcRequest_Create{Create: &RpcCreateRequest{
			Path: "/test",
			Acl:  OpenAllUnsafe,
		}},
	}
	err = stream.Send(request)

	// send a ZNode check-exists for /test with watcher
	request = &RpcRequest{
		Header: &RpcRequestHeader{
			Xid: GetXid(),
		},
		Detail: &RpcRequest_Exists{Exists: &RpcExistsRequest{
			Path: "/test",
			Watch:  true,
		}},
	}
	err = stream.Send(request)

	// send a ZNode delete for /test
	request = &RpcRequest{
		Header: &RpcRequestHeader{
			Xid: GetXid(),
		},
		Detail: &RpcRequest_Delete{Delete: &RpcDeleteRequest{
			Path:    "/test",
			Version: -1,
		}},
	}
	err = stream.Send(request)

	<-waitc
}
