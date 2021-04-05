package goSend

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"encoding/json"

	"github.com/killtheverse/go-send/src/utils"
)

//GoSend - this function is exported to the main module
func GoSend(fileName string, serverAddr string, port string) {
	fmt.Println("File name is:", fileName)
	listenAddrString, err := utils.ExternalIP()
	if err != nil {
		panic(err)
	}
	listenAddrString = listenAddrString + port
	registerSend(fileName, listenAddrString, serverAddr)
}

func registerSend(fileName string, listenAddrstring string, serverAddrstring string) {
	fmt.Println("Dialing:", serverAddrstring)
	listenAddr, _ := net.ResolveUDPAddr("udp", listenAddrstring)
	conn, err := net.ListenUDP("udp", listenAddr)
	if err!= nil {
		fmt.Println(err)
	}

	go func() {
		sendData := make(map[string]string)
		sendData["INSTRUCTION"] = "REGISTER"
		sendData["FILENAME"] = fileName
		utils.SendData(serverAddrstring, conn, sendData)
	} ()

	go utils.KeepAlive(conn, serverAddrstring)

	handleConnectionSend(conn, listenAddrstring)
}

func handleConnectionSend(conn *net.UDPConn, listenAddr string) {
	buffer := make([]byte, 1024)
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
		if instruction == "SUCCESS" {
			fmt.Println("File is ready for sharing")
		} else if instruction == "REQUEST" {
			peerAddr := recvData["RECIEVER"]
			fileName := recvData["FILENAME"]
			fmt.Println("File requested from:", peerAddr)
			data := make(map[string]string)
			data["INSTRUCTION"] = "HOLEPUNCH"
			utils.AwaitResponse(peerAddr, conn, data, "HOLEPUNCH")
			sendFile(fileName, peerAddr, conn)
		}
	}	
}

func sendFile(fileName string, peerAddrString string, conn *net.UDPConn) {
	data := make(map[string]string)
	name := strings.Split(fileName, ".")[0]
	ext := strings.Split(fileName, ".")[1]
	data["INSTRUCTION"] = "SENDFILE"
	data["FILENAME"] = name
	data["EXTENSION"] = ext
	utils.AwaitResponse(peerAddrString, conn, data, "SENDFILE-OK")

	data2 := make(map[string]string)
	data2["INSTRUCTION"] = "REQUESTTCP"
	data2["FILENAME"] = fileName 
	recvData := utils.AwaitResponse(peerAddrString, conn, data2, "REQUESTTCP-OK")
	
	tcpAddrString := strings.Split(peerAddrString, ":")[0] + ":" + recvData["TCPPORT"]
	fmt.Println("Tcpaddress:", tcpAddrString)
	tcpConn, err := net.Dial("tcp", tcpAddrString)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	readFile(fileName, tcpConn, peerAddrString)
	fmt.Println("File sent")
	os.Exit(0)
}


func readFile(fileName string, conn net.Conn, peerAddrString string) {
	
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Can't read the file", err)
		panic(err)
	}
	defer f.Close()
	
	r := bufio.NewReader(f)
	
	for {
		buf := make([]byte, 4*1024) 
		n, err := r.Read(buf) 
		buf = buf[:n]
		
		if n == 0 {
			if err != nil {
				fmt.Println(err)
				break
			}
			if err == io.EOF {
				break
			}
			break
		}
		
		_, err = conn.Write(buf)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	conn.Write([]byte("EXIT"))
}