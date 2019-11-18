package zookeeper

import (
	"math"
	"sync"
)

var xid int32 = 1
var mutex = &sync.Mutex{}

func GetXid() int32 {
	mutex.Lock()
	defer mutex.Unlock()
	// Avoid negative cxid values.  In particular, cxid values of -4, -2, and -1 are special and
	// must not be used for requests -- see SendThread.readResponse.
	// Skip from MAX to 1.
	if xid == math.MaxInt32 {
		xid = 1;
	}
	var result = xid
	xid += 1
	return result;
}

var OpenAllUnsafe = []*RpcAcl{
	&RpcAcl{
		Perms: []RpcPerms{RpcPerms_All},
		Id: &RpcId{
			Scheme: "world",
			Id:     "anyone",
		},
	},
}
