package store

import (
	"encoding/json"
	"fmt"
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/judge/g"
	"github.com/signmem/falcon-plus/modules/judge/proc"
	"log"
	"sync"
)


// 全局读写锁：解决 expressionMap 并发读写崩溃 (修复falcon原生bug)
var (
	ExprLock sync.RWMutex // 新增
	MapLock sync.Mutex   // 公共锁
)


func Judge(L *SafeLinkedList, firstItem *model.JudgeItem, now int64) {
	CheckStrategy(L, firstItem, now)
	CheckExpression(L, firstItem, now)
}

func CheckStrategy(L *SafeLinkedList, firstItem *model.JudgeItem, now int64) {
	// 检查策略告警
	// 根据监控指标，找到对应的告警策略  Strategy

	// 拼接 key = 机器名/指标名
	// 从内存中获取所有策略
	// 找到这条监控对应的策略列表
	// 匹配标签（tag）
	// 匹配成功 ---> 进入真正判断逻辑

	key := firstItem.Endpoint + "/" + firstItem.Metric
	// key := fmt.Sprintf("%s/%s", firstItem.Endpoint, firstItem.Metric)

	// ========== 加读锁保护 ==========
	MapLock.Lock()
	defer MapLock.Unlock()

	strategyMap := g.StrategyMap.Get()
	strategies, exists := strategyMap[key]
	if !exists {
		return
	}

	for _, s := range strategies {
		// 因为key仅仅是endpoint和metric，所以得到的strategies并不一定是与当前judgeItem相关的
		// 比如lg-dinp-docker01.bj配置了两个proc.num的策略，一个name=docker，一个name=agent
		// 所以此处要排除掉一部分
		related := true
		for tagKey, tagVal := range s.Tags {
			if myVal, exists := firstItem.Tags[tagKey]; !exists || myVal != tagVal {
				related = false
				break
			}
		}

		if !related {
			continue
		}

		judgeItemWithStrategy(L, s, firstItem, now)
	}
}

func judgeItemWithStrategy(L *SafeLinkedList, strategy model.Strategy, firstItem *model.JudgeItem, now int64) {

	// 执行单条策略判断
	// 解析告警函数（avg、max、diff、lookup、all (diff) 等）
	// 执行计算 -->  判断是否触发告警
	// 构造告警事件
	// 判断是否需要发送告警

	fn, err := ParseFuncFromString(strategy.Func, strategy.Operator, strategy.RightValue)
	if err != nil {
		log.Printf("[ERROR] parse func %s fail: %v. strategy id: %d", strategy.Func, err, strategy.Id)
		return
	}

	historyData, leftValue, isTriggered, isEnough := fn.Compute(L)
	if !isEnough {
		return
	}

	event := &model.Event{
		Id:         fmt.Sprintf("s_%d_%s", strategy.Id, firstItem.PrimaryKey()),
		Strategy:   &strategy,
		Endpoint:   firstItem.Endpoint,
		LeftValue:  leftValue,
		EventTime:  firstItem.Timestamp,
		PushedTags: firstItem.Tags,
	}

	sendEventIfNeed(historyData, isTriggered, now, event, strategy.MaxStep)
}

func sendEvent(event *model.Event) {

	// 发送告警到 Redis
	// 把告警事件推送到 Redis，供 alarm 模块消费

	// update last event
	g.LastEvents.Set(event.Id, event)

	bs, err := json.Marshal(event)
	if err != nil {
		log.Printf("json marshal event %v fail: %v", event, err)
		return
	}

	// send to redis
	redisKey := fmt.Sprintf(g.Config().Alarm.QueuePattern, event.Priority())
	rc := g.RedisConnPool.Get()
	defer rc.Close()
	// rc.Do("LPUSH", redisKey, string(bs))
	_, err = rc.Do("LPUSH", redisKey, string(bs))

	proc.SendToRedisCntTotal.Incr()

	if err != nil {
		proc.SendToRedisCntDrop.Incr()
		log.Printf("[Error] push redis key: %s, value: %s, error %s",  redisKey, string(bs), err)
	} else {
		proc.SendToRedisCntSuccess.Incr()
	}

}

func CheckExpression(L *SafeLinkedList, firstItem *model.JudgeItem, now int64) {

	// 检查表达式告警
	// 处理更复杂的告警表达式  通用表达式告警
	// 支持多维度匹配
	// 支持更灵活的 tag 筛选
	// 支持按 endpoint、tag 键值对筛选
	// 给当前监控数据生成多个 key，用于匹配表达式

	keys := buildKeysFromMetricAndTags(firstItem)
	if len(keys) == 0 {
		return
	}

	// expression可能会被多次重复处理，用此数据结构保证只被处理一次
	handledExpression := make(map[int]struct{})

	// ========== 加读锁保护（崩溃点修复） ==========
	MapLock.Lock()
	defer MapLock.Unlock()

	expressionMap := g.ExpressionMap.Get()
	for _, key := range keys {
		expressions, exists := expressionMap[key]
		if !exists {
			continue
		}

		related := filterRelatedExpressions(expressions, firstItem)
		for _, exp := range related {
			if _, ok := handledExpression[exp.Id]; ok {
				continue
			}
			handledExpression[exp.Id] = struct{}{}
			judgeItemWithExpression(L, exp, firstItem, now)
		}
	}
}

func buildKeysFromMetricAndTags(item *model.JudgeItem) (keys []string) {
	for k, v := range item.Tags {
		keys = append(keys, fmt.Sprintf("%s/%s=%s", item.Metric, k, v))
	}
	keys = append(keys, fmt.Sprintf("%s/endpoint=%s", item.Metric, item.Endpoint))
	return
}

func filterRelatedExpressions(expressions []*model.Expression, firstItem *model.JudgeItem) []*model.Expression {

	// 过滤相关表达式
	// 从一堆表达式里，筛选出和当前监控数据相关的表达式

	size := len(expressions)
	if size == 0 {
		return []*model.Expression{}
	}

	exps := make([]*model.Expression, 0, size)

	for _, exp := range expressions {

		related := true

		itemTagsCopy := firstItem.Tags
		// 注意：exp.Tags 中可能会有一个endpoint=xxx的tag
		if _, ok := exp.Tags["endpoint"]; ok {
			itemTagsCopy = copyItemTags(firstItem)
		}

		for tagKey, tagVal := range exp.Tags {
			if myVal, exists := itemTagsCopy[tagKey]; !exists || myVal != tagVal {
				related = false
				break
			}
		}

		if !related {
			continue
		}

		exps = append(exps, exp)
	}

	return exps
}

func copyItemTags(item *model.JudgeItem) map[string]string {

	// 复制 item tags

	ret := make(map[string]string)
	ret["endpoint"] = item.Endpoint
	if item.Tags != nil && len(item.Tags) > 0 {
		for k, v := range item.Tags {
			ret[k] = v
		}
	}
	return ret
}

func judgeItemWithExpression(L *SafeLinkedList, expression *model.Expression, firstItem *model.JudgeItem, now int64) {

	// 执行表达式判断
	// 执行表达式告警计算，逻辑和策略判断一样


	fn, err := ParseFuncFromString(expression.Func, expression.Operator, expression.RightValue)
	if err != nil {
		log.Printf("[ERROR] parse func %s fail: %v. expression id: %d", expression.Func, err, expression.Id)
		return
	}

	historyData, leftValue, isTriggered, isEnough := fn.Compute(L)
	if !isEnough {
		return
	}

	event := &model.Event{
		Id:         fmt.Sprintf("e_%d_%s", expression.Id, firstItem.PrimaryKey()),
		Expression: expression,
		Endpoint:   firstItem.Endpoint,
		LeftValue:  leftValue,
		EventTime:  firstItem.Timestamp,
		PushedTags: firstItem.Tags,
	}

	sendEventIfNeed(historyData, isTriggered, now, event, expression.MaxStep)

}

func sendEventIfNeed(historyData []*model.HistoryData, isTriggered bool, now int64, event *model.Event, maxStep int) {

	// 核心告警判断
	// 第一次触发  --->  发送 PROBLEM 告警
	// 持续触发 -->  连续发送  最大次数 maxStep
	// 恢复正常  -->  发送 OK 告警
	// 防止频繁发送 -->  最小间隔控制

	lastEvent, exists := g.LastEvents.Get(event.Id)
	if isTriggered {
		event.Status = "PROBLEM"
		if !exists || lastEvent.Status[0] == 'O' {
			// 本次触发了阈值，之前又没报过警，得产生一个报警Event
			event.CurrentStep = 1

			// 但是有些用户把最大报警次数配置成了0，相当于屏蔽了，要检查一下
			if maxStep == 0 {
				return
			}

			sendEvent(event)
			return
		}
		//modified by vincent.zhang for ignoring to check max step based on config
		if g.Config().Alarm.CheckMaxStep {
			// 逻辑走到这里，说明之前Event是PROBLEM状态
			if lastEvent.CurrentStep >= maxStep {
				// 报警次数已经足够多，到达了最多报警次数了，不再报警
				return
			}
		}

		if historyData[len(historyData)-1].Timestamp <= lastEvent.EventTime {
			// 产生过报警的点，就不能再使用来判断了，否则容易出现一分钟报一次的情况
			// 只需要拿最后一个historyData来做判断即可，因为它的时间最老
			return
		}

		if now-lastEvent.EventTime < g.Config().Alarm.MinInterval {
			// 报警不能太频繁，两次报警之间至少要间隔MinInterval秒，否则就不能报警
			return
		}

		event.CurrentStep = lastEvent.CurrentStep + 1
		sendEvent(event)
	} else {
		// 如果LastEvent是Problem，报OK，否则啥都不做
		if exists && lastEvent.Status[0] == 'P' {
			event.Status = "OK"
			event.CurrentStep = 1
			sendEvent(event)
		}
	}
}
