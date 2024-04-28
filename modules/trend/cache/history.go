package cache

import (
	"sync"
	"time"

	"github.com/open-falcon/falcon-plus/common/model"
	"github.com/open-falcon/falcon-plus/modules/trend/g"
	"github.com/open-falcon/falcon-plus/modules/trend/writer"
)

type NewFunction func(key int64, pk string, val *model.TrendItem) interface{}
type UpdateFunction func(item interface{}, val *model.TrendItem)
type ResultFunction func(item interface{}) *g.TrendResult

type ItemMap struct {
	sync.RWMutex
	NewFunc    NewFunction
	UpdateFunc UpdateFunction
	ResultFunc ResultFunction
	M          map[string]interface{}
}

func NewItemMap(newF NewFunction, updateF UpdateFunction, resultF ResultFunction) *ItemMap {
	return &ItemMap{NewFunc: newF, UpdateFunc: updateF, ResultFunc: resultF, M: make(map[string]interface{})}
}

func (this *ItemMap) Get(key string) (interface{}, bool) {
	this.RLock()
	defer this.RUnlock()
	val, ok := this.M[key]
	return val, ok
}

func (this *ItemMap) Set(key string, val interface{}) {
	this.Lock()
	defer this.Unlock()
	if val == nil {
		return
	}
	this.M[key] = val
}

func (this *ItemMap) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.M)
}

func (this *ItemMap) Delete(key string) {
	this.Lock()
	defer this.Unlock()
	delete(this.M, key)
}

func (this *ItemMap) BatchDelete(keys []string) {
	count := len(keys)
	if count == 0 {
		return
	}

	this.Lock()
	defer this.Unlock()
	for i := 0; i < count; i++ {
		delete(this.M, keys[i])
	}
}

func (this *ItemMap) DeleteAll() {
	this.Lock()
	defer this.Unlock()
	for key, _ := range this.M {
		delete(this.M, key)
	}
}

func (this *ItemMap) Update(key int64, pk string, val *model.TrendItem) {
	if val == nil {
		return
	}
	if item, exists := this.Get(pk); exists {
		this.UpdateFunc(item, val)
	} else {
		item := this.NewFunc(key, pk, val)
		if item != nil {
			this.Set(pk, item)
		}
	}
}

func (this *ItemMap) Flush(key int64) {
	this.RLock()
	defer this.RUnlock()
	var results []*g.TrendResult
	for _, item := range this.M {
		result := this.ResultFunc(item)
		if result != nil {
			results = append(results, result)
		}
	}
	writer.FlushToDB(key, results)
}

type CacheBigMap struct {
	sync.RWMutex
	CreateTime int64
	M          map[string]*ItemMap
}

func NewCacheBigMap(newF NewFunction, updateF UpdateFunction, resultF ResultFunction) *CacheBigMap {
	bigMap := &CacheBigMap{CreateTime: time.Now().Unix(), M: make(map[string]*ItemMap)}
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			bigMap.M[BigMapIndexArray[i]+BigMapIndexArray[j]] = NewItemMap(newF, updateF, resultF)
		}
	}
	return bigMap
}

func (this *CacheBigMap) GetCreateTime() int64 {
	this.RLock()
	defer this.RUnlock()
	return this.CreateTime
}

func (this *CacheBigMap) Get(key string) (*ItemMap, bool) {
	this.RLock()
	defer this.RUnlock()
	val, ok := this.M[key]
	return val, ok

}

func (this *CacheBigMap) Delete(key string) {
	this.Lock()
	defer this.Unlock()
	delete(this.M, key)
}

type CacheHistory struct {
	sync.RWMutex
	M map[int64]*CacheBigMap
}

func NewCacheHistory() *CacheHistory {
	return &CacheHistory{M: make(map[int64]*CacheBigMap)}
}

func (this *CacheHistory) Get(key int64) (*CacheBigMap, bool) {
	this.RLock()
	defer this.RUnlock()
	val, ok := this.M[key]
	return val, ok
}

func (this *CacheHistory) Set(key int64, val *CacheBigMap) *CacheBigMap {
	this.Lock()
	defer this.Unlock()
	bm, ok := this.M[key]
	if !ok {
		this.M[key] = val
		return val
	}
	return bm
}

func (this *CacheHistory) GetItem(key int64, pk string) (interface{}, bool) {
	big_map, ok := this.Get(key)
	if !ok || big_map == nil {
		return nil, false
	} else {
		item_map, ok := big_map.Get(pk[0:2])
		if !ok || item_map == nil {
			return nil, false
		} else {
			return item_map.Get(pk)
		}
	}
}

func (this *CacheHistory) Keys() []int64 {
	this.RLock()
	defer this.RUnlock()
	var keys []int64
	for k := range this.M {
		keys = append(keys, k)
	}
	return keys
}

func (this *CacheHistory) Len() int {
	this.RLock()
	defer this.RUnlock()
	return len(this.M)
}

func (this *CacheHistory) Delete(key int64) {
	this.Lock()
	defer this.Unlock()
	delete(this.M, key)
}
