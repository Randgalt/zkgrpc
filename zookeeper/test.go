package zookeeper

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
	"sync"
	"sync/atomic"
)

func Test() {
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

	var lastSeenZxid int64 = 0
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			if in.Header != nil {
				atomic.StoreInt64(&lastSeenZxid, in.Header.Zxid)
			}

			if in.GetConnect() != nil {
				wg.Done()
			}
			log.Printf("Header: %v - Response: %v", in.Header, in.Detail)
		}
	}()

	var request = &RpcRequest{
		Header: &RpcRequestHeader{
			Xid: GetXid(),
		},
		Detail: &RpcRequest_Connect{Connect: &RpcConnectRequest{
			TimeOut: 100000,
		}},
	}
	err = stream.Send(request)

	wg.Wait()

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
