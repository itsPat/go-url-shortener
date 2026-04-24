package links

import (
	"context"
	"sync"
	"time"
)

type InMemory struct {
	mu    sync.Mutex
	links map[string]Link
}

// getLocked returns the Link for code. Caller must hold s.mu.
func (s *InMemory) getLocked(code string) (Link, error) {
	link, ok := s.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	return link, nil
}

func NewInMemory() *InMemory {
	return &InMemory{
		links: make(map[string]Link),
	}
}

func (s *InMemory) Create(ctx context.Context, code, url string) (Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.links[code]; exists {
		return Link{}, ErrCodeTaken
	}

	link := Link{Code: code, URL: url, CreatedAt: time.Now()}
	s.links[code] = link
	return link, nil
}

func (s *InMemory) Get(ctx context.Context, code string) (Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.getLocked(code)
}

func (s *InMemory) GetAndIncrement(ctx context.Context, code string) (Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	link, err := s.getLocked(code)
	if err != nil {
		return link, err
	}

	link.Hits++
	s.links[code] = link
	return link, nil
}

var _ Store = (*InMemory)(nil)
