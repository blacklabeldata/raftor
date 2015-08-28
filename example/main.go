package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/coreos/etcd/raft"
	"github.com/coreos/etcd/raft/raftpb"
	"github.com/hashicorp/memberlist"
	"github.com/swiftkick-io/xbinary"
)

// Raft support
// - Setup Memberlist
// 	- If the cluster does not already exist, create a new raft.Node instance with no peers.
// 	- If the cluster exists, get the metadata from each node and create a new raft.Node from the peer IDs
// - If a node joins, propose a config change (?) to add the peer
// - If a node leaves, propose a config change (?) to remove the peer
// - To send messages, propose a message

// TODO: Write BoltDB storage engine for raft.

import "flag"

var name = flag.String("name", "", "node name")
var port = flag.Int("port", 7496, "bind port")
var tcpport = flag.Int("tcpport", 9000, "tcp bind port")
var addr = flag.String("addr", "127.0.0.1", "bind addr")
var tcpaddr = flag.String("tcpaddr", "127.0.0.1", "tcp bind addr")
var hosts = flag.String("hosts", "", "address of peer nodes")

func main() {
	flag.Parse()
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	/* Create the initial memberlist from a safe configuration.
	   Please reference the godoc for other default config types.
	   http://godoc.org/github.com/hashicorp/memberlist#Config
	*/
	evtDelegate := MultiEventDelegate{}
	var delegate = NewDelegate(*name, *tcpport)
	var cfg = memberlist.DefaultLocalConfig()
	cfg.Events = &evtDelegate
	cfg.Delegate = delegate
	cfg.Name = *name
	cfg.BindPort = *port
	cfg.BindAddr = *addr
	cfg.AdvertisePort = *port
	cfg.AdvertiseAddr = *addr

	list, err := memberlist.Create(cfg)
	if err != nil {
		log.Fatalln("Failed to create memberlist: " + err.Error())
		return
	}

	if len(*hosts) > 0 {
		// Join an existing cluster by specifying at least one known member.
		_, err = list.Join(strings.Split(*hosts, ","))
		if err != nil {
			log.Println("Failed to join cluster: " + err.Error())
		}
	}

	// Ask for members of the cluster
	for _, member := range list.Members() {
		fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
	}

	lookup := NewNodeLookup()
	transport := NewTcpTransporter()

	// Setup raft
	id := uint32(Jesteress([]byte(cfg.Name)))
	log.Printf("Name: %s ID: %d", cfg.Name, id)
	storage := raft.NewMemoryStorage()
	c := &raft.Config{
		ID:              uint64(id),
		ElectionTick:    10,
		HeartbeatTick:   1,
		Storage:         storage,
		MaxSizePerMsg:   4096,
		MaxInflightMsgs: 256,
	}
	log.Println("Node ID: ", c.ID)
	r := NewRaftNode(*name, c, storage, lookup, transport)
	evtDelegate.AddEventDelegate(r)

	// Listen for incoming connections.
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *tcpport))
	if err != nil {
		log.Println("Error listening:", err.Error())
		return
	}

	// Start TCP server
	server := TCPServer{l, r}
	go server.Start()
	fmt.Printf("Listening on %s:%d\n", *tcpaddr, *tcpport)

	// Close the listener when the application closes.
	defer server.Stop()

	// Start raft server
	go r.Start()

	// Handle signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	// Wait for signal
	log.Println("Cluster open for business")

	// Block until signal is received
	<-sig
	log.Println("Shutting down node")
	if err = list.Leave(time.Second); err != nil {
		log.Println("Error leaving cluster: " + err.Error())
	}

	if err = list.Shutdown(); err != nil {
		log.Println("Error shutting down node: " + err.Error())
	}
}

type TCPServer struct {
	listener net.Listener
	raft     *raftNode
}

func (t *TCPServer) Start() {
	for {
		// Listen for an incoming connection.
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}

		//logs an incoming message
		fmt.Printf("Received message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())

		// Handle connections in a new goroutine.
		go t.handleRequest(conn)
	}
}

func (t *TCPServer) Stop() error {
	return t.listener.Close()
}

// Handles incoming requests.
func (t *TCPServer) handleRequest(conn net.Conn) {
	var b []byte
	buffer := bytes.NewBuffer(b)
	bufReader := bufio.NewReader(conn)

	size := make([]byte, 4)
	ack := []byte{byte(AckMessageType)}

	for {
		typ, err := bufReader.ReadByte()
		if err != nil {
			log.Printf("Failed to read message type from TCP client")
			break
		}

		switch UserMessageType(typ) {
		case RaftMessageType:
			log.Printf("Received Raft message type")

			// Read size
			n, err := bufReader.Read(size)
			if err != nil || n != 4 {
				log.Printf("Failed to read message size from TCP client")
				break
			}

			byteCount, _ := xbinary.LittleEndian.Uint32(size, 0)
			if n, err := buffer.ReadFrom(io.LimitReader(bufReader, int64(byteCount))); err != nil || n != int64(byteCount) {
				log.Printf("Failed to read message from TCP client")
				break
			}

			m := raftpb.Message{}
			if err := m.Unmarshal(buffer.Bytes()); err != nil {
				log.Printf("Failed to decode message from TCP client")
				break
			}
			log.Printf("Applying Raft message")
			t.raft.Step(m)

			// Write the message in the connection channel.
			conn.Write(ack)

			// Reset Buffer
			buffer.Reset()

		default:
			break
		}
	}

	// Close the connection when you're done with it.
	conn.Close()
}
