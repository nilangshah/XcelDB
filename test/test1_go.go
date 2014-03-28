package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
)

type counter struct {
	mutex sync.Mutex
	count int
}

var wg sync.WaitGroup
var sem *counter

func main() {
	sem = new(counter)
	fmt.Println(sem.count)
	for j := 0; j < 500; j++ {
		wg.Add(1)
		go setValues(j)
	}

	wg.Wait()
	for j := 0; j < 500; j++ {
		wg.Add(1)
		go getValues(j)
	}
	wg.Wait()

	fmt.Println("Thank you Test successfull")
}

func getValues(i int) {
	defer wg.Done()
	i++
	conn, err := net.Dial("tcp", "127.0.0.1:13131")
	defer conn.Close()
	if err != nil {
		panic("Connection Error:" + err.Error())
	}
	reader := bufio.NewReader(conn)
	sentence := "get " + strconv.Itoa(i) + "\n"

	fmt.Println("SEND SERVER:" + sentence)
	conn.Write([]uint8(sentence))
	content, err := reader.ReadString('\n')
	if err == io.EOF {
	} else if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("FROM SERVER:" + content)
	sentence = "quit" + "\n"
	fmt.Println("SEND SERVER:" + sentence)
	conn.Write([]uint8(sentence))
	fmt.Println("FROM SERVER:" + content)

}

func setValues(i int) {
	defer wg.Done()

	i++
	conn, err := net.Dial("tcp", "127.0.0.1:11211")
	defer conn.Close()
	if err != nil {
		panic("Connection error: " + err.Error())
	}
	reader := bufio.NewReader(conn)
	sentence := "set " + strconv.Itoa(i) + " a" + strconv.Itoa(i) + "\n"
	fmt.Println("SEND SERVER:" + sentence)
	conn.Write([]uint8(sentence))
	content, err := reader.ReadString('\n')
	if err == io.EOF {
	} else if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("FROM SERVER:" + content)
	sentence = "quit" + "\n"
	fmt.Println("SEND SERVER:" + sentence)
	conn.Write([]uint8(sentence))
	fmt.Println("FROM SERVER:" + content)

}
