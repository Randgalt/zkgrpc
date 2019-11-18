package zookeeper

import (
	"math"
	"sync"
	"sync/atomic"
)

var xid int32 = 1
var mutex = &sync.Mutex{}

// corresponds to ClientCnxn.getXid()
func GetXid() int32 {
	mutex.Lock()
	defer mutex.Unlock()
	// Avoid negative cxid values.  In particular, cxid values of -4, -2, and -1 are special and
	// must not be used for requests -- see SendThread.readResponse.
	// Skip from MAX to 1.
	if xid == math.MaxInt32 {
		xid = 1
	}
	var result = xid
	xid += 1
	return result
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

var lastSeenZxid int64 = 0	// would be used it reconnect is needed

func SetLastSeenZxid(zkid int64) {
	atomic.StoreInt64(&lastSeenZxid, zkid)
}

func GetLastSeenZxid() int64 {
	atomic.LoadInt64(&lastSeenZxid)
	return lastSeenZxid

}