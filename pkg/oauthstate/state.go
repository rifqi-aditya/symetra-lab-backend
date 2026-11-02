package oauthstate

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type stateEntry struct {
	expiresAt time.Time
}

// Manager mengelola token state OAuth yang aman dari serangan CSRF
type Manager struct {
	mu     sync.RWMutex
	states map[string]stateEntry
	ttl    time.Duration
}

func NewManager(ttl time.Duration) *Manager {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	m := &Manager{
		states: make(map[string]stateEntry),
		ttl:    ttl,
	}

	// Periodic cleanup untuk menghapus state yang sudah kedaluwarsa
	go m.cleanupLoop(5 * time.Minute)
	return m
}

// GenerateState membuat token acak 32 bytes (64 hex characters) dan menyimpannya sementara
func (m *Manager) GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := hex.EncodeToString(b)

	m.mu.Lock()
	m.states[state] = stateEntry{
		expiresAt: time.Now().Add(m.ttl),
	}
	m.mu.Unlock()

	return state, nil
}

// ValidateAndConsume memvalidasi apakah state valid dan belum kedaluwarsa, lalu langsung menghapusnya (one-time token)
func (m *Manager) ValidateAndConsume(state string) bool {
	if state == "" {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.states[state]
	if !exists {
		return false
	}

	// Selalu hapus setelah diperiksa (one-time use)
	delete(m.states, state)

	// Pastikan belum expired
	return time.Now().Before(entry.expiresAt)
}

func (m *Manager) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for k, v := range m.states {
			if now.After(v.expiresAt) {
				delete(m.states, k)
			}
		}
		m.mu.Unlock()
	}
}

