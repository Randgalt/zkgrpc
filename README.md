## ZooKeeper gRPC Test

### Build gRPC ZooKeeper

```
git clone -b wip-grpc https://github.com/Randgalt/zookeeper.git zkgrpc
cd zkgrpc
mvn -DskipTests -Darguments="-DskipTests" package
cp conf/zoo_sample.cfg conf/zoo.cfg
export SERVER_JVMFLAGS=-Dzookeeper.serverCnxnFactory=org.apache.zookeeper.grpc.RpcServerCnxnFactory
bin/zkServer.sh start-foreground
```

### Build gRPC Go Client Test

In your Go directory...

```
git clone https://github.com/Randgalt/zkgrpc.git
cd zkgrpc
go get
go build main.go
go run zkgrpc
```

You should see some output

```
2019/11/18 09:04:04 Header: <nil> - Response: &{sessionId:72096894323720192 sessionTimeout:40000 }
2019/11/18 09:04:04 Header: xid:2 zxid:51  - Response: &{path:"/test" stat:<czxid:51 mzxid:51 ctime:1574085844640 mtime:1574085844640 pzxid:51 > }
2019/11/18 09:04:04 Header: xid:3 zxid:51  - Response: <nil>
2019/11/18 09:04:04 Header: xid:-1 zxid:-1  - Response: &{keeperState:SyncConnected eventType:NodeDeleted path:"/test" }
2019/11/18 09:04:04 Header: xid:4 zxid:52  - Response: <nil>
```
