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
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
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
	delete(s.Peers, "")
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
	conn, err := net.DialTimeout("tcp", addr, 1*time.Second)
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
			//delete(s.Peers, peer.Conn.RemoteAddr().String())
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
	for la, pr := range s.Peers {
		logrus.WithFields(logrus.Fields{
			"ListenAddress": la,
			"Peer":          pr,
		}).Info("Peers after handshake")
	}
	_, err := s.handshake(peer)
	for la, pr := range s.Peers {
		logrus.WithFields(logrus.Fields{
			"ListenAddress": la,
			"Peer":          pr,
		}).Info("Peers after handshake")
	}
	if err != nil {
		err := peer.Conn.Close()
		if err != nil {
			return err
		}
		//delete(s.Peers, peer.Conn.RemoteAddr().String())
		return fmt.Errorf("%s:handshake with incoming node failed: %s", s.ListenAddress, err)
	}

	go peer.ReadLoop(s.MsgCh)

	if !peer.Outbound {
		if err := s.SendHandshake(peer); err != nil {
			err := peer.Conn.Close()
			if err != nil {
				return err
			}
			//delete(s.Peers, peer.Conn.RemoteAddr().String())

			return fmt.Errorf("failed to send handshake with peer : %s", err)
		}
	}

	s.RegisterPeer(peer)

	logrus.WithFields(logrus.Fields{
		"peer":       peer.Conn.RemoteAddr(),
		"ListenAddr": peer.ListenAddr,
		"we":         s.ListenAddress,
	}).Info("handshake successfully: new node connected")

	for la, pr := range s.Peers {
		logrus.WithFields(logrus.Fields{
			"ListenAddress": la,
			"Peer":          pr,
		}).Info("Peers after handshake")
	}

	return nil
}

func (s *Server) Broadcast(broadcastMsg BroadcastTo, Merge bool) error {
	/*TO delete*/
	for b, p := range s.Peers {
		fmt.Printf("Listen Address -> %v\n", b)
		fmt.Printf("Peer -> %v\n", *p)
	}

	/*TO delete*/
	logrus.WithFields(logrus.Fields{
		"Merge":     Merge,
		"To":        broadcastMsg.To,
		"Peer List": s.Peers,
	}).Info("Broadcasting Message")
	msg := NewMessage(s.ListenAddress, broadcastMsg.Payload, Merge)

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
	p.ListenAddr = hs.ListenAddr

	_, ok := s.LastWriterWinsGraph.Clock.Clock[hs.ListenAddr]
	if !ok {
		s.LastWriterWinsGraph.Clock.Clock[hs.ListenAddr] = hs.CurrentGraphState.Clock.Clock[hs.ListenAddr]
	}

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

	//Reading current state of graph
	noOfCommits, err := utils.GetNumberOfChildrenDir(utils.AadvcsCommitDirPath)
	if err != nil {
		return err
	}
	if noOfCommits > 0 {
		currentDir := filepath.Join(utils.AadvcsCommitDirPath, fmt.Sprintf("v%v", noOfCommits))
		pth := filepath.Join(currentDir, "graph.json")
		file, err := os.ReadFile(pth)
		if err != nil {
			return err
		}
		latestCurrentVersion := crdt.NewLastWriterWinsGraph(s.ListenAddress)
		err = json.Unmarshal(file, latestCurrentVersion)
		if err != nil {
			return err
		}
		s.LastWriterWinsGraph = latestCurrentVersion
	}

	fmt.Printf("%+v\n", msg.Payload)
	graph := msg.Payload.(crdt.LastWriterWinsGraph)

	eventOrder := s.LastWriterWinsGraph.Clock.Compare(graph.Clock)
	if eventOrder == clock.HappensAfter {
		logrus.Info("Incoming event is Latest")
		s.LastWriterWinsGraph.Clock.Merge(graph.Clock)
		err := handleHappensAfter(&graph, s.LastWriterWinsGraph, s)
		if err != nil {
			return err
		}

	} else if eventOrder == clock.HappensBefore {
		handleHappensBefore()
		logrus.Info("Incoming event is stale and would be discarded")
	} else if eventOrder == clock.CONCURRENT {
		logrus.Info("Conflict is there with incoming event. Trying To merge changes...")
		s.LastWriterWinsGraph.Clock.Merge(graph.Clock)
		err := HandleMerge(&graph, s.LastWriterWinsGraph, s)
		if err != nil {
			return err
		}
	} else if eventOrder == clock.NotComparable {
		logrus.Info("2 Events are not comparable")
	} else if eventOrder == clock.IDENTICAL {
		logrus.Info("Nodes are eventually consistent")
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

func handleHappensAfter(incomingState *crdt.LastWriterWinsGraph, currentState *crdt.LastWriterWinsGraph, s *Server) error {
	// If incoming event happens after event at the server, then we accept those changes.

	//Deep copying edges, vertices, paths, latestCommit and timestamp
	currentState = crdt.DeepCopy(currentState, incomingState)

	// Create directory structure as it was on other node
	logrus.Info("Creating Directory Structure....")
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
	logrus.Info("Creating Commit Structure....")
	noOfDirectory, err := utils.GetNumberOfChildrenDir(utils.AadvcsCommitDirPath)
	if err != nil {
		return err
	}
	commitVtx := currentState.LatestCommit
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
	logrus.Info("Creating Graph file....")

	fp, err := utils.CreateNestedFile(filepath.Join(newCommitDirName, "graph.json"))
	jsonData, err := json.MarshalIndent(currentState, "", "")
	_, _ = fp.Write(jsonData)
	err = fp.Close()
	if err != nil {
		return err
	}
	s.LastWriterWinsGraph = currentState

	return nil
}

func handleHappensBefore() {
	// Here we have to discard the incoming event
	logrus.Info("Incoming event was the older event. Changes will not be applied.")
}

func HandleMerge(incomingState *crdt.LastWriterWinsGraph, currentState *crdt.LastWriterWinsGraph, s *Server) error {
	for _, edge := range incomingState.Edges {
		to := edge.To
		from := edge.From
		var currFrom *crdt.Vertex
		var currTo *crdt.Vertex
		switch from.ModType {
		case crdt.Commit:
		case crdt.Tree:
			var treeModel models.Tree
			err := mapstructure.Decode(from.Value, &treeModel)
			if err != nil {
				return err
			}
			currentVtx := currentState.GetVertexByFilePath(treeModel.FileName, crdt.Tree)
			if currentVtx == nil {
				currentState.AddVertex(treeModel, crdt.Tree)
			}

			currFrom = currentState.GetVertexByFilePath(treeModel.FileName, crdt.Tree)
		}
		switch to.ModType {
		case crdt.Blob:
			var blobModel models.Blob
			err := mapstructure.Decode(to.Value, &blobModel)
			if err != nil {
				return err
			}
			currentVtx := currentState.GetVertexByFilePath(blobModel.FileName, crdt.Blob)
			if currentVtx == nil {
				currentState.AddVertex(blobModel, crdt.Blob)
			} else if currentVtx.TimeStamp.Before(to.TimeStamp) {
				currentVtx.Value = blobModel
				currentVtx.TimeStamp = to.TimeStamp
			}
			currTo = currentState.GetVertexByFilePath(blobModel.FileName, crdt.Blob)

		case crdt.Tree:
			var treeModel models.Tree
			err := mapstructure.Decode(to.Value, &treeModel)
			if err != nil {
				return err
			}
			currentVtx := currentState.GetVertexByFilePath(treeModel.FileName, crdt.Tree)
			if currentVtx == nil {
				currentState.AddVertex(treeModel, crdt.Tree)
			}
			currTo = currentState.GetVertexByFilePath(treeModel.FileName, crdt.Tree)
		case crdt.Commit:
		}
		if currFrom != nil && currTo != nil && !currentState.EdgeExists(currFrom, currTo) {
			currentState.AddEdge(currTo, currFrom)
		}
	}
	rootVtx := currentState.GetRootVertex()
	currCommitModel := models.CommitModel{
		CommitMsg:     "Merge commit",
		ParentCommit:  nil,
		CommitVersion: currentState.LatestCommit.CommitVersion + 1,
	}

	currentState.LatestCommit = &currCommitModel

	currentState.AddVertex(currCommitModel, crdt.Commit)
	currentState.AddEdge(rootVtx, currentState.GetVertexByValue(currCommitModel, crdt.Commit))

	logrus.Info("Creating Directory Structure....")
	paths := make(map[string]string)
	for p := range incomingState.Paths {
		fmt.Printf("FilePath Base --> %v", filepath.Base(p))
		currentState.PrintGraph()
		vtx := currentState.GetVertexByFilePath(filepath.Base(p), crdt.Blob)
		var blob models.Blob
		mapstructure.Decode(vtx.Value, &blob)
		paths[p] = blob.Content
	}
	for p := range currentState.Paths {
		vtx := currentState.GetVertexByFilePath(filepath.Base(p), crdt.Blob)
		if _, exists := paths[p]; !exists {
			var blob models.Blob
			mapstructure.Decode(vtx.Value, &blob)
			paths[p] = blob.Content
		}
	}

	currentState.Paths = paths

	for p, c := range paths {
		file, err := utils.CreateNestedFile(p)
		if err != nil {
			return err
		}
		_, err = file.Write([]byte(c))
		if err != nil {
			return err
		}
	}

	logrus.Info("Creating Commit Structure....")
	noOfDirectory, err := utils.GetNumberOfChildrenDir(utils.AadvcsCommitDirPath)
	if err != nil {
		return err
	}
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

	_, _ = commitMetadataFP.WriteString(fmt.Sprintf("%v%v%v", currCommitModel.CommitMsg, utils.Separator, currentState.TimeStamp.Format(utils.AadvcsTimeFormat)))

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
	logrus.Info("Creating Graph file....")

	fp, err := utils.CreateNestedFile(filepath.Join(newCommitDirName, "graph.json"))
	jsonData, err := json.MarshalIndent(currentState, "", "")
	_, _ = fp.Write(jsonData)
	err = fp.Close()
	if err != nil {
		return err
	}

	s.LastWriterWinsGraph = currentState

	return nil
}
