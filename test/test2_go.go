package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
)

func main() {
	port:=14130
	//leader_Id:=0

	//conn:=make([]net.Conn,5) 
	for i:=1;i<6;i++{
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port+i))
	
		if err != nil {
			fmt.Println("Connection error: " + err.Error())
		}else{
			//leader_Id=i	
	
//	fmt.Println(leader_Id)



	reader := bufio.NewReader(conn)
	sentence := "set alpha gamma\n"
	fmt.Println("SEND SERVER:" + sentence)
	conn.Write([]uint8(sentence))
	content, err := reader.ReadString('\n')
	if err == io.EOF {
	} else if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("FROM SERVER:" + content)
	sentence = "set alpha1 gamma1\n"

	fmt.Println("SEND SERVER:" + sentence)
	conn.Write([]uint8(sentence))
	content, err = reader.ReadString('\n')
	if err == io.EOF {
	} else if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("FROM SERVER:" + content)
	sentence = "delete alpha\n"

	fmt.Println("SEND SERVER:" + sentence)
	conn.Write([]uint8(sentence))
	content4, err := reader.ReadString('\n')
	if err == io.EOF {
	} else if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("FROM SERVER:" + content4)

	sentence = "get alpha\n"

	fmt.Println("SEND SERVER:" + sentence)
	conn.Write([]uint8(sentence))
	content1, err := reader.ReadString('\n')
	if err == io.EOF {
	} else if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("FROM SERVER:" + content1)

	sentence = "get alpha1\n"

	fmt.Println("SEND SERVER:" + sentence)
	conn.Write([]uint8(sentence))
	content2, err := reader.ReadString('\n')
	if err == io.EOF {
	} else if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("FROM SERVER:" + content2)
	sentence = "quit" + "\n"
	fmt.Println("SEND SERVER:" + sentence)
	conn.Write([]uint8(sentence))
	content5, err := reader.ReadString('\n')
	if err == io.EOF {
	} else if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("FROM SERVER:" + content5)

	fmt.Println("Thank you Test successfull")
		defer conn.Close()	
		}
	}

}
