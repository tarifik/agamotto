package agamotto

import (
	"container/heap"
	"fmt"
	"sync"
)

// EventQueue is the simple interface
// to manupulate ordering of the time events
// Time is represented as a uint64 numbers in seconds
type EventQueue interface {
	Add(when uint64)
	Next() (uint64, error)
}

// onceUponATime is a time in the future that never comes true
// it is needed for the waiting timer that never shots
const onceUponATime = 9999999999

// heapUniqueQueue is the inner struct
// for the container/heap implementation
// for that purpose here we have following methods
// * Len() int
// * Less(i, j int) bool
// * Swap(i, j int)
// * Push(x interface{})
// * Pop() interface{}
//
// For the uniqueness we use map and slice
// as suggested in the standart Set implementation
type heapUniqueQueue struct {
	s []uint64
	m map[uint64]struct{}
}

func NewHeapUniqueQueue() *heapUniqueQueue {
	return &heapUniqueQueue{
		s: make([]uint64, 0),
		m: make(map[uint64]struct{}),
	}
}

func (oq heapUniqueQueue) Len() int {
	len := len(oq.s)
	return len
}

func (oq heapUniqueQueue) Less(i, j int) bool {
	isLess := oq.s[i] < oq.s[j]
	return isLess
}

func (oq heapUniqueQueue) Swap(i, j int) {
	oq.s[i], oq.s[j] = oq.s[j], oq.s[i]
}

func (oq *heapUniqueQueue) Push(x interface{}) {
	xUint, typeCorrect := x.(uint64)
	if !typeCorrect {
		panic("it is not uint64 type")
	}
	_, exist := oq.m[xUint]
	if exist {
		return
	}
	oq.s = append(oq.s, xUint)
	oq.m[xUint] = struct{}{}
}

func (oq *heapUniqueQueue) Pop() interface{} {
	old := oq.s
	n := len(old)
	x := old[n-1]
	oq.s = old[0 : n-1]
	delete(oq.m, x)
	return x
}

// HeapQueue is the implementation of the EventQueue interface
// using Heap with inner heapUniqueQueue struct
type HeapQueue struct {
	mu     sync.Mutex
	h      *heapUniqueQueue
	logger Logger
}

// NewHeapQueue is the HeapQueue constructor
func NewHeapQueue(from []uint64, logger Logger) *HeapQueue {
	h := NewHeapUniqueQueue()
	for _, el := range from {
		h.Push(el)
	}
	heap.Init(h)
	logger.Debugf("initiate heap queue from slice with len %v", len(from))
	return &HeapQueue{sync.Mutex{}, h, logger}
}

// Add new event to HeapQueue
func (hq *HeapQueue) Add(when uint64) {
	if when == onceUponATime {
		hq.logger.Debugf("tried to add once upon a time event, skip")
		return
	}
	hq.mu.Lock()
	defer hq.mu.Unlock()
	hq.logger.Debugf("add event to heap queue %v", when)
	heap.Push(hq.h, when)
}

// Next get one new next event from the queue
// important that this event is deleted from queue
// once it was getted from
// if there is no events Next will return onceUponATime
func (hq *HeapQueue) Next() (uint64, error) {
	hq.mu.Lock()
	defer hq.mu.Unlock()
	if hq.h.Len() == 0 {
		hq.logger.Debugf("queue is empty, nothing to return")
		return onceUponATime, fmt.Errorf("the queue is empty")
	}
	when := heap.Pop(hq.h).(uint64)
	hq.logger.Debugf("get next event from queue: %v", when)
	return when, nil
}
