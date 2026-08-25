package session

import (
	"strings"
	"sync"
	"time"
)

const (
	loginWindow          = time.Minute
	loginAttemptsPerUser = 10
	loginAttemptsPerIP   = 30
	maxLoginKeys         = 10_000
	maxPasswordChecks    = 4
)

type loginWindowEntry struct {
	started time.Time
	count   int
}

type loginGuard struct {
	mu      sync.Mutex
	entries map[string]loginWindowEntry
	active  chan struct{}
	now     func() time.Time
}

func newLoginGuard() *loginGuard {
	return &loginGuard{entries: map[string]loginWindowEntry{}, active: make(chan struct{}, maxPasswordChecks), now: time.Now}
}

func (g *loginGuard) allow(ip, username string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()
	if len(g.entries) >= maxLoginKeys {
		for key, entry := range g.entries {
			if now.Sub(entry.started) >= loginWindow {
				delete(g.entries, key)
			}
		}
	}
	return g.increment(now, "user:"+strings.ToLower(username), loginAttemptsPerUser) &&
		g.increment(now, "ip:"+ip, loginAttemptsPerIP)
}

func (g *loginGuard) increment(now time.Time, key string, limit int) bool {
	entry, exists := g.entries[key]
	if !exists && len(g.entries) >= maxLoginKeys {
		return false
	}
	if !exists || now.Sub(entry.started) >= loginWindow {
		g.entries[key] = loginWindowEntry{started: now, count: 1}
		return true
	}
	if entry.count >= limit {
		return false
	}
	entry.count++
	g.entries[key] = entry
	return true
}

func (g *loginGuard) acquire() bool {
	select {
	case g.active <- struct{}{}:
		return true
	default:
		return false
	}
}

func (g *loginGuard) release() { <-g.active }
