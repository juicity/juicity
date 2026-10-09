package server

import (
	"bytes"
	"testing"
	"time"

	"github.com/daeuniverse/outbound/protocol"
	"github.com/daeuniverse/outbound/protocol/juicity"
	"github.com/daeuniverse/outbound/protocol/trojanc"
	"github.com/daeuniverse/outbound/protocol/tuic"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func TestAuthenticationPreservesUnderlayReadAhead(t *testing.T) {
	logger := zerolog.Nop()
	id := uuid.New()
	s := &Server{logger: &logger, users: map[uuid.UUID]string{id: "password"}, inFlightUnderlayKey: NewInFlightUnderlayKey(time.Second)}
	conn := testQUICConnection(t, s)
	token, err := tuic.GenToken(conn.ConnectionState(), id, "password")
	if err != nil {
		t.Fatal(err)
	}
	var flight bytes.Buffer
	if err := tuic.NewAuthenticate(id, token, juicity.Version0).WriteTo(&flight); err != nil {
		t.Fatal(err)
	}
	auth := &juicity.UnderlayAuth{IV: bytes.Repeat([]byte{1}, juicity.CipherConf.SaltLen), Psk: bytes.Repeat([]byte{2}, juicity.CipherConf.KeyLen), Metadata: &trojanc.Metadata{Metadata: protocol.Metadata{Hostname: "127.0.0.1", Port: 53}, Network: "udp"}}
	record := auth.PackFromPool()
	flight.Write(record)
	record.Put()
	stream, err := conn.OpenUniStream()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stream.Write(flight.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := stream.Close(); err != nil {
		t.Fatal(err)
	}
	stored := s.inFlightUnderlayKey.Wait(inFlightKey(auth.IV))
	if stored == nil || !bytes.Equal(stored.Psk, auth.Psk) || stored.Metadata.Hostname != "127.0.0.1" || stored.Metadata.Port != 53 {
		t.Fatalf("stored credential = %+v, want same-flight underlay record", stored)
	}
}
