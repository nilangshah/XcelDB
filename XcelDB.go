package main

import (
	"bytes"
	"encoding/gob"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/nilangshah/Raft"
	"github.com/nilangshah/Raft/cluster"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	leaderUnknown = -2
	leaderNotMe   = -1
	updateFailed  = 1
	updateSuccess = 0
)

type Jsonobject struct {
	Object ObjectType
}

type ObjectType struct {
	Servers       []ServerInfo
	No_of_servers uint64
}

type ServerInfo struct {
	Id   uint64
	Host string
}

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

type StateMachine struct {
	sync.RWMutex
	kvMap map[string][]byte
}

type XcelDB struct {
	xcelId         uint64            //id of server
	xcelSM         *StateMachine     // state machine
	xcelReplicator Raft.Replicator   // raft replicator
	xcelPeers      []uint64          // id of all peers
	xcelPeermap    map[uint64]string // address of all peers
}

// apply command to state machine
func ApplyCommandTOSM(xcel *Xcel) {
	switch xcel.Command {

	case "GET":
		xcelDB.xcelSM.RLock()
		key := xcel.Key
		val, ok := xcelDB.xcelSM.kvMap[string(key)]
		if ok {
			xcel.Value = val
			xcel.ServerResponse = updateSuccess

		} else {
			xcel.Value = []byte("NIL")
			xcel.ServerResponse = updateFailed

		}
		xcelDB.xcelSM.RUnlock()

	case "SET":
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(xcel)
		if err != nil {
			panic("gob error: " + err.Error())
		}
		response := make(chan bool)
		command := Raft.CommandTuple{Command: []byte(buf.String()), CommandResponse: response}
		(xcelDB.xcelReplicator).Outbox() <- command
		select {
		case t := <-response:
			if t {
				xcelDB.xcelSM.Lock()
				key := xcel.Key
				val := xcel.Value
				xcelDB.xcelSM.kvMap[string(key)] = []byte(val)
				xcel.ServerResponse = updateSuccess

				xcelDB.xcelSM.Unlock()
			} else {
				xcel.ServerResponse = updateFailed

			}

		}

	case "DELETE":
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(xcel)
		if err != nil {
			panic("gob error: " + err.Error())
		}
		response := make(chan bool)
		command := Raft.CommandTuple{Command: []byte(buf.String()), CommandResponse: response}
		(xcelDB.xcelReplicator).Outbox() <- command
		select {
		case t := <-response:
			if t {
				xcelDB.xcelSM.Lock()
				key := xcel.Key
				delete(xcelDB.xcelSM.kvMap, string(key))
				xcel.ServerResponse = updateSuccess

				xcelDB.xcelSM.Unlock()
			} else {
				xcel.ServerResponse = updateFailed
			}

		}

	}

}

//http handler to handle http request
func kvHandler(w http.ResponseWriter, r *http.Request) {
	var xcel Xcel
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	xml.Unmarshal(body, &xcel)
	fmt.Println(xcel)
	if xcelDB.xcelReplicator.IsLeader() {
		fmt.Println(xcelDB.xcelId, ":  ", xcel)
		ApplyCommandTOSM(&xcel)
		//xcel.ServerResponse = true
		responseXML, _ := xml.Marshal(xcel)
		fmt.Fprintf(w, string(responseXML))
	} else {

		leader := xcelDB.xcelReplicator.GetLeader()
		if leader == 0 {
			xcel.ServerResponse = leaderUnknown
			xcel.Leader = "unknown"
		} else {
			xcel.ServerResponse = leaderNotMe
			xcel.Leader = "http://" + xcelDB.xcelPeermap[leader]

		}
	}
	responseXML, _ := xml.Marshal(xcel)
	fmt.Fprintf(w, string(responseXML))

}

//client must send object of xcel as http post to communicate with XcelDBs
type Xcel struct {
	Command        string
	Key            []byte
	Value          []byte
	ServerResponse int
	Leader         string
}

//initialize new state machine
func NewStateMachine() *StateMachine {
	d := &StateMachine{
		kvMap: map[string][]byte{},
	}
	return d
}

//all past commited command will come into replicators inbox
func ListenInBox(Replicator Raft.Replicator) {
	count := 0
	var xcel Xcel
	for {
		select {
		case t := <-Replicator.Inbox():
			count++
			fmt.Println("enter receinved:", count)
			buf := bytes.NewBufferString(string(*t))
			dec := gob.NewDecoder(buf)

			err := dec.Decode(&xcel)
			if err != nil {
				panic(fmt.Sprintf("decode:", err))
			} else {
				ApplyOldCommandTOSM(&xcel)

			}

		}
	}

}

func ApplyOldCommandTOSM(xcel *Xcel) {
	switch xcel.Command {
	case "SET":
		xcelDB.xcelSM.Lock()
		key := xcel.Key
		val := xcel.Value
		xcelDB.xcelSM.kvMap[string(key)] = []byte(val)
		xcelDB.xcelSM.Unlock()
	case "DELETE":
		xcelDB.xcelSM.Lock()
		key := xcel.Key
		delete(xcelDB.xcelSM.kvMap, string(key))
		xcelDB.xcelSM.Unlock()
	}
}

var xcelDB *XcelDB

// create new xceldb
func main() {

	var Cluster cluster.Server
	var Replicator Raft.Replicator
	Id := flag.Int("id", 1, "a int")
	flag.Parse()
	configFname := GetPath() + "/src/github.com/nilangshah/XcelDB/config.xml"
	Confname := GetPath() + "/src/github.com/nilangshah/XcelDB/c_config.xml"
	LogPath := GetPath() + "/src/github.com/nilangshah/XcelDB/Raftlog" + strconv.Itoa(*Id)

	//intialize state machine
	xcelSM := NewStateMachine()

	//intialize replicator and cluster
	Cluster = cluster.New(*Id, Confname)
	Replicator = Raft.New(Cluster, LogPath)

	// listen inbox for past commited commands
	go ListenInBox(Replicator)

	//initialize db
	xcelDB = &XcelDB{
		xcelId:         uint64(*Id),
		xcelSM:         xcelSM,
		xcelReplicator: Replicator,
	}

	//start replicator
	Replicator.Start()

	//intialize xceldb

	//read config file
	var Jsontype Jsonobject
	file, e := ioutil.ReadFile(configFname)
	if e != nil {
		panic("File error: " + e.Error())
	}

	xml.Unmarshal(file, &Jsontype)
	//fmt.Println(xml.Marshal(&Jsontype))
	// store configuration
	count := 0
	xcelDB.xcelPeers = make([]uint64, len(Jsontype.Object.Servers))
	xcelDB.xcelPeermap = make(map[uint64]string, len(Jsontype.Object.Servers))
	for i := range Jsontype.Object.Servers {
		if Jsontype.Object.Servers[i].Id == uint64(*Id) {

		} else {
			xcelDB.xcelPeers[count] = Jsontype.Object.Servers[i].Id
			count++
		}
		xcelDB.xcelPeermap[Jsontype.Object.Servers[i].Id] = Jsontype.Object.Servers[i].Host

	}

	//start http server
	http.HandleFunc("/", kvHandler)
	select {
	case <-time.After(2 * time.Second):
		fmt.Println(xcelDB.xcelReplicator.IsRunning())

	}
	http.ListenAndServe(xcelDB.xcelPeermap[uint64(*Id)], nil)

}
