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

	peers = make(map[string]string)
	mutex sync.Mutex
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

		fmt.Printf("\n[NETWORK] Connected peer: %s (%s)\n", msg.NodeID, msg.Addr)
		printPeers()
		fmt.Print("> ")
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
	mutex.Lock()
	defer mutex.Unlock()

	if len(peers) == 0 {
		fmt.Println("No connected peers.")
		return
	}

	fmt.Println("Connected peers:")
	for id, a := range peers {
		fmt.Printf("- %s @ %s\n", id, a)
	}
}

func startCLI() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Interactive CLI started. Type `help`.")
	fmt.Print("> ")

	for {
		input, _ := reader.ReadString('\n')
		cmd := strings.TrimSpace(input)

		switch cmd {

		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  peers  - list connected peers")
			fmt.Println("  id     - show node id")
			fmt.Println("  help   - show this help")
			fmt.Println("  exit   - stop node")

		case "peers":
			printPeers()

		case "id":
			fmt.Println("Node ID:", nodeID)

		case "exit":
			fmt.Println("Shutting down node.")
			os.Exit(0)

		case "":
			// ignore empty input

		default:
			fmt.Println("Unknown command:", cmd)
		}

		fmt.Print("> ")
	}
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

	startCLI()
}
