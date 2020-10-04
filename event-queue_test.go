package agamotto

import (
	"sync"
	"testing"
)

func TestTimeOrderedQueue(t *testing.T) {
	initSl := []uint64{3, 5}
	toq := NewHeapQueue(initSl, DefaultLogger(DEBUG, "event-queue"))
	toq.Add(0)
	toq.Add(6)
	toq.Add(0)
	toq.Add(onceUponATime)
	m, _ := toq.Next()
	if m != 0 {
		t.Fatalf("wrong max %v, %v", m, 0)
	}
	var m1, m2 uint64
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		m1, _ = toq.Next()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		m2, _ = toq.Next()
	}()
	wg.Wait()
	if m1 != 3 && m1 != 5 {
		t.Fatalf("wrong min %v, %v or %v", m1, 3, 5)
	}
	if m2 != 3 && m2 != 5 {
		t.Fatalf("wrong min %v, %v or %v", m2, 3, 5)
	}
	m, _ = toq.Next()
	if m != 6 {
		t.Fatalf("wrong min %v, %v", m, 6)
	}
	_, err := toq.Next()
	if err == nil {
		t.Fatalf("should be an error")
	}
}
