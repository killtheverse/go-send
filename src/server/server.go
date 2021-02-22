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
	listenAddrString, err := util.ExternalIP()
	if err != nil {
		panic(err)
	}
	listenAddrString = listenAddrString + ":8000"
	listenAddr, _ := net.ResolveUDPAddr("udp", listenAddrString)
	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Listening on", listenAddrString)

	for {
		handleConnection(conn)
	}
}

func handleConnection(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		panic(err)
	}
	fmt.Println("Buffer:", string(buffer[0:bytesRead]))
	message := strings.Split(string(buffer[0:bytesRead]), ",")[0]
	fileName := strings.Split(string(buffer[0:bytesRead]), ",")[1]
	peerAddrString := strings.Split(string(buffer[0:bytesRead]), ",")[2] 
	peerAddr, _ := net.ResolveUDPAddr("udp", peerAddrString)
	fmt.Println("Message:", message)
	fmt.Println("Filename:", fileName)
	fmt.Println("peerAddrString:", peerAddrString)
	if message == "REGISTER" {
		fileAddress[fileName] = peerAddrString
		fmt.Println("Sending", "SUCCESS", "to", peerAddrString)
		conn.WriteTo([]byte("SUCCESS"), peerAddr)
	} else if message == "CHECK" {
		addr, ok := fileAddress[fileName]
		var sendString string
		if ok == true {
			sendString = "SUCCESS," + addr 	
		} else {
			sendString = "NOTFOUND"
		}
		
		fmt.Println("Sending", sendString, "to", peerAddrString)
		conn.WriteTo([]byte(sendString), peerAddr)

		if ok == true {
			peerAddr2,_ := net.ResolveUDPAddr("udp", addr) 
			sendString = "REQUEST," + peerAddrString + "," + fileName
			conn.WriteTo([]byte(sendString), peerAddr2)
		}
	}
}