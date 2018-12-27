package ui

import (
	"crypto/rand"
	"sync"

	"github.com/privatix/dappctrl/data"
)

// TokenMakeChecker is token maker and checker.
type TokenMakeChecker interface {
	Check(string) bool
	Make() (string, error)
}

// SimpleToken is a in memory random token.
type SimpleToken struct {
	token string
	mtx   *sync.RWMutex
}

// NewSimpleToken return simple token.
func NewSimpleToken() *SimpleToken {
	return &SimpleToken{mtx: &sync.RWMutex{}}
}

// Check returns true if given string matches stored.
func (t *SimpleToken) Check(s string) bool {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return s == t.token
}

// Make makes new random token.
func (t *SimpleToken) Make() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.token = string(data.FromBytes(b))
	return t.token, nil
}
