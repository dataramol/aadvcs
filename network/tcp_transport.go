package network

import (
	"encoding/gob"
	"github.com/sirupsen/logrus"
	"net"
)

type NetAddr string

func (n NetAddr) String() string  { return string(n) }
func (n NetAddr) Network() string { return "tcp" }

type Peer struct {
	Conn       net.Conn
	Outbound   bool
	ListenAddr string
}

func (p *Peer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

func (p *Peer) ReadLoop(msgCh chan *Message) {
	for {
		msg := new(Message)
		if err := gob.NewDecoder(p.Conn).Decode(msg); err != nil {
			logrus.Errorf("decode message error: %s", err)
			break
		}
		msgCh <- msg
	}
	p.Conn.Close()
}

type TCPTransport struct {
	ListenAddress string
	Listener      net.Listener
	AddPeer       chan *Peer `json:"-"`
	DelPeer       chan *Peer `json:"-"`
}

func NewTCPTransport(addr string) *TCPTransport {
	return &TCPTransport{
		ListenAddress: addr,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	ln, err := net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}

	t.Listener = ln

	for {
		conn, err := ln.Accept()
		if err != nil {
			logrus.Error(err)
			continue
		}

		peer := &Peer{
			Conn:     conn,
			Outbound: false,
		}

		t.AddPeer <- peer
	}
}
