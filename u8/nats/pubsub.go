package internal

import (
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

type PubSub struct {
	conn *nats.EncodedConn
}

func NewPubSub() PubSub {
	//conn, err := nats.Connect(addr, nats.UserInfo(user, password))
	conn, err := nats.Connect("nats://54.160.229.90:80")

	if err != nil {
		log.Fatal().Err(err).Msg("PubSub.Init")
	}

	enc, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)

	if err != nil {
		log.Fatal().Err(err).Msg("PubSub.NATS.Encoder")
	}

	return PubSub{enc}
}

func (p *PubSub) Subscribe(event string, handler any) *nats.Subscription {
	sub, err := p.conn.Subscribe(event, handler)

	if err != nil {
		log.Error().Err(err).Str("event", event).Msg("PubSub.Subscribe")
	}

	return sub
}

func (p *PubSub) Publish(event string, payload any) {
	err := p.conn.Publish(event, payload)

	if err != nil {
		log.Error().Err(err).Str("event", event).Msg("PubSub.Publish")
	}
}

func (p PubSub) JetStream() nats.JetStream {
	js, err := p.conn.Conn.JetStream()

	if err != nil {
		log.Error().Err(err).Msg("PubSub.JetStream")
	}

	return js
}

func (p PubSub) Close() {
	p.conn.Close()
}
