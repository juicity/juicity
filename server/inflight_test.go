package server

import (
	"testing"
	"time"

	"github.com/daeuniverse/outbound/protocol/juicity"
)

func TestUnusedUnderlayCredentialExpires(t *testing.T) {
	for _, replace := range []bool{false, true} {
		t.Run(map[bool]string{false: "unused", true: "replaced"}[replace], func(t *testing.T) {
			i := NewInFlightUnderlayKey(20 * time.Millisecond)
			k := inFlightKey{1}
			i.Store(k, &juicity.UnderlayAuth{})
			if replace {
				i.Store(k, &juicity.UnderlayAuth{})
			}
			i.mu.Lock()
			cc := i.notify[k]
			i.mu.Unlock()
			<-cc.Ctx.Done()
			// Wait for the expiration callback, rather than assuming it ran with the timer.
			deadline := time.NewTimer(time.Second)
			defer deadline.Stop()
			tick := time.NewTicker(time.Millisecond)
			defer tick.Stop()
			for {
				i.mu.Lock()
				_, pending := i.notify[k]
				_, retained := i.m[k]
				i.mu.Unlock()
				if !pending {
					if retained {
						t.Fatal("expired unused credential retained")
					}
					if auth := i.Wait(k); auth != nil {
						t.Fatal("expired credential was usable")
					}
					return
				}
				select {
				case <-tick.C:
				case <-deadline.C:
					t.Fatal("expiration callback did not remove the entry")
				}
			}
		})
	}
}

func TestVerifiedCredentialCannotConsumeReplacement(t *testing.T) {
	i := NewInFlightUnderlayKey(time.Second)
	k := inFlightKey{1}
	first, replacement := &juicity.UnderlayAuth{}, &juicity.UnderlayAuth{}
	i.Store(k, first)
	verified := i.Lookup(k)
	i.Store(k, replacement)
	if i.Consume(k, verified) {
		t.Fatal("verification of an old credential consumed its replacement")
	}
	if got := i.Lookup(k); got != replacement {
		t.Fatal("replacement credential was lost")
	}
	if !i.Consume(k, replacement) {
		t.Fatal("verified replacement could not be consumed")
	}
	if i.Lookup(k) != nil {
		t.Fatal("consumed credential remained available")
	}
}
