package localserver

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"web/of/science/client"
	"web/of/science/pb"
)

func init() {
	log.SetPrefix("local-socket5-server: ")
}

//LSocketServer provide socket5 service
type LSocketServer struct {
	handle        client.Client
	address, port string
}

//NewLSocketServer return a new local socket5 server
func NewLSocketServer(address, port string, h client.Client) *LSocketServer {
	return &LSocketServer{
		address: address,
		port:    port,
		handle:  h,
	}
}

//Start run the local server
func (s *LSocketServer) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.address, s.port))
	defer lis.Close()
	if err != nil {
		log.Fatal("can not start a tcp listener: ", err)
	}
	for {
		con, err := lis.Accept()
		if err != nil {
			log.Println("receive a bad connection ", getConInfo(con))
			continue
		}
		go handleLocalConn(con, s.handle)
	}
}

func socket5Handshake(con net.Conn) (address, port string, addressType pb.AddressType, e error) {
	e = errors.New("handshake fail")
	data := make([]byte, 512)
	n, err := con.Read(data)
	if err != nil {
		log.Println("error handle information", getConInfo(con))
		return
	}
	//only support the socket5
	if data[0] != 0x5 {
		log.Println("error not socket5 clients version:", data[0], getConInfo(con))
		return
	}
	methodsCnt := int(data[1])
	if n-2 != methodsCnt {
		//error socket5 handle information
		//To do, here need to check whether the methods contains 0x00
		log.Println("error handle information", getConInfo(con))
		return
	}

	//clear buffer
	for i := 0; i < n; i++ {
		data[i] = 0x00
	}
	//good handleshake. return the client
	reply := []byte{0x05, 0x00}
	n, err = con.Write(reply)
	if err != nil {
		log.Println("error while response to handshake", getConInfo(con))
		return
	}
	if n != 2 {
		log.Println("error response partial write", getConInfo(con))
		return
	}

	//start handle Connect request
	n, err = con.Read(data)
	if err != nil {
		log.Println("error connection request", getConInfo(con))
		return
	}
	if data[0] != 0x05 {
		log.Println("error not socket5 connection request", getConInfo(con))
		return
	}
	if data[1] != 0x01 {
		log.Println("error only support TCP", getConInfo(con))
		return
	}
	switch data[3] {
	case 0x01: //ipv4
		address = net.IPv4(data[4], data[5], data[6], data[7]).String()
		addressType = pb.AddressType_Ipv4
	case 0x03: //hostname
		address = string(data[5 : n-2])
		addressType = pb.AddressType_HostName
	default: //not support
		log.Println("error not support address:", data[3], getConInfo(con))
		return
	}
	port = strconv.Itoa(int(data[n-2])<<8 | int(data[n-1]))
	e = nil
	return
}

func handleLocalConn(con net.Conn, h client.Client) {
	address, port, addressType, err := socket5Handshake(con)
	if err != nil {
		con.Close()
		return
	}
	h.OnRequest(address, port, addressType, con)

}

func getConInfo(con net.Conn) string {
	return fmt.Sprintf("local: %v, remote: %v", con.LocalAddr(), con.RemoteAddr())

}
