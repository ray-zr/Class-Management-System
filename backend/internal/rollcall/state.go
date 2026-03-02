package rollcall

import "sync"

type State struct {
	mu      sync.Mutex
	roundID string
	fair    bool
}

func (s *State) Start(roundID string, fair bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.roundID = roundID
	s.fair = fair
}

func (s *State) Get() (roundID string, fair bool, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.roundID == "" {
		return "", false, false
	}
	return s.roundID, s.fair, true
}

func (s *State) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.roundID = ""
	s.fair = false
}
