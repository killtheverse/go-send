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
func GoRecv(fileName string, serverAddr string) {
	
	fmt.Println("File name is:", fileName)
	listenAddr, err := ExternalIP()
	if err != nil {
		panic(err)
	}
	listenAddr = listenAddr + ":9000"
	
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		listenAddr = strings.Split(listenAddr, ":")[0]
		listenAddr = listenAddr + ":10000"
		l, _ = net.Listen("tcp", listenAddr)
	}

	fmt.Println("Listening on:", listenAddr)
	registerRecv(fileName, listenAddr, serverAddr)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go handleConnectionRecv(conn)
	}
	
}

func registerRecv(fileName string, listenAddr string, serverAddr string) {
	fmt.Println("Dialing:", serverAddr)
	conn, err := net.Dial("tcp", serverAddr)
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
	conn.Close()
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