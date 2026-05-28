package kfk

import (
	"time"
)

type MItemList struct {
	items []*MItem
}


func (s *MItemList) push() {
	Produce(s.items)  // write to kafka
	defer s.clear()
}

func (s *MItemList) merge(item *MItem) {
	s.items = append(s.items, item)
}

func (s *MItemList) clear() {
	s.items = s.items[:0]
}


func (s *MItemList) Start(itemCh chan *MItem,  duration time.Duration) {


	for {
		item := <- itemCh
		s.merge(item)
		if len(s.items) >= 100 {
			s.push()
		}
		time.Sleep(1*time.Millisecond)
	}
}
