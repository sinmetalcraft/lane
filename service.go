package lane

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Service struct {
	lanes map[string]*Buzzer
	mu    sync.Mutex
}

func NewService(ctx context.Context) (*Service, error) {
	return &Service{
		lanes: make(map[string]*Buzzer),
	}, nil
}

type Buzzer struct {
	GoSign chan bool
	Expire time.Time
	once   sync.Once
}

func (s *Service) LaneUp(ctx context.Context, key string, timeout time.Duration) (buzzer *Buzzer, goSign bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	buzzer, ok := s.lanes[key]
	if !ok {
		buzzer := &Buzzer{
			GoSign: make(chan bool),
			Expire: time.Now().Add(timeout),
		}
		s.lanes[key] = buzzer
		return buzzer, true
	}

	return buzzer, false
}

func (s *Service) Done(ctx context.Context, key string) (err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	buzzer, ok := s.lanes[key]
	if !ok {
		return fmt.Errorf("not found in the wait lane")
	}

	buzzer.once.Do(func() {
		close(buzzer.GoSign)
	})

	delete(s.lanes, key)
	return nil
}
