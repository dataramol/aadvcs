package network

import "github.com/dataramol/aadvcs/crdt"

type Message struct {
	Payload any
	From    string
	Merge   bool
}

type BroadcastTo struct {
	To      []string
	Payload any
	Merge   bool
}

func NewMessage(from string, payload any, Merge bool) *Message {
	return &Message{
		From:    from,
		Payload: payload,
		Merge:   Merge,
	}
}

type Handshake struct {
	ListenAddr        string
	CurrentGraphState *crdt.LastWriterWinsGraph
}
