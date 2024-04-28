package g

const (
	Version = "3.0.1"
)

var (
	TransferCheck  = false
	RedisNormalHost []string
	TotalLru = make(map[int]LruCache, 1)
	SkipAlarm = false
)

// falcon.GetRedisHostsExpire  maintain  RedisNormalHost