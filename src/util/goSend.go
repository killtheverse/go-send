package util

import (
	"bufio"
	"fmt"
	"os"
	"net"
	"strings"
	"io"
)

//GoSend - this function is exported to the main module
func GoSend(fileName string, serverAddr string, port string) {
	fmt.Println("File name is:", fileName)
	listenAddrString, err := ExternalIP()
	if err != nil {
		panic(err)
	}
	listenAddrString = listenAddrString + port
	registerSend(fileName, listenAddrString, serverAddr)
}

func registerSend(fileName string, listenAddrstring string, serverAddrstring string) {
	fmt.Println("Dialing:", serverAddrstring)
	listenAddr, _ := net.ResolveUDPAddr("udp", listenAddrstring)
	serverAddr, _ := net.ResolveUDPAddr("udp", serverAddrstring)
	conn, err := net.ListenUDP("udp", listenAddr)
	if err!= nil {
		fmt.Println(err)
	}

	go func() {
		sendString := "REGISTER," + fileName + "," + listenAddrstring
		bytesWritten, err := conn.WriteTo([]byte(sendString), serverAddr)
		if err != nil {
			panic(err)
		}
		fmt.Println("Sending:", sendString)
		fmt.Println(bytesWritten, "bytes sent")
	} ()

	handleConnectionSend(conn, listenAddrstring)
}

func handleConnectionSend(conn *net.UDPConn, listenAddr string) {

	for {
		fmt.Println("Listening")
		buffer := make([]byte, 1024)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			panic(err)
		}
		reply := strings.Split(string(buffer[0:bytesRead]), ",")[0]
		fmt.Println("Buffer:", string(buffer[0:bytesRead]))
		if reply == "SUCCESS" {
			fmt.Println("File is ready for sharing")
		} else if reply == "REQUEST" {
			peerAddr := strings.Split(string(buffer[0:bytesRead]), ",")[1]
			fileName := strings.Split(string(buffer[0:bytesRead]), ",")[2]
			fmt.Println("File requested from:", peerAddr)
			sendFile(fileName, listenAddr, peerAddr, conn)	
		}
	}	
}

func sendFile(fileName string, listenAddrstring string, peerAddrstring string, conn *net.UDPConn) {
	
	peerAddr, _ := net.ResolveUDPAddr("udp", peerAddrstring)
	name := strings.Split(fileName, ".")[0]
	ext := strings.Split(fileName, ".")[1]
	conn.WriteTo([]byte("SENDING,"+name+","+ext), peerAddr)

	buffer := make([]byte, 1024)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		panic(err)
	}
	reply := string(buffer[0:bytesRead])
	fmt.Println("Reply:", reply)
	if reply == "OK" {
		readFile(fileName, conn, peerAddrstring)
		fmt.Println("File sent")
		conn.WriteTo([]byte("EXIT"), peerAddr)
	} else {
		fmt.Println("Can't establish connection")
	}	
}

func readFile(fileName string, conn *net.UDPConn, peerAddrString string) {
	peerAddr, _ := net.ResolveUDPAddr("udp", peerAddrString)
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println("Can't read the file", err)
		panic(err)
	}
	defer f.Close()
	
	r := bufio.NewReader(f)
	for {
		buf := make([]byte,4*1024) 
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
		fmt.Println("SENDING:", string(buf))
		fmt.Println("=====================")
		conn.WriteTo(buf, peerAddr)
	}
}
