package agamotto

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestService_Serve(t *testing.T) {
	events := []*Event{
		{
			When:   uint64(time.Now().Unix()) + 1,
			Entity: &Entity{data: []byte(fmt.Sprintf("%v", uint64(time.Now().Unix())+1))},
		},
		{
			When:   uint64(time.Now().Unix()) + 2,
			Entity: &Entity{data: []byte(fmt.Sprintf("%v", uint64(time.Now().Unix())+2))},
		},
		{
			When:   uint64(time.Now().Unix()) + 4,
			Entity: &Entity{data: []byte(fmt.Sprintf("%v", uint64(time.Now().Unix())+4))},
		},
		{
			When:   uint64(time.Now().Unix()) + 1,
			Entity: &Entity{data: []byte(fmt.Sprintf("%v_2", uint64(time.Now().Unix())+1))},
		},
		{
			When:   uint64(time.Now().Unix()) + 4,
			Entity: &Entity{data: []byte(fmt.Sprintf("%v_2", uint64(time.Now().Unix())+4))},
		},
		{
			When:   uint64(time.Now().Unix()) + 3,
			Entity: &Entity{data: []byte(fmt.Sprintf("%v_w", uint64(time.Now().Unix())+3))},
		},
		{
			When:   uint64(time.Now().Unix()) + 1,
			Entity: &Entity{data: []byte(fmt.Sprintf("%v_w", uint64(time.Now().Unix())+1))},
		},
	}

	s, in, out := NewService(
		NewHeapQueue([]uint64{}, DefaultLogger(DEBUG, "event-queue")),
		NewInMemoryEntityRepository(DefaultLogger(DEBUG, "entity-repo")),
		NewInMemoryTimeSlotRepository(DefaultLogger(DEBUG, "timeslot-repo")),
		DefaultLogger(DEBUG, "agamotto"),
	)

	go s.Serve()
	go func() {
		for i, event := range events {
			if i == 5 || i == 6 {
				time.Sleep(1 * time.Second)
			}
			in <- event
		}
	}()
	go func() {
		time.Sleep(5 * time.Second)
		s.Cancel()
	}()
	for entity := range out {
		eventTime := strings.Split(string(entity.data), "_")[0]
		currentTime := fmt.Sprintf("%v", time.Now().Unix())
		if eventTime != currentTime {
			t.Fatalf("incorrect result, %v-%v, for event data %v", eventTime, currentTime, string(entity.data))
		}
	}
}
