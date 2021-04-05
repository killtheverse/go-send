package goRecv

import (
	"fmt"
	"os"
	"net"
	"strings"
	"io/ioutil"
	"encoding/json"

	"github.com/killtheverse/go-send/src/utils"
)


//GoRecv - this function is exported to the main module
func GoRecv(fileName string, serverAddr string, port string, tcpport string) {
	fmt.Println("File name is:", fileName)
	listenAddrOnly, err := utils.ExternalIP()
	if err != nil {
		panic(err)
	}
	listenAddrString := listenAddrOnly + port
	tcpAddrString := listenAddrOnly + tcpport

	registerRecv(fileName, listenAddrString, serverAddr, tcpAddrString)
}

func registerRecv(fileName string, listenAddrString string, serverAddrString string, tcpAddrString string) {
	fmt.Println("Dialing:", serverAddrString)
	listenAddr, _ := net.ResolveUDPAddr("udp", listenAddrString)
	conn, err := net.ListenUDP("udp", listenAddr)
	if err!= nil {
		fmt.Println(err)
	}
	
	go func() {
		data := make(map[string]string)
		data["INSTRUCTION"] = "CHECK"
		data["FILENAME"] = fileName
		utils.SendData(serverAddrString, conn, data)
	} ()
	
	handleConnectionRecv(conn, tcpAddrString)
}

func handleConnectionRecv(conn *net.UDPConn, tcpAddrString string) {
	buffer := make([]byte, 4*1024)
	peerAddrString := ""
	fileName := ""
	for {
		fmt.Println("Listening")
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			panic(err)
		}
		recvData := make(map[string]string)
		err = json.Unmarshal(buffer[0:bytesRead], &recvData)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		fmt.Printf("\n%+v\n", recvData)
		instruction := recvData["INSTRUCTION"]
		if instruction == "NOTFOUND" {
			fmt.Println("Requested file not found")
			os.Exit(0)
		} else if instruction == "SUCCESS" {
			peerAddrString = recvData["SENDER"]
			data := make(map[string]string)
			data["INSTRUCTION"] = "HOLEPUNCH"
			utils.AwaitResponse(peerAddrString, conn, data, "HOLEPUNCH")
			fmt.Println("File found on address:", peerAddrString)
		} else if instruction == "REQUESTTCP" {
			data := make(map[string]string)
			data["INSTRUCTION"] = "REQUESTTCP-OK"
			data["TCPPORT"] = strings.Split(tcpAddrString, ":")[1]
			fmt.Println("filename:", fileName)
			ln, _ := net.Listen("tcp", tcpAddrString)
			utils.SendData(peerAddrString, conn, data)
			for {
				fmt.Println("Listening TCP on", tcpAddrString)
				tcpConn, _ := ln.Accept()
				recieveFile(fileName, tcpConn)
				fmt.Println("File recieved")
				os.Exit(0)
			}

		} else if instruction == "SENDFILE" {
			fileName = recvData["FILENAME"] + "(copy)." + recvData["EXTENSION"]
			data := make(map[string]string)
			data["INSTRUCTION"] = "SENDFILE-OK"
			utils.SendData(peerAddrString, conn, data)
		}
	}
}

func recieveFile(fileName string, conn net.Conn) {
	
	buffer := make([]byte, 4*1024)
	for {
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			break
		}
		data := string(buffer[:bytesRead])
		
		if data == "EXIT" {
			break
		} else {
			f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0600)
			if err != nil {
				ioutil.WriteFile(fileName, buffer[:bytesRead], 0777)
			}
			defer f.Close()

			f.Write(buffer[:bytesRead])
		}
	}
}
