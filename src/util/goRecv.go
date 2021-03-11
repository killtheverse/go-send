package util

import (
	"fmt"
	// "os"
	"net"
	"strings"
	"io/ioutil"
	"errors"
)


//GoRecv - this function is exported to the main module
func GoRecv(fileName string, serverAddr string, port string) {
	fmt.Println("File name is:", fileName)
	listenAddrString, err := ExternalIP()
	if err != nil {
		panic(err)
	}
	listenAddrString = listenAddrString + port

	registerRecv(fileName, listenAddrString, serverAddr)
}

func registerRecv(fileName string, listenAddrString string, serverAddrString string) {
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
	
	handleConnectionRecv(conn)
}

func handleConnectionRecv(conn *net.UDPConn) {
	peerAddrString := ""
	for {
		fmt.Println("Listening")
		buffer := make([]byte, 1024)
		bytesRead, err := conn.Read(buffer)
		if err != nil {
			panic(err)
		}
		reply := strings.Split(string(buffer[0:bytesRead]), ",")[0]
		
		if reply == "NOTFOUND" {
			fmt.Println("Requested file not found")
		} else if reply == "SUCCESS" {
			peerAddrString = strings.Split(string(buffer[0:bytesRead]), ",")[1]
			peerAddr, _ := net.ResolveUDPAddr("udp", peerAddrString)
			for i:=0;i<1;i++ {
				conn.WriteTo([]byte("HOLEPUNCH"), peerAddr)
			}	
			fmt.Println("File found on address:", peerAddrString)
		} else if reply == "SENDING" {
			name := strings.Split(string(buffer[0:bytesRead]), ",")[1]
			ext := strings.Split(string(buffer[0:bytesRead]), ",")[2]
			name = name+"(copy)"
			fileName := name + "." + ext
			fmt.Println("Filename:", fileName)
			peerAddr, _ := net.ResolveUDPAddr("udp", peerAddrString)
			conn.WriteTo([]byte("OK"), peerAddr)
			for {
				bytesRead, err := conn.Read(buffer)
				if err != nil {
					panic(err)
				}
				data := string(buffer[:bytesRead])
				fmt.Println("data:", data)
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