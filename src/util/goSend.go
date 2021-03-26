package util

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
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
		sendString := "REGISTER," + fileName
		bytesWritten, err := conn.WriteTo([]byte(sendString), serverAddr)
		if err != nil {
			panic(err)
		}
		fmt.Println("Sending:", sendString)
		fmt.Println(bytesWritten, "bytes sent")
	} ()

	// go keepAlive(conn, serverAddrstring)

	handleConnectionSend(conn, listenAddrstring)
}

func keepAlive(conn *net.UDPConn, serverAddrstring string) {
	serverAddr, _ := net.ResolveUDPAddr("udp", serverAddrstring)
	conn.WriteTo([]byte("KEEPALIVE"), serverAddr)
	time.Sleep(10*time.Second)
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
		fmt.Println("Reply:", reply)
		if reply == "SUCCESS" {
			fmt.Println("File is ready for sharing")
		} else if reply == "REQUEST" {
			peerAddr := strings.Split(string(buffer[0:bytesRead]), ",")[1]
			fileName := strings.Split(string(buffer[0:bytesRead]), ",")[2]
			fmt.Println("File requested from:", peerAddr)
			// sendFile(fileName, peerAddr, conn)	
			holePunchSend(fileName, peerAddr, conn)
		} 
	}	
}

func holePunchSend(fileName string, peerAddrString string, conn *net.UDPConn) {
	peerAddr, _ := net.ResolveUDPAddr("udp", peerAddrString)
	for i:=0;i<1;i++ {
		fmt.Println("Sending: HOLEPUNCH")
		conn.WriteTo([]byte("HOLEPUNCH"), peerAddr)
	}	
	sendFile(fileName, peerAddrString, conn)
}

func sendFile(fileName string, peerAddrstring string, conn *net.UDPConn) {
	peerAddr, _ := net.ResolveUDPAddr("udp", peerAddrstring)
	name := strings.Split(fileName, ".")[0]
	ext := strings.Split(fileName, ".")[1]

	buffer := make([]byte, 1024)
	reply := ""
	for {
		conn.WriteTo([]byte("SENDING,"+name+","+ext), peerAddr)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			panic(err)
		}
		reply = string(buffer[0:bytesRead])
		fmt.Println("Reply:", reply)
		if reply == "OK" {
			break
		}
	}
	
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
