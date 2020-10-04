package agamotto

import (
	"sync"
	"time"
)

// Event struct holds info about event
// when and which data it contains
type Event struct {
	When   uint64
	Entity *Entity
}

// Service main agamotto service struct
type Service struct {
	// InChan is exported to client and used
	// to send events: entity with when in unix timestamp in seconds
	InChan chan *Event
	// OutChan is exported to client and used
	// to receive entities when the time has come
	OutChan chan *Entity

	// mutex to sync isServing flag
	mu sync.RWMutex
	// queue to store sorted timestamps and get the min
	eventQueue EventQueue
	// entity repository to get entity by key
	entityRepo EntityRepository
	// timeslot repo to get entity IDs from entity repo by timestamp
	timeSlotRepo TimeSlotRepository
	// newEvent is the signal channel to process new event
	newEvent chan uint64
	// cancel is the signal channel to process cancel event
	cancel chan struct{}
	// isServing – flag that service is serving now
	isServing bool
	// logger interface to use external loggers
	logger Logger
}

// NewService constructor
func NewService(eq EventQueue, er EntityRepository,
	ts TimeSlotRepository, logger Logger) (*Service, chan<- *Event, <-chan *Entity) {
	s := &Service{
		InChan:       make(chan *Event, 100),
		OutChan:      make(chan *Entity, 100),
		mu:           sync.RWMutex{},
		eventQueue:   eq,
		entityRepo:   er,
		timeSlotRepo: ts,
		newEvent:     make(chan uint64),
		cancel:       make(chan struct{}),
		isServing:    false,
		logger:       logger,
	}
	return s, s.InChan, s.OutChan
}

// addEvent internal method to add new event
// we add event in three step transaction:
// * add the entity to entity repo and get the ID
// * add the entity ID with the corresponging timestamp to timeslot repo
// * add timestamp to the eventQueue
func (s *Service) addEvent(event *Event) {
	if event.When < uint64(time.Now().Unix()) {
		s.logger.Errorf("new event %v is in the past, skip", event.When)
		return
	}
	key := s.entityRepo.Add(event.Entity)
	s.timeSlotRepo.Add(event.When, key)
	s.eventQueue.Add(event.When)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.isServing {
		s.newEvent <- event.When
	}
}

// Cancel serving
func (s *Service) Cancel() {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.isServing {
		s.cancel <- struct{}{}
	}
}

// Serve main function
func (s *Service) Serve() {

	var err error
	var nextEvent uint64
	var timer *time.Timer

	// get new events from chan and add it
	go func() {
		for event := range s.InChan {
			s.addEvent(event)
		}
	}()

	// cycle to proper start
	// if we get the timestamp from the past we skip
	// but we can start serving with empty queue using onceUponATime const
LOOP:
	for {
		nextEvent, err = s.eventQueue.Next()
		switch {
		case err != nil:
			s.logger.Infof("start serving with no events in queue")
			nextEvent = onceUponATime
			break LOOP
		case time.Duration(nextEvent-uint64(time.Now().Unix())) < 0:
			s.logger.Errorf("value %v is less than now, trying next", nextEvent)
		default:
			break LOOP
		}
	}

	// start serving
	s.mu.Lock()
	s.isServing = true
	s.mu.Unlock()
	timer = setTimer(nextEvent)
	for {
		select {
		// if the timer went off we work on corresponding entities
		case <-timer.C:
			eventIDs := s.timeSlotRepo.Get(nextEvent)
			go func() {
				for _, eventID := range eventIDs {
					s.mu.Lock()
					s.OutChan <- s.entityRepo.Get(eventID)
					s.mu.Unlock()
				}
			}()
			nextEvent, err = s.eventQueue.Next()
			if err != nil {
				s.logger.Warningf("queue is empty, now serving with no events in queue")
			}
			timer = setTimer(nextEvent)
		// if there is new event we check for the needed changing of the next event
		case newEvent := <-s.newEvent:
			if newEvent <= nextEvent {
				s.eventQueue.Add(nextEvent)
				s.logger.Infof("add new event to queue with changing next")
				nextEvent, _ = s.eventQueue.Next()
				if !timer.Stop() {
					<-timer.C
				}
				timer = setTimer(nextEvent)
			} else {
				s.logger.Infof("new event added without changing next")
			}
		// if there is a cancel event we clean up and go away
		case <-s.cancel:
			s.logger.Infof("cancel serving")
			if nextEvent != onceUponATime {
				s.eventQueue.Add(nextEvent)
			}
			s.mu.Lock()
			defer s.mu.Unlock()
			s.isServing = false
			if !timer.Stop() {
				<-timer.C
			}
			close(s.OutChan)
			return
		}
	}
}

// setTimer helper func to set the timer
// nearestStart is the unix time uint64 representation in seconds
func setTimer(nearestStart uint64) *time.Timer {
	return time.NewTimer(time.Duration(nearestStart-uint64(time.Now().Unix())) * time.Second)
}
