package XcelDB

import (
	//"bufio"
	"encoding/xml"
	//"flag"
	"encoding/json"
	"fmt"
	"github.com/nilangshah/Raft"
	"github.com/nilangshah/Raft/cluster"
	//"io"
	"bytes"
	"encoding/gob"
	"io/ioutil"
	//"net"
	"net/http"
	"os"
	//"strconv"
	"strings"
	"sync"
	"time"
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
	xcelId         uint64
	xcelSM         *StateMachine
	xcelServer     *http.Server
	xcelReplicator Raft.Replicator
	xcelPeers      []uint64
	xcelPeermap    map[uint64]string
}

//var replicator Raft.Replicator

func ApplyCommandTOSM(xcel *Xcel) {
	switch xcel.Command {

	case "GET":
		xcelDB.xcelSM.RLock()
		key := xcel.Key
		val, ok := xcelDB.xcelSM.kvMap[string(key)]
		if ok {
			xcel.Value = val
			xcel.ServerResponse = true

		} else {
			xcel.Value = []byte("NIL")
			xcel.ServerResponse = false

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
				xcel.ServerResponse = true

				xcelDB.xcelSM.Unlock()
			} else {
				xcel.ServerResponse = false

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
				xcel.ServerResponse = true

				xcelDB.xcelSM.Unlock()
			} else {
				xcel.ServerResponse = false
			}

		}

	}

}

func kvHandler(w http.ResponseWriter, r *http.Request) {
	var xcel Xcel

	body, _ := ioutil.ReadAll(r.Body)
	xml.Unmarshal(body, &xcel)
	ApplyCommandTOSM(&xcel)
	responseXML, _ := xml.Marshal(xcel)
	fmt.Fprintf(w, string(responseXML))
}

type Xcel struct {
	Command        string
	Key            []byte
	Value          []byte
	ServerResponse bool
}

func NewStateMachine() *StateMachine {
	d := &StateMachine{
		kvMap: map[string][]byte{},
	}
	return d
}

var xcelDB *XcelDB

func NewServer(xcelId uint64, configFname string, clusterConfname string, raftLogPath string) {

	//intialize state machine
	xcelSM := NewStateMachine()

	// initialize http server
	xcelS := &http.Server{
		Addr:           ":12345",
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	http.HandleFunc("/", kvHandler)

	//start http server
	go xcelS.ListenAndServe()

	//intialize replicator
	xcelCluster := cluster.New(int(xcelId), clusterConfname)
	xcelReplicator := Raft.New(xcelCluster, raftLogPath)
	xcelReplicator.Start()

	//intialize xceldb
	xcelDB := &XcelDB{
		xcelId:         xcelId,
		xcelSM:         xcelSM,
		xcelServer:     xcelS,
		xcelReplicator: xcelReplicator,
	}

	//read config file
	var Jsontype Jsonobject
	file, e := ioutil.ReadFile(configFname)
	if e != nil {
		panic("File error: " + e.Error())
	}

	json.Unmarshal(file, &Jsontype)

	// store configuration
	count := 0
	for i := range Jsontype.Object.Servers {
		if Jsontype.Object.Servers[i].Id == xcelId {
		} else {
			xcelDB.xcelPeers[count] = Jsontype.Object.Servers[i].Id
			count++
		}
		xcelDB.xcelPeermap[Jsontype.Object.Servers[i].Id] = Jsontype.Object.Servers[i].Host

	}

}
