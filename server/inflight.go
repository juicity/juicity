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

func (i *InFlightUnderlayKey) Evict(k [juicity.UnderlaySaltLen]byte) *juicity.UnderlayAuth {
	i.mu.Lock()
	cc, ok := i.notify[k]
	if !ok {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(i.ttl))
		cc = &ContextCancel{
			Ctx:    ctx,
			Cancel: cancel,
		}
		i.notify[k] = cc
		i.mu.Unlock()
		_ = context.AfterFunc(ctx, func() {
			i.mu.Lock()
			defer i.mu.Unlock()
			if i.notify[k] == cc {
				delete(i.notify, k)
			}
		})
		<-cc.Ctx.Done()
		i.mu.Lock()
		defer i.mu.Unlock()
	} else {
		delete(i.notify, k)
		defer cc.Cancel()
		defer i.mu.Unlock()
	}
	auth, ok := i.m[k]
	if !ok {
		return nil
	}
	delete(i.m, k)
	return auth
}

func (i *InFlightUnderlayKey) Store(k [juicity.UnderlaySaltLen]byte, auth *juicity.UnderlayAuth) {
	i.mu.Lock()
	defer i.mu.Unlock()
	cc, ok := i.notify[k]
	if !ok {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(i.ttl))
		cc = &ContextCancel{
			Ctx:    ctx,
			Cancel: cancel,
		}
		i.notify[k] = cc
		i.m[k] = auth
		_ = context.AfterFunc(ctx, func() {
			i.mu.Lock()
			defer i.mu.Unlock()
			if i.notify[k] == cc {
				delete(i.notify, k)
			}
		})
	} else {
		i.m[k] = auth
		delete(i.notify, k)
		cc.Cancel()
	}
}
