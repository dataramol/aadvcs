package network

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/dataramol/aadvcs/clock"
	"github.com/dataramol/aadvcs/crdt"
	"github.com/dataramol/aadvcs/models"
	"github.com/dataramol/aadvcs/utils"
	"github.com/sirupsen/logrus"
	"net"
)

type Server struct {
	Peers               map[string]*Peer
	ListenAddress       string
	AddPeer             chan *Peer       `json:"-"`
	DelPeer             chan *Peer       `json:"-"`
	MsgCh               chan *Message    `json:"-"`
	BroadcastCh         chan BroadcastTo `json:"-"`
	Transport           *TCPTransport
	LastWriterWinsGraph *crdt.LastWriterWinsGraph
}

func NewServer(listenAddr string) *Server {

	s := &Server{
		ListenAddress: listenAddr,
		Peers:         make(map[string]*Peer),
		AddPeer:       make(chan *Peer),
		DelPeer:       make(chan *Peer),
		MsgCh:         make(chan *Message),
		BroadcastCh:   make(chan BroadcastTo),
	}

	tr := NewTCPTransport(s.ListenAddress)
	s.Transport = tr

	tr.AddPeer = s.AddPeer
	tr.DelPeer = s.DelPeer
	return s
}

func (s *Server) Start() {
	go s.loop()

	logrus.WithFields(logrus.Fields{
		"port": s.ListenAddress,
	}).Info("Started New AAD Version Control System Node")

	err := s.Transport.ListenAndAccept()
	if err != nil {
		return
	}
}

func (s *Server) RegisterPeer(p *Peer) {
	s.Peers[p.ListenAddr] = p
}

func (s *Server) SendHandshake(p *Peer) error {
	hs := &Handshake{
		CurrentGraphState: s.LastWriterWinsGraph,
		ListenAddr:        s.ListenAddress,
	}

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(hs); err != nil {
		return err
	}

	return p.Send(buf.Bytes())
}

func (s *Server) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	peer := &Peer{
		Conn:     conn,
		Outbound: true,
	}
	s.AddPeer <- peer

	return s.SendHandshake(peer)
}

func (s *Server) loop() {
	for {
		select {
		case peer := <-s.DelPeer:
			logrus.WithFields(logrus.Fields{
				"addr": peer.Conn.RemoteAddr(),
			}).Info("new node disconnected")
			delete(s.Peers, peer.Conn.RemoteAddr().String())
		case peer := <-s.AddPeer:
			if err := s.handleNewPeer(peer); err != nil {
				logrus.Errorf("handle peer error: %s", err)
			}
		case msg := <-s.MsgCh:
			if err := s.handleMessage(msg); err != nil {
				logrus.Errorf("Error while handling msg : %s", err)
			}
		}
	}
}

func (s *Server) handleNewPeer(peer *Peer) error {
	s.RegisterPeer(peer)
	_, err := s.handshake(peer)
	if err != nil {
		peer.Conn.Close()
		delete(s.Peers, peer.Conn.RemoteAddr().String())
		return fmt.Errorf("%s:handshake with incoming node failed: %s", s.ListenAddress, err)
	}

	go peer.ReadLoop(s.MsgCh)

	if !peer.Outbound {
		if err := s.SendHandshake(peer); err != nil {
			peer.Conn.Close()
			delete(s.Peers, peer.Conn.RemoteAddr().String())

			return fmt.Errorf("failed to send handshake with peer : %s", err)
		}
	}

	logrus.WithFields(logrus.Fields{
		"peer":       peer.Conn.RemoteAddr(),
		"ListenAddr": peer.ListenAddr,
		"we":         s.ListenAddress,
	}).Info("handshake successfully: new node connected")

	//s.RegisterPeer(peer)

	return nil
}

func (s *Server) Broadcast(broadcastMsg BroadcastTo) error {
	msg := NewMessage(s.ListenAddress, broadcastMsg.Payload)

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, addr := range broadcastMsg.To {
		peer, ok := s.Peers[addr]
		fmt.Printf("Peers -> %v", s.Peers)
		fmt.Printf("Peer -> %v\n", peer)
		fmt.Printf("Ok ? %v\n", ok)
		if ok {
			go func(peer *Peer) {
				if err := peer.Send(buf.Bytes()); err != nil {
					logrus.Errorf("broadcast to peer error : %s", err)
				}
			}(peer)
		}
	}

	return nil
}

func (s *Server) handshake(p *Peer) (*Handshake, error) {
	hs := &Handshake{}
	if err := gob.NewDecoder(p.Conn).Decode(hs); err != nil {
		return nil, err
	}

	_, ok := s.LastWriterWinsGraph.Clock.Clock[hs.ListenAddr]
	if !ok {
		s.LastWriterWinsGraph.Clock.Clock[hs.ListenAddr] = hs.CurrentGraphState.Clock.Clock[hs.ListenAddr]
	}

	p.ListenAddr = hs.ListenAddr
	/**This for updating connections in network */
	ws := &models.WritableServer{}
	fp, err := utils.CreateOrOpenFileRWMode(utils.AadvcsNetworkConfigFilePath)
	err = utils.ClearFileContent(fp)
	ws.ListAddr = s.ListenAddress
	ws.Connections = append(ws.Connections, p.ListenAddr)

	data, err := json.Marshal(ws)
	if err != nil {
		return nil, err
	}
	fp.Write(data)
	fp.Close()

	return hs, nil
}

func (s *Server) handleMessage(msg *Message) error {
	fmt.Printf("%+v\n", msg.Payload)
	graph := msg.Payload.(crdt.LastWriterWinsGraph)

	fmt.Printf("Now Comparing Clock \n")
	eventOrder := s.LastWriterWinsGraph.Clock.Compare(graph.Clock)
	fmt.Printf("Event Order :- %+v", eventOrder)
	if eventOrder == clock.HappensAfter {
		logrus.Info("Incoming event is Latest")
	} else if eventOrder == clock.HappensBefore {
		logrus.Info("Incoming event is stale and would be discarded")
	} else if eventOrder == clock.CONCURRENT {
		logrus.Info("Conflict is there with incoming event. Trying To merge changes...")
	} else if eventOrder == clock.NotComparable {
		logrus.Info("2 Events are not comparable")
	}

	fmt.Printf("After Comparing Clocks")

	for p, c := range graph.Paths {
		file, err := utils.CreateNestedFile(p)
		if err != nil {
			return err
		}
		_, err = file.Write([]byte(c))
		if err != nil {
			return err
		}
	}
	return nil
}

func init() {
	gob.Register(crdt.LastWriterWinsGraph{})
	gob.Register(crdt.Vertex{})
	gob.Register(crdt.Edge{})
	gob.Register(models.Blob{})
	gob.Register(models.Tree{})
	gob.Register(models.CommitModel{})
}
