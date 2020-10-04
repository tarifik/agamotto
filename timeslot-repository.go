package agamotto

import "sync"

// TimeSlotRepository is the interface
// to store timeslots of the events
// Timeslot is the time rounded to second
// in which we can store a lot of events
// Here we store event IDs that we get from
// entity repository when we save
// * Add(when, what uint64) – add an event data ID to event timeslot
// * Get(when uint64) []uint64 – get the events IDs for the special timeslot
// * Delete(when uint64) – delete timeslot and all the data
type TimeSlotRepository interface {
	Add(when, what uint64)
	Get(when uint64) []uint64
	Delete(when uint64)
}

// InMemoryTimeSlotRepository in-memory implementation of the TimeSlotRepository interface
type InMemoryTimeSlotRepository struct {
	mu      sync.RWMutex
	storage map[uint64][]uint64
	logger  Logger
}

// NewInMemoryTimeSlotRepository constructor
func NewInMemoryTimeSlotRepository(logger Logger) *InMemoryTimeSlotRepository {
	return &InMemoryTimeSlotRepository{
		mu:      sync.RWMutex{},
		storage: make(map[uint64][]uint64),
		logger:  logger,
	}
}

// Add an event data ID to event timeslot
func (i *InMemoryTimeSlotRepository) Add(when, what uint64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	_, ok := i.storage[when]
	if !ok {
		i.storage[when] = make([]uint64, 0)
	}
	i.storage[when] = append(i.storage[when], what)
	i.logger.Debugf("add event to timeslot %v, now there is %v IDs", when, len(i.storage[when]))
}

// Get the events IDs for the special timeslot
func (i *InMemoryTimeSlotRepository) Get(when uint64) []uint64 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	entities, ok := i.storage[when]
	if !ok {
		i.logger.Debugf("tried to get event for %v time slot, but nothing", when)
		return nil
	}
	i.logger.Debugf("found %v events for the timeslot %v", len(entities), when)
	return entities
}

// Delete timeslot and all the data
func (i *InMemoryTimeSlotRepository) Delete(when uint64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if entities, ok := i.storage[when]; ok {
		i.logger.Debugf("delete %v events for the timeslot %v", len(entities), when)
		delete(i.storage, when)
		return
	}
	i.logger.Debugf("tried delete events for the %v timeslot, but nothing", when)
}
