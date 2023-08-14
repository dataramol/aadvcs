package main

import (
	"github.com/dataramol/aadvcs/cli"
	"github.com/dataramol/aadvcs/crdt"
	"github.com/dataramol/aadvcs/models"
	"github.com/dataramol/aadvcs/network"
	"time"
)

func main() {
	cli.Execute()

	/*
		node1 := createServerAndStart(":3000")
		time.Sleep(time.Second * 1)
		err := node1.Dial(":4000")
		time.Sleep(time.Second * 2)
		buf := new(bytes.Buffer)
		if err := gob.NewEncoder(buf).Encode(node1); err != nil {
			fmt.Printf("Error --> %v", err)
		}
		fp, err := utils.CreateOrOpenFileRWMode("D:\\Study\\MSc Project\\test1\\.aadvcs\\network")
		err = utils.ClearFileContent(fp)
		if err != nil {
			fmt.Printf("Error ---> %v", err)
		}*/

	/*fp, err = utils.CreateOrOpenFileRWMode("D:\\Study\\MSc Project\\test1\\.aadvcs\\network.json")
	Server := &network.Server{}
	err = gob.NewDecoder(fp).Decode(Server)
	if err != nil {
		fmt.Printf("Error ---> %v+", err)
	}
	newServer := network.NewServer(Server.ListenAddress)
	time.Sleep(time.Second * 2)
	constructLwwGraph(newServer.LastWriterWinsGraph)

	To := make([]string, len(newServer.Peers))
	for _, peer := range newServer.Peers {
		To = append(To, peer.ListenAddr)
	}

	err = newServer.Broadcast(network.BroadcastTo{
		To:      To,
		Payload: newServer.LastWriterWinsGraph,
	})
	if err != nil {
		fmt.Printf("Error ---> %v+", err)
	}

	time.Sleep(time.Second * 2)*/

	/*node1 := createServerAndStart(":3000")
	node2 := createServerAndStart(":4000")
	time.Sleep(time.Second * 1)
	err := node1.Dial(node2.ListenAddress)
	if err != nil {
		return
	}
	time.Sleep(time.Second * 1)
	node1.LastWriterWinsGraph.Clock.Print()
	node2.LastWriterWinsGraph.Clock.Print()

	constructLwwGraph(node1.LastWriterWinsGraph)

	time.Sleep(time.Second * 1)

	To := make([]string, len(node1.Peers))
	for _, peer := range node1.Peers {
		fmt.Printf("Broadcasting message to %v\n", peer.Conn.RemoteAddr())
		To = append(To, peer.ListenAddr)
	}

	err = node1.Broadcast(network.BroadcastTo{
		To:      To,
		Payload: node1.LastWriterWinsGraph,
	})
	if err != nil {
		return
	}

	time.Sleep(time.Second * 2)*/
}

func createServerAndStart(addr string) *network.Server {
	server := network.NewServer(addr)
	lwwGraph := crdt.NewLastWriterWinsGraph(server.ListenAddress)
	server.LastWriterWinsGraph = lwwGraph
	go server.Start()
	time.Sleep(time.Millisecond * 200)

	return server
}

func constructLwwGraph(lwwGraph *crdt.LastWriterWinsGraph) {
	val1 := models.Tree{
		FileName: "dir",
	}
	val2 := models.Blob{
		Content:  "ndskjgjdf",
		FileName: "file1.txt",
	}
	val3 := models.Tree{
		FileName: "lib",
	}
	val4 := models.Blob{
		Content:  "123234356",
		FileName: "file2.txt",
	}

	lwwGraph.AddVertex(val1, crdt.Tree)
	lwwGraph.AddVertex(val2, crdt.Blob)
	lwwGraph.AddVertex(val3, crdt.Tree)
	lwwGraph.AddVertex(val4, crdt.Blob)

	lwwGraph.AddEdge(lwwGraph.GetVertexByFilePath("file1.txt", crdt.Blob), lwwGraph.GetVertexByFilePath("dir", crdt.Tree))
	lwwGraph.AddEdge(lwwGraph.GetVertexByFilePath("lib", crdt.Tree), lwwGraph.GetVertexByFilePath("dir", crdt.Tree))
	lwwGraph.AddEdge(lwwGraph.GetVertexByFilePath("file2.txt", crdt.Blob), lwwGraph.GetVertexByFilePath("lib", crdt.Tree))

	lwwGraph.IncrementClock()
}
