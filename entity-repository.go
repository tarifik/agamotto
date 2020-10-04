package agamotto

import (
	"sync"
	"sync/atomic"
)

// Entity struct that holds all the data
// corresponding to the event
type Entity struct {
	data []byte
}

// EntityRepository interface of the entity repository
// * Add(entity *Entity) uint64 - returns an ID of the event (simple sequence)
// * Get(key uint64) *Entity – get the stored event by key
// * Delete(key uint64) – delete an event from repo by key
type EntityRepository interface {
	Add(entity *Entity) uint64
	Get(key uint64) *Entity
	Delete(key uint64)
}

// InMemoryEntityRepository in-memory EntityRepository implementation
type InMemoryEntityRepository struct {
	mu      sync.RWMutex
	seq     uint64
	storage map[uint64]*Entity
	logger  Logger
}

// NewInMemoryEntityRepository constructor
func NewInMemoryEntityRepository(logger Logger) *InMemoryEntityRepository {
	return &InMemoryEntityRepository{
		mu:      sync.RWMutex{},
		seq:     0,
		storage: make(map[uint64]*Entity),
		logger:  logger,
	}
}

// Add entity to repository
func (i *InMemoryEntityRepository) Add(entity *Entity) uint64 {
	id := atomic.AddUint64(&i.seq, 1)
	i.mu.Lock()
	i.storage[id] = entity
	i.mu.Unlock()
	i.logger.Debugf("add entity to repo and get id %v", id)
	return id
}

// Get entity from repository
func (i *InMemoryEntityRepository) Get(key uint64) *Entity {
	i.mu.RLock()
	defer i.mu.RUnlock()
	if entity, ok := i.storage[key]; ok {
		i.logger.Debugf("got entity with the key %v from repo", key)
		return entity
	}
	i.logger.Debugf("tried to get entity with id %v, but nothing", key)
	return nil
}

// Delete entity from repository
func (i *InMemoryEntityRepository) Delete(key uint64) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, ok := i.storage[key]; ok {
		delete(i.storage, key)
		i.logger.Debugf("delete entity with the key %v from repo", key)
		return
	}
	i.logger.Debugf("tried to delete entity with id %v, but nothing", key)
}
