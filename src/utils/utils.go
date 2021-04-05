package utils

import (
	"net"
	"errors"
	"encoding/json"
	"fmt"
	"time"
)


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
	return "", errors.New("Are you connected to the network?")
}

func SendData(address string, conn *net.UDPConn, message map[string]string) {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	
	jsonObject, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	fmt.Printf("\n%+v\n", message)
	conn.WriteTo(jsonObject, addr)
}

func AwaitResponse(address string, conn *net.UDPConn, data map[string]string, reply string) map[string]string {
	success := false
	var m map[string]string 
	go func () {
		buffer := make([]byte, 1024)
		for {
			bytesRead, err := conn.Read(buffer)
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			recvData := make(map[string]string)
			err = json.Unmarshal(buffer[0:bytesRead], &recvData)
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			response := recvData["INSTRUCTION"]
			if response == reply {
				success = true
				m = recvData
				return
			}
		}
	} ()

	for {
		SendData(address, conn, data)
		time.Sleep(3*time.Second)
		if success == true {
			return m
		}
	}
}

func KeepAlive(conn *net.UDPConn, serverAddrstring string) {
	for {
		data := make(map[string]string)
		data["INSTRUCTION"] = "KEEPALIVE"
		SendData(serverAddrstring, conn, data)
		time.Sleep(10*time.Second)
	}
}