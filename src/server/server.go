package main

import (
	"fmt"
	"net"
	"strings"
	"github.com/killtheverse/go-send/src/util"
)	

type m map[string]string
var fileAddress = m{}

func main() {
	listenAddr, err := util.ExternalIP()
	if err != nil {
		panic(err)
	}
	// fmt.Println("address is:", listenAddr)
	listenAddr = listenAddr + ":8000"
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Listening on", listenAddr)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	fmt.Println("Serving", remoteAddr)

	buffer := make([]byte, 1024)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		panic(err)
	}
	fmt.Println("Buffer:", string(buffer[0:bytesRead]))
	message := strings.Split(string(buffer[0:bytesRead]), ",")[0]
	fileName := strings.Split(string(buffer[0:bytesRead]), ",")[1]
	peerAddr := strings.Split(string(buffer[0:bytesRead]), ",")[2] 
	fmt.Println("Message:", message)
	fmt.Println("Filename:", fileName)
	fmt.Println("peerAddr:", peerAddr)
	if message == "REGISTER" {
		fileAddress[fileName] = peerAddr
		conn, err := net.Dial("tcp", peerAddr)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Sending", "SUCCESS", "to", conn.RemoteAddr().String())
		conn.Write([]byte("SUCCESS"))
	} else if message == "CHECK" {
		conn, err := net.Dial("tcp", peerAddr)
		if err != nil {
			fmt.Println(err)
		}
		addr, ok := fileAddress[fileName]
		var sendString string
		if ok == true {
			sendString = "SUCCESS," + addr 	
		} else {
			sendString = "NOTFOUND"
		}
		
		fmt.Println("Sending", sendString, "to", conn.RemoteAddr().String())
		conn.Write([]byte(sendString))

		if ok == true {
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				fmt.Println(err)
			}
			sendString = "REQUEST," + peerAddr + "," + fileName
			conn.Write([]byte(sendString))
		}
	}
	conn.Close()
}