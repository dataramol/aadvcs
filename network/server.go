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
	"io"
	"net"
	"os"
	"path/filepath"
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
		err := peer.Conn.Close()
		if err != nil {
			return err
		}
		delete(s.Peers, peer.Conn.RemoteAddr().String())
		return fmt.Errorf("%s:handshake with incoming node failed: %s", s.ListenAddress, err)
	}

	go peer.ReadLoop(s.MsgCh)

	if !peer.Outbound {
		if err := s.SendHandshake(peer); err != nil {
			err := peer.Conn.Close()
			if err != nil {
				return err
			}
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
	_, err = fp.Write(data)
	if err != nil {
		return nil, err
	}
	err = fp.Close()
	if err != nil {
		return nil, err
	}

	return hs, nil
}

func (s *Server) handleMessage(msg *Message) error {
	fmt.Printf("%+v\n", msg.Payload)
	graph := msg.Payload.(crdt.LastWriterWinsGraph)

	eventOrder := s.LastWriterWinsGraph.Clock.Compare(graph.Clock)
	if eventOrder == clock.HappensAfter {
		logrus.Info("Incoming event is Latest")
		err := handleHappensAfter(s.ListenAddress, &graph, s.LastWriterWinsGraph)
		if err != nil {
			return err
		}

	} else if eventOrder == clock.HappensBefore {
		handleHappensBefore(&graph, s.LastWriterWinsGraph)
		logrus.Info("Incoming event is stale and would be discarded")
	} else if eventOrder == clock.CONCURRENT {
		handleMerge(&graph, s.LastWriterWinsGraph)
		logrus.Info("Conflict is there with incoming event. Trying To merge changes...")
	} else if eventOrder == clock.NotComparable {
		logrus.Info("2 Events are not comparable")
	}

	fmt.Printf("After Comparing Clocks")

	return nil
}

func init() {
	gob.Register(crdt.LastWriterWinsGraph{})
	gob.Register(crdt.Vertex{})
	gob.Register(crdt.Edge{})
	gob.Register(models.Blob{})
	gob.Register(models.Tree{})
	gob.Register(models.CommitModel{})
	gob.Register(map[string]interface{}{})
}

func handleHappensAfter(serverAddress string, incomingState *crdt.LastWriterWinsGraph, currentState *crdt.LastWriterWinsGraph) error {
	// If incoming event happens after event at the server, then we accept those changes.
	currentState = incomingState
	currentState.NodeId = serverAddress
	// Create directory structure as it was on other node
	for p, c := range currentState.Paths {
		file, err := utils.CreateNestedFile(p)
		if err != nil {
			return err
		}
		_, err = file.Write([]byte(c))
		if err != nil {
			return err
		}
	}

	// now create same directories in commit folder
	noOfDirectory, err := utils.GetNumberOfChildrenDir(utils.AadvcsCommitDirPath)
	if err != nil {
		return err
	}
	commitVtx := incomingState.LatestCommit
	newCommitDirName := filepath.Join(utils.AadvcsCommitDirPath, fmt.Sprintf("v%v", noOfDirectory+1))
	commitMetadataFP, err := utils.CreateNestedFile(filepath.Join(newCommitDirName, utils.AadvcsCommitMetadataFile))
	if err != nil {
		return err
	}
	defer func(commitMetadataFP *os.File) {
		err := commitMetadataFP.Close()
		if err != nil {

		}
	}(commitMetadataFP)

	_, _ = commitMetadataFP.WriteString(fmt.Sprintf("%v%v%v", commitVtx.CommitMsg, utils.Separator, currentState.TimeStamp.Format(utils.AadvcsTimeFormat)))

	for p := range currentState.Paths {
		destCommitFilePath := filepath.Join(newCommitDirName, p)

		destFilePtr, _ := utils.CreateNestedFile(destCommitFilePath)
		originalFilePtr, _ := os.Open(p)
		_, _ = io.Copy(destFilePtr, originalFilePtr)

		err = destFilePtr.Close()
		if err != nil {
			return err
		}
		err = originalFilePtr.Close()
		if err != nil {
			return err
		}
	}

	// create graph.json
	fp, err := utils.CreateNestedFile(filepath.Join(newCommitDirName, "graph.json"))
	jsonData, err := json.MarshalIndent(currentState, "", "")
	_, _ = fp.Write(jsonData)
	err = fp.Close()
	if err != nil {
		return err
	}
	return nil
}

func handleHappensBefore(incomingState *crdt.LastWriterWinsGraph, currentState *crdt.LastWriterWinsGraph) {
	// Here we have to discard the incoming event

}

func handleMerge(incomingState *crdt.LastWriterWinsGraph, currentState *crdt.LastWriterWinsGraph) {

}
