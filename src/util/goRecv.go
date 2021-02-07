package util

import (
	"fmt"
	// "os"
	"net"
	"strings"
	"io/ioutil"
)


//GoRecv - this function is exported to the main module
func GoRecv(fileName string, listenAddr string) {
	
	fmt.Println("File name is:", fileName)
	registerRecv(fileName, listenAddr)
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		panic(err)
	}
	fmt.Println("Listening on:", listenAddr)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go handleConnectionRecv(conn)
	}
	
}

func registerRecv(fileName string, listenAddr string) {
	conn, err := net.Dial("tcp", "127.0.0.1:8000")
	if err!= nil {
		fmt.Println(err)
	}
	sendString := "CHECK," + fileName + "," + listenAddr
	fmt.Println("Sending:", sendString)
	conn.Write([]byte(sendString))
}

func handleConnectionRecv(conn net.Conn) {
	buffer := make([]byte, 1024)
	bytesRead, err := conn.Read(buffer)
	if err != nil {
		panic(err)
	}
	reply := strings.Split(string(buffer[0:bytesRead]), ",")[0]
	// fmt.Println("Buffer:", reply)
	if reply == "NOTFOUND" {
		fmt.Println("Requested file not found")
	} else if reply == "SUCCESS" {
		peerAddr := strings.Split(string(buffer[0:bytesRead]), ",")[1]
		fmt.Println("File found on address:", peerAddr)
	} else if reply == "SENDING" {
		name := strings.Split(string(buffer[0:bytesRead]), ",")[1]
		ext := strings.Split(string(buffer[0:bytesRead]), ",")[2]
		name = name+"(copy)"
		fileName := name + "." + ext
		
		for {
			bytesRead, err := conn.Read(buffer)
			if err != nil {
				panic(err)
			}
			data := string(buffer[:bytesRead])
			if data == "EXIT" {
				break

			} else {
				fmt.Println("Writing", data)
				err := ioutil.WriteFile(fileName, buffer[:bytesRead], 0777)
				if err != nil {
					fmt.Println(err)
				}
			}
		}

		fmt.Println("File recieved")
	}
	conn.Close()
}