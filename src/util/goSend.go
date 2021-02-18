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
func GoSend(fileName string, serverAddr string) {
	
	fmt.Println("File name is:", fileName)
	listenAddr, err := ExternalIP()
	if err != nil {
		panic(err)
	}
	// fmt.Println("address is:", listenAddr)
	listenAddr = listenAddr + ":9000"
	
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		listenAddr = strings.Split(listenAddr, ":")[0]
		listenAddr = listenAddr + ":11000"
		l, _ = net.Listen("tcp", listenAddr)
	}
	fmt.Println("Listening on:", listenAddr)
	registerSend(fileName, listenAddr, serverAddr)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go handleConnectionSend(conn)
	}
	
}

func registerSend(fileName string, listenAddr string, serverAddr string) {
	fmt.Println("Dialing:", serverAddr)
	conn, err := net.Dial("tcp", serverAddr)
	if err!= nil {
		fmt.Println(err)
	}
	sendString := "REGISTER," + fileName + "," + listenAddr
	fmt.Println("Sending:", sendString)
	conn.Write([]byte(sendString))
}

func handleConnectionSend(conn net.Conn) {
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
		sendFile(fileName, peerAddr)	
	}

	conn.Close()
}

func sendFile(fileName string, peerAddr string) {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		panic(err)
	}
	name := strings.Split(fileName, ".")[0]
	ext := strings.Split(fileName, ".")[1]
	conn.Write([]byte("SENDING,"+name+","+ext))
	readFile(fileName, conn)
	fmt.Println("File sent")
	conn.Write([]byte("EXIT"))
}

func readFile(fileName string, conn net.Conn) {
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
		conn.Write(buf)
	}
}
