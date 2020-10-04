package agamotto

import (
	"reflect"
	"sync/atomic"
	"testing"
)

func TestInMemoryEntityRepository(t *testing.T) {
	s := NewInMemoryEntityRepository(DefaultLogger(INFO, "entity-repo"))
	s.Add(&Entity{[]byte("hello1")})
	id2 := s.Add(&Entity{[]byte("hello2")})
	id3 := s.Add(&Entity{[]byte("hello3")})
	s.Delete(id2)
	if !reflect.DeepEqual(s.Get(id3), &Entity{[]byte("hello3")}) {
		t.Fatal("not expected result", s.Get(3))
	}
	if s.Get(id2) != nil {
		t.Fatal("not expected result", s.Get(2))
	}
	var maxID uint64
	for i := 0; i < 1_000; i++ {
		var storeID uint64
		go func() {
			id := s.Add(&Entity{[]byte("hello")})
			atomic.StoreUint64(&storeID, id)
		}()
		go func() {
			s.Delete(atomic.LoadUint64(&storeID) - 1)
		}()
		if maxID < atomic.LoadUint64(&storeID) {
			maxID = atomic.LoadUint64(&storeID)
		}
	}
	if s.Get(maxID) == nil {
		t.Fatal("must not be nil")
	}
}
