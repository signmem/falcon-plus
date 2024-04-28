package proc

import (
	nproc "github.com/toolkits/proc"
)

// Rpc
var (
	RpcRecvCnt       = nproc.NewSCounterQps("RpcRecvCnt")
	WriteToDBCnt     = nproc.NewSCounterQps("WriteToDBCnt")
	WriteToDBFailCnt = nproc.NewSCounterQps("WriteToDBFailCnt")
	RpcKeyTooSmallCnt   = nproc.NewSCounterQps("RpcKeyTooSmallCnt")
	RpcKeyTooBigCnt     = nproc.NewSCounterQps("RpcKeyTooBigCnt")
	RcpCacheHistoryFullCnt = nproc.NewSCounterQps("RcpCacheHistoryFullCnt")
	RpcRecvGuegeCnt     = nproc.NewSCounterQps("RpcRecvGuegeCnt")
	RpcRecvCounterCnt   = nproc.NewSCounterQps("RpcRecvCounterCnt")
	WriteToRedisCnt  = nproc.NewSCounterQps("WriteToRedisCnt")
	WriteToRedisFailCnt  = nproc.NewSCounterQps("WriteToRedisFailCnt")
)

func GetAll() []interface{} {
	ret := make([]interface{}, 0)

	// rpc recv
	ret = append(ret, RpcRecvCnt.Get())
	ret = append(ret, WriteToDBCnt.Get())
	ret = append(ret, WriteToDBFailCnt.Get())
	ret = append(ret, RpcKeyTooSmallCnt.Get())
	ret = append(ret, RpcKeyTooBigCnt.Get())
	ret = append(ret, RcpCacheHistoryFullCnt.Get())
	ret = append(ret, RpcRecvGuegeCnt.Get())
	ret = append(ret, RpcRecvCounterCnt.Get())
	ret = append(ret, WriteToRedisCnt.Get())
	ret = append(ret, WriteToRedisFailCnt.Get())

	return ret
}
