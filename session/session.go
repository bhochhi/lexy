package session

import "sync"

// Session is shared data stored per clientId.
type Session struct {
	ClientID      string
	CurrentStepID string
	StepsState    map[string]*StepState
	Context       map[string]any
	Options       []Option
}

type StepState struct {
	RetryCount     int
	MaxRetries     int
	LastPromptIdx  int
	CapturedSlot   string
	ValidationDone bool
}

type Option struct {
	Text string
	URL  string
}

// Store is a simple in-memory map with a mutex

type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewStore() *Store {
	return &Store{sessions: map[string]*Session{}}
}

func (s *Store) Get(clientID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[clientID]
	return sess, ok
}

func (s *Store) Save(sess *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sess.ClientID] = sess
}
