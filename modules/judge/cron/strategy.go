package cron

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/judge/g"
	"github.com/signmem/falcon-plus/modules/judge/store"
)

func SyncStrategies() {
	duration := time.Duration(g.Config().Hbs.Interval) * time.Second
	for {
		// 函数改名！避免递归死循环
		syncAllStrategies()
		syncExpression()
		syncFilter()
		time.Sleep(duration)
	}
}

func syncAllStrategies() {
	var strategiesResponse model.StrategiesResponse
	err := g.HbsClient.Call("Hbs.GetStrategies", model.NullRpcRequest{}, &strategiesResponse)
	if err != nil {
		log.Println("[ERROR] Hbs.GetStrategies:", err)
		return
	}

	rebuildStrategyMap(&strategiesResponse)
}

func rebuildStrategyMap(strategiesResponse *model.StrategiesResponse) {

	m := make(map[string][]model.Strategy)

	// 2. 遍历所有主机策略，填充这个 Map
	for _, hs := range strategiesResponse.HostStrategies {

		hostname := hs.Hostname

		if g.Config().Debug && hostname == g.Config().DebugHost {
			log.Println(hostname, "strategies:")
			bs, _ := json.Marshal(hs.Strategies)
			fmt.Println(string(bs))
		}

		for _, strategy := range hs.Strategies {
			key :=  hostname + "/" +  strategy.Metric
			s := strategy
			m[key] = append(m[key], s)
		}
	}

	store.MapLock.Lock()
	defer store.MapLock.Unlock()
	// 3. 原子替换全局 Map
	g.StrategyMap.ReInit(m)
}

func syncExpression() {
	var expressionResponse model.ExpressionResponse
	err := g.HbsClient.Call("Hbs.GetExpressions", model.NullRpcRequest{}, &expressionResponse)
	if err != nil {
		log.Println("[ERROR] Hbs.GetExpressions:", err)
		return
	}

	rebuildExpressionMap(&expressionResponse)
}

func rebuildExpressionMap(expressionResponse *model.ExpressionResponse) {
	m := make(map[string][]*model.Expression)
	for _, exp := range expressionResponse.Expressions {
		// ========== 关键修复：复制表达式 ==========
		e := exp

		for k, v := range exp.Tags {
			// key := fmt.Sprintf("%s/%s=%s", exp.Metric, k, v)
			key := exp.Metric + "/" +  k + "=" + v
			m[key] = append(m[key], e)
		}
	}

	// ========== 加写锁 ==========
	store.MapLock.Lock()
	defer store.MapLock.Unlock()
	g.ExpressionMap.ReInit(m)
}

func syncFilter() {
	m := make(map[string]string)

	//M map[string][]model.Strategy
	// 加读锁
	store.MapLock.Lock()
	strategyMap := g.StrategyMap.Get()
	store.MapLock.Unlock()

	for _, strategies := range strategyMap {
		for _, strategy := range strategies {
			m[strategy.Metric] = strategy.Metric
		}
	}

	//M map[string][]*model.Expression
	// 加读锁
	store.MapLock.Lock()
	expressionMap := g.ExpressionMap.Get()
	store.MapLock.Unlock()

	for _, expressions := range expressionMap {
		for _, expression := range expressions {
			m[expression.Metric] = expression.Metric
		}
	}

	g.FilterMap.ReInit(m)
}
