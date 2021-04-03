package util

import (
	"fmt"
	"os"
	"net"
	"strings"
	"io/ioutil"
	"errors"
)


//GoRecv - this function is exported to the main module
func GoRecv(fileName string, serverAddr string, port string, tcpport string) {
	fmt.Println("File name is:", fileName)
	listenAddrOnly, err := ExternalIP()
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
	serverAddr, _ := net.ResolveUDPAddr("udp", serverAddrString)
	conn, err := net.ListenUDP("udp", listenAddr)
	if err!= nil {
		fmt.Println(err)
	}
	
	go func() {
		sendString := "CHECK," + fileName 
		bytesWritten, err := conn.WriteTo([]byte(sendString), serverAddr)
		if err != nil {
			panic(err)
		}
		fmt.Println("Sending:", sendString)
		fmt.Println(bytesWritten, "bytes sent")
	} ()
	
	handleConnectionRecv(conn, tcpAddrString)
}

func handleConnectionRecv(conn *net.UDPConn, tcpAddrString string) {
	peerAddrString := ""
	for {
		fmt.Println("Listening")
		buffer := make([]byte, 4*1024)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			panic(err)
		}
		fmt.Println("buffer:", string(buffer[0:bytesRead]))
		reply := strings.Split(string(buffer[0:bytesRead]), ",")[0]
		fmt.Println("Reply:", reply)
		if reply == "NOTFOUND" {
			fmt.Println("Requested file not found")
		} else if reply == "SUCCESS" {
			peerAddrString = strings.Split(string(buffer[0:bytesRead]), ",")[1]
			peerAddr, _ := net.ResolveUDPAddr("udp", peerAddrString)
			fmt.Println("PUNCHING HOLE")
			for i:=0;i<1;i++ {
				conn.WriteTo([]byte("HOLEPUNCH"), peerAddr)
			}	
			fmt.Println("File found on address:", peerAddrString)
		} else if reply == "TCPADDRESS" {
			peerAddr, _ := net.ResolveUDPAddr("udp", peerAddrString)
			fmt.Println("Sending:",":"+strings.Split(tcpAddrString, ":")[1])
			conn.WriteTo([]byte(":"+strings.Split(tcpAddrString, ":")[1]), peerAddr)

			name := strings.Split(string(buffer[0:bytesRead]), ",")[1]
			ext := strings.Split(string(buffer[0:bytesRead]), ",")[2]
			name = name+"(copy)"
			fileName := name + "." + ext
			fmt.Println("Filename:", fileName)

			ln, _ := net.Listen("tcp", tcpAddrString)
			for {
				fmt.Println("Listening TCP on", tcpAddrString)
				tcpConn, _ := ln.Accept()
				recieveFile(fileName, tcpConn)
				fmt.Println("File recieved")
				os.Exit(0)
			}
		} else if reply == "SENDING" {
			
			peerAddr, _ := net.ResolveUDPAddr("udp", peerAddrString)
			conn.WriteTo([]byte("OK"), peerAddr)
			
			
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

// ExternalIP is exported for use in goSend.go
func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}