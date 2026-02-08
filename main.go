package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

type Message struct {
	Type   string `json:"type"`
	NodeID string `json:"node_id"`
	Addr   string `json:"addr"`
}

var (
	nodeID string
	addr   string
	peers  = make(map[string]string)
	mutex  sync.Mutex
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	data, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	var msg Message
	json.Unmarshal([]byte(data), &msg)

	if msg.Type == "HELLO" {
		mutex.Lock()
		peers[msg.NodeID] = msg.Addr
		mutex.Unlock()

		fmt.Printf("Connected peer: %s (%s)\n", msg.NodeID, msg.Addr)
		printPeers()
	}
}

func startServer() {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Listening on", addr)

	for {
		conn, _ := ln.Accept()
		go handleConnection(conn)
	}
}

func connectToPeer(peerAddr string) {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return
	}
	defer conn.Close()

	msg := Message{
		Type:   "HELLO",
		NodeID: nodeID,
		Addr:   addr,
	}

	data, _ := json.Marshal(msg)
	conn.Write(append(data, '\n'))
}

func printPeers() {
	fmt.Println("Current peers:")
	for id, a := range peers {
		fmt.Printf("- %s @ %s\n", id, a)
	}
	fmt.Println("----")
}

func main() {
	peerList := flag.String("peers", "", "comma-separated peers")
	flag.StringVar(&nodeID, "id", "", "node id")
	flag.StringVar(&addr, "addr", "", "listening address")
	flag.Parse()

	if nodeID == "" || addr == "" {
		fmt.Println("Missing --id or --addr")
		os.Exit(1)
	}

	go startServer()

	if *peerList != "" {
		timePeers := strings.Split(*peerList, ",")
		for _, p := range timePeers {
			connectToPeer(p)
		}
	}

	select {} // keep running
}
