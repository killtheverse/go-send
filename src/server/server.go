package main

import (
	"fmt"
	"net"
	"strings"
	"github.com/killtheverse/go-send/src/util"
)	

type m map[string]string
var peerAddress = m{}

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
	bytesRead, peerAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		panic(err)
	}
	fmt.Println("Buffer:", string(buffer[0:bytesRead]))
	message := strings.Split(string(buffer[0:bytesRead]), ",")[0]
	fileName := strings.Split(string(buffer[0:bytesRead]), ",")[1]

	fmt.Println("Message:", message)
	fmt.Println("Filename:", fileName)
	fmt.Println("peerAddr:", peerAddr.String())

	if message == "REGISTER" {
		peerAddress[fileName] = peerAddr.String()
		fmt.Println("Sending", "SUCCESS", "to local:", peerAddr.String())
		conn.WriteTo([]byte("SUCCESS"), peerAddr)
	} else if message == "CHECK" {
		senderAddr, ok := peerAddress[fileName]
		var sendString string
		if ok == true {
			sendString = "SUCCESS," + senderAddr  	
		} else {
			sendString = "NOTFOUND"
		}
		
		fmt.Println("Sending", sendString, "to", peerAddr)
		conn.WriteTo([]byte(sendString), peerAddr)

		if ok == true {
			publicAddr2,_ := net.ResolveUDPAddr("udp", senderAddr) 
			sendString = "REQUEST," + peerAddr.String() + "," + fileName
			conn.WriteTo([]byte(sendString), publicAddr2)
		}
	}
}