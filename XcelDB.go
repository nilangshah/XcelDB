package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/nilangshah/Raft"
	"github.com/nilangshah/Raft/cluster"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func GetPath() string {
	data := os.Environ()
	for _, item := range data {
		key, val := getkeyval(item)
		if key == "GOPATH" {
			return val
		}
	}
	return ""
}

func getkeyval(item string) (key, val string) {
	splits := strings.Split(item, "=")
	key = splits[0]
	newval := strings.Join(splits[1:], "=")
	vals := strings.Split(newval, ":")
	val = vals[0]
	return
}

type GoDB struct {
	sync.RWMutex
	kvMap map[string][]byte
}

var goDB *GoDB
var replicator Raft.Replicator

func main() {

	myid := flag.Int("id", 1, "a int")
	flag.Parse()
	logfile := os.Getenv("GOPATH") + "/src/github.com/nilangshah/Raft/Raftlog/log" + strconv.Itoa(*myid)
	path := GetPath() + "/src/github.com/nilangshah/Raft/cluster/config.json"
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("error opening file: %v", err))
	} else {
		//defer f.Close()

	}
	port := 14130
	server1 := cluster.New(*myid, path)
	replicator := Raft.New(server1, f, path)
	replicator.Start()
	fmt.Println(*myid)
	fmt.Println("hiii")

	goDB = New()
	done := false
	var listener net.Listener
	select {
	case <-time.After(5 * time.Second):
		if replicator.IsLeader() {

			fmt.Println("start listen")
			listener, err = net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port+*myid))
			if err != nil {
				panic("Error listening on" + strconv.Itoa(port+*myid) + err.Error())

			} else {
				done = true
			}

		}
	}
	out := make(chan int)
	//listener:=new(net.Listener)
	if done {
		for {

			netconn, err := listener.Accept()
			if err != nil {
				panic("Accept error: " + err.Error())
			}

			go handleConn(netconn, &replicator)
		}
	} else {
		<-out
	}

}

func New() *GoDB {
	d := &GoDB{
		kvMap: map[string][]byte{},
	}

	return d
}

/*
* Networking
 */
func handleConn(conn net.Conn, replicator *Raft.Replicator) {
	defer conn.Close()
	fmt.Println("hiii")
	reader := bufio.NewReader(conn)
	for {

		// Fetch

		content, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
			return
		}

		content = content[:len(content)-1] // Chop \n

		// Handle

		subContent := strings.Split(content, " ")
		cmd := subContent[0]
		switch cmd {

		case "get":
			goDB.RLock()
			key := subContent[1]
			val, ok := goDB.kvMap[key]
			if ok {
				conn.Write([]uint8(string(val) + "\r\n"))
			} else {
				conn.Write([]uint8("NIL\r\n"))
			}
			goDB.RUnlock()
			continue
		case "set":

			response := make(chan bool)
			command := Raft.CommandTuple{Command: []byte(content), CommandResponse: response}
			(*replicator).Outbox() <- command
			select {
			case t := <-response:
				if t {
					goDB.Lock()
					key := subContent[1]
					val := subContent[2]
					goDB.kvMap[key] = []byte(val)
					conn.Write([]uint8("STORED\r\n"))
					goDB.Unlock()
				} else {
					conn.Write([]uint8("FAILED.. PLEASE RETRY\r\n"))
				}

			}
			continue
		case "delete":

			response := make(chan bool)
			command := Raft.CommandTuple{Command: []byte(content), CommandResponse: response}
			(*replicator).Outbox() <- command
			select {
			case t := <-response:
				if t {
					goDB.Lock()
					key := subContent[1]
					delete(goDB.kvMap, key)
					conn.Write([]uint8("DELETED\r\n"))

					goDB.Unlock()
				} else {
					conn.Write([]uint8("FAILED.. PLEASE RETRY\r\n"))
				}

			}
			continue
		case "quit":
			break

		}
		break
	}
}
