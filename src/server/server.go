package main

import (
	"fmt"
	"net"
	"encoding/json"

	"github.com/killtheverse/go-send/src/utils"
)	

type m map[string]string
var peerAddress = m{}

func main() {
	listenAddrString, err := utils.ExternalIP()
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

	handleConnection(conn)
}

func handleConnection(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	for {
		bytesRead, peerAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			panic(err)
		}
		
		recvData := make(map[string]string)
		err = json.Unmarshal(buffer[0:bytesRead], &recvData)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		fmt.Printf("%+v\n", recvData)
		
		instruction := recvData["INSTRUCTION"]
		fmt.Println("instruction: ", instruction)

		if instruction == "KEEPALIVE" {
			continue
		} else if instruction == "REGISTER" {
			fileName := recvData["FILENAME"]
			peerAddress[fileName] = peerAddr.String()
			fmt.Println("Sending", "SUCCESS", "to local:", peerAddr.String())
			sendData := make(map[string]string)
			sendData["INSTRUCTION"] = "SUCCESS"
			utils.SendData(peerAddr.String(), conn, sendData)
		} else if instruction == "CHECK" {
			fileName := recvData["FILENAME"]
			senderAddr, ok := peerAddress[fileName]
			sendData := make(map[string]string)
			if ok == true {
				sendData["INSTRUCTION"] = "SUCCESS"
				sendData["SENDER"] = senderAddr
				
				sendData2 := make(map[string]string)
				sendData2["INSTRUCTION"] = "REQUEST"
				sendData2["FILENAME"] = fileName
				sendData2["RECIEVER"] = peerAddr.String()
				utils.SendData(senderAddr, conn, sendData2)
			} else {
				sendData["INSTRUCTION"] = "NOTFOUND"
			}
			utils.SendData(peerAddr.String(), conn, sendData)
		}	
	}	
}