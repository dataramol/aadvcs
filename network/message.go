package network

import "github.com/dataramol/aadvcs/crdt"

type Message struct {
	Payload any
	From    string
}

type BroadcastTo struct {
	To      []string
	Payload any
}

func NewMessage(from string, payload any) *Message {
	return &Message{
		From:    from,
		Payload: payload,
	}
}

type Handshake struct {
	ListenAddr        string
	CurrentGraphState *crdt.LastWriterWinsGraph
}
