package g

import (
	"sync"
	"time"

	"github.com/signmem/falcon-plus/common/model"
)

type SafeStrategyMap struct {
	sync.RWMutex
	// endpoint:metric => [strategy1, strategy2 ...]
	M map[string][]model.Strategy
}

type SafeExpressionMap struct {
	sync.RWMutex
	// metric:tag1 => [exp1, exp2 ...]
	// metric:tag2 => [exp1, exp2 ...]
	M map[string][]*model.Expression
}

type SafeEventMap struct {
	sync.RWMutex
	M map[string]*model.Event
}

type SafeFilterMap struct {
	sync.RWMutex
	M map[string]string
}

var (
	HbsClient     *SingleConnRpcClient
	StrategyMap   = &SafeStrategyMap{M: make(map[string][]model.Strategy)}
	ExpressionMap = &SafeExpressionMap{M: make(map[string][]*model.Expression)}
	LastEvents    = &SafeEventMap{M: make(map[string]*model.Event)}
	FilterMap     = &SafeFilterMap{M: make(map[string]string)}
)

func InitHbsClient() {
	HbsClient = &SingleConnRpcClient{
		RpcServers:  Config().Hbs.Servers,
		Timeout:     time.Duration(Config().Hbs.Timeout) * time.Millisecond,
		CallTimeout: time.Duration(Config().Hbs.CallTimeout) * time.Millisecond,
	}
}

/*
func (this *SafeStrategyMap) ReInit(m map[string][]model.Strategy) {
	this.Lock()
	defer this.Unlock()
	this.M = m
}
*/

func (this *SafeStrategyMap) ReInit(m map[string][]model.Strategy) {
    this.Lock()
    defer this.Unlock()

    // 关键优化：复用已有 map，不新建，大幅降低内存占用
    if this.M == nil {
        // 第一次初始化：创建新map
        this.M = make(map[string][]model.Strategy, len(m))
    } else {
        // 复用旧map：清空所有key，不创建新map
        for k := range this.M {
            delete(this.M, k)
        }
    }

    // 把新数据导入到复用的map里
    for k, v := range m {
        this.M[k] = v
    }
}



func (this *SafeStrategyMap) Get() map[string][]model.Strategy {
	this.RLock()
	defer this.RUnlock()
	return this.M
}

/*
func (this *SafeExpressionMap) ReInit(m map[string][]*model.Expression) {
	this.Lock()
	defer this.Unlock()
	this.M = m
}
*/

func (this *SafeExpressionMap) ReInit(m map[string][]*model.Expression) {
    this.Lock()
    defer this.Unlock()

    // 关键优化：复用已有 map，不新建，大幅降低内存占用
    if this.M == nil {
        // 第一次初始化：创建新map
        this.M = make(map[string][]*model.Expression, len(m))
    } else {
        // 复用旧map：清空所有key，不创建新map
        for k := range this.M {
            delete(this.M, k)
        }
    }

    // 把新数据导入到复用的map里
    for k, v := range m {
        this.M[k] = v
    }
}


func (this *SafeExpressionMap) Get() map[string][]*model.Expression {
	this.RLock()
	defer this.RUnlock()
	return this.M
}

func (this *SafeEventMap) Get(key string) (*model.Event, bool) {
	this.RLock()
	defer this.RUnlock()
	event, exists := this.M[key]
	return event, exists
}

func (this *SafeEventMap) Set(key string, event *model.Event) {
	this.Lock()
	defer this.Unlock()
	this.M[key] = event
}

/*
func (this *SafeFilterMap) ReInit(m map[string]string) {
	this.Lock()
	defer this.Unlock()
	this.M = m
}
*/

func (this *SafeFilterMap) ReInit(m map[string]string) {
    this.Lock()
    defer this.Unlock()

    // 关键优化：复用已有 map，不新建，大幅降低内存占用
    if this.M == nil {
        // 第一次初始化：创建新map
        this.M = make(map[string]string, len(m))
    } else {
        // 复用旧map：清空所有key，不创建新map
        for k := range this.M {
            delete(this.M, k)
        }
    }

    // 把新数据导入到复用的map里
    for k, v := range m {
        this.M[k] = v
    }
}


func (this *SafeFilterMap) Exists(key string) bool {
	this.RLock()
	defer this.RUnlock()
	if _, ok := this.M[key]; ok {
		return true
	}
	return false
}



