package main

import (
	"bytes"
	"log"
	"time"

	. "github.com/open-falcon/falcon-plus/common/backend_pool"
	. "github.com/open-falcon/falcon-plus/common/model"
	//. "github.com/open-falcon/falcon-plus/transfer/sender/conn_pool"
)

type TsdbSender struct {
	items []*TsdbItem
	// 批处理数量
	batch       int
	helper      *TsdbConnPoolHelper
	disableSend bool
}

func NewTsdbSender(tsdbServer string, disableSend bool) *TsdbSender {
	tsdbConnPoolHelper := NewTsdbConnPoolHelper(tsdbServer, 8, 8, 1000, 5000)
	return &TsdbSender{
		helper:      tsdbConnPoolHelper,
		batch:       20,
		disableSend: disableSend,
	}
}

func (s *TsdbSender) IsClear() bool {
	return len(s.items) == 0
}

func (s *TsdbSender) clear() {
	s.items = s.items[:0]
}

func (s *TsdbSender) merge(item *TsdbItem) {
	s.items = append(s.items, item)
}

func (s *TsdbSender) fire() error {
	if s.IsClear() {
		return nil
	}
	// 无论发送成功与否，清理缓存数据
	// TODO: 添加重试机制
	defer s.clear()

	var buf bytes.Buffer
	for _, item := range s.items {
		buf.WriteString(item.TsdbString())
		buf.WriteString("\n")
	}

	log.Printf("%s\n", buf.Bytes())

	// 用于开发模式下禁用发送数据
	if s.disableSend {
		return nil
	}

	err := s.helper.Send(buf.Bytes())
	return err
}

func (s *TsdbSender) Start(itemCh chan *TsdbItem, duration time.Duration) {
	ticker := time.NewTicker(duration)

	for {
		select {
		case <-ticker.C:
			// 达到一定时间执行发送
			_ = s.fire()
		case item := <-itemCh:
			s.merge(item)
			// 达到数量执行发送
			if len(s.items) == s.batch {
				_ = s.fire()
			}
		}
	}
}
