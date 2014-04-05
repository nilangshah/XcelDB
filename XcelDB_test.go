package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"testing"
	//"time"
	"io/ioutil"
	"os/exec"
	"sync"
	"time"
)

type counter struct {
	mutex sync.Mutex
	count int
}

var wg sync.WaitGroup
var sem *counter
var getPassed int
var setPassed int
var leaderURl string
var nos int

func TestXcelDB(t *testing.T) {
	//change this url if u change url in config file
	leaderURl = "http://127.0.0.1:14961"
	// change the nos if u want to chnage the no of servers in cluster, this should be same as no of server in config file
	nos = 3
	cmd := make([]*exec.Cmd, nos)
	path := GetPath() + "/bin/XcelDB"
	// start all server
	for i := 1; i < nos+1; i++ {
		cmd[i-1] = exec.Command(path, "-id", strconv.Itoa(i))
		cmd[i-1].Start()
	}
	select {
	case <-time.After(4 * time.Second):
	}
	//set values 
	sem = new(counter)
	for j := 0; j < 450; j++ {
		wg.Add(1)
		go setValues(j)

	}

	wg.Wait()
	// get values
	for j := 0; j < 450; j++ {
		wg.Add(1)
		go getValues(j)
	}
	fmt.Println("command done")
	wg.Wait()
	fmt.Println(getPassed, setPassed)
	if getPassed == setPassed && setPassed == 450 {
		fmt.Println("Thank you Test successful")
		kill_all_server(cmd)
	} else {
		kill_all_server(cmd)
		panic("test failed")
	}
}

func kill_all_server(cmd []*exec.Cmd) {
	for i := 1; i < nos; i++ {
		cmd[i-1].Process.Kill()
		cmd[i-1].Wait()
	}

}
func getValues(i int) {
	defer wg.Done()
	i++
	xcel := &Xcel{"GET", []byte(strconv.Itoa(i)), nil, 1, ""}
	buf, _ := xml.Marshal(xcel)

	res := sendPostRequest(leaderURl, "text/xml", buf)
	if res == 0 {
		getPassed++
	}
	//wg.Done()

}

func sendPostRequest(url string, contentType string, buf []byte) int {
	var xcel Xcel

	body1 := bytes.NewBuffer(buf)
	resp, err := http.Post(url, contentType, body1)
	if err != nil {
		// handle error
		fmt.Println(err)
		return 1

	} else {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		xml.Unmarshal(body, &xcel)
		//fmt.Println(xcel.ServerResponse)
		switch xcel.ServerResponse {
		case updateFailed:
			return updateFailed
		case updateSuccess:
			return updateSuccess
		case leaderNotMe:
			leaderURl = xcel.Leader
			return sendPostRequest(leaderURl, contentType, buf)
		case leaderUnknown:
			return updateFailed

		}

	}
	return 1
}

func setValues(i int) {

	defer wg.Done()
	i++
	xcel := &Xcel{"SET", []byte(strconv.Itoa(i)), []byte("a " + strconv.Itoa(i)), 1, ""}
	buf, _ := xml.Marshal(xcel)
	//body := bytes.NewBuffer(buf)

	res := sendPostRequest(leaderURl, "text/xml", buf)
	if res == 0 {
		setPassed++
	}

}
