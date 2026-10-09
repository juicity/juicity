package server

import (
	"context"
	"sync"
	"time"

	"github.com/daeuniverse/outbound/protocol/juicity"
)

type inFlightKey = [juicity.UnderlaySaltLen]byte

type ContextCancel struct {
	Ctx    context.Context
	Cancel func()
	Ready  chan struct{}
}
type InFlightUnderlayKey struct {
	ttl    time.Duration
	mu     sync.Mutex
	m      map[inFlightKey]*juicity.UnderlayAuth
	notify map[inFlightKey]*ContextCancel
}

func NewInFlightUnderlayKey(ttl time.Duration) *InFlightUnderlayKey {
	return &InFlightUnderlayKey{
		ttl:    ttl,
		mu:     sync.Mutex{},
		m:      make(map[inFlightKey]*juicity.UnderlayAuth, 64),
		notify: make(map[inFlightKey]*ContextCancel, 64),
	}
}

// Lookup returns an unexpired credential without consuming it.
func (i *InFlightUnderlayKey) Lookup(k inFlightKey) *juicity.UnderlayAuth {
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.lookup(k)
}

func (i *InFlightUnderlayKey) lookup(k inFlightKey) *juicity.UnderlayAuth {
	if cc := i.notify[k]; cc == nil || cc.Ctx.Err() != nil {
		return nil
	}
	return i.m[k]
}

// Wait waits up to the TTL for a credential without consuming it.
func (i *InFlightUnderlayKey) Wait(k inFlightKey) *juicity.UnderlayAuth {
	i.mu.Lock()
	if auth := i.lookup(k); auth != nil {
		i.mu.Unlock()
		return auth
	}
	cc := i.notify[k]
	if cc == nil {
		cc = i.newEntry(k)
	}
	i.mu.Unlock()
	select {
	case <-cc.Ready:
	case <-cc.Ctx.Done():
	}
	return i.Lookup(k)
}

// Consume removes only the credential that was successfully verified.
func (i *InFlightUnderlayKey) Consume(k inFlightKey, auth *juicity.UnderlayAuth) bool {
	i.mu.Lock()
	defer i.mu.Unlock()
	if auth == nil || i.lookup(k) != auth {
		return false
	}
	cc := i.notify[k]
	delete(i.m, k)
	delete(i.notify, k)
	cc.Cancel()
	return true
}

func (i *InFlightUnderlayKey) newEntry(k inFlightKey) *ContextCancel {
	ctx, cancel := context.WithTimeout(context.Background(), i.ttl)
	cc := &ContextCancel{Ctx: ctx, Cancel: cancel, Ready: make(chan struct{})}
	i.notify[k] = cc
	context.AfterFunc(ctx, func() {
		i.mu.Lock()
		defer i.mu.Unlock()
		if i.notify[k] == cc {
			delete(i.notify, k)
			delete(i.m, k)
		}
	})
	return cc
}

func (i *InFlightUnderlayKey) Store(k [juicity.UnderlaySaltLen]byte, auth *juicity.UnderlayAuth) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if cc := i.notify[k]; cc != nil {
		if i.m[k] == nil {
			close(cc.Ready)
		}
		cc.Cancel()
	}
	i.newEntry(k)
	i.m[k] = auth
}
