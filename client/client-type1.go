package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"web/of/science/pb"
	"web/of/science/utils"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func init() {
	log.SetPrefix("Client of wos : ")
}

type clientType1 struct {
	pubPath       string
	address, port string
	rpc           pb.WebOfScienceClient
}

func (c *clientType1) loadTLSCredentials() (credentials.TransportCredentials, error) {

	creds, err := credentials.NewClientTLSFromFile(c.pubPath, "localhost")
	if err != nil {
		return nil, err
	}
	return creds, nil
}

func NewClientType1(address, port, pub string) Client {
	c := clientType1{
		pubPath: pub,
		address: address,
		port:    port,
	}
	//tlsCredentials, err := c.loadTLSCredentials()
	//if err != nil {
	//	log.Fatal("cannot load TLS credentials: ", err)
	//}

	addressPort := fmt.Sprintf("%s:%s", c.address, c.port)
	//conn, err := grpc.Dial(addressPort, grpc.WithTransportCredentials(tlsCredentials))
	conn, err := grpc.Dial(addressPort, grpc.WithInsecure())
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}
	c.rpc = pb.NewWebOfScienceClient(conn)
	return &c
}

func (c *clientType1) OnRequest(address, port string, atype pb.AddressType, conn net.Conn) {
	//log.Printf("new connect request, %v:%v, %v", address, port, addressType)
	defer conn.Close()
	r, err := c.rpc.Request(context.Background(), &pb.ConnectRequest{
		Address:     address,
		Port:        port,
		AddressType: atype,
	})
	if err != nil {
		log.Printf("Connect request, %v:%v, %v failed:", address, port, atype, err)
		return
	}
	bindPort := r.GetBindPort()
	bindAddress := r.GetBindAddress()
	magicPort := r.GetMagicPort()
	magicToken := r.GetMagicToken()
	log.Printf("Connect request %v:%v, %v success with %v:%v magic port:%v token:%v", address, port, atype, bindAddress, bindPort, magicPort, magicToken)

	proxy_link, err := net.Dial("tcp", fmt.Sprintf("%s:%s", bindAddress, magicPort))
	if err != nil {
		log.Printf("Connect to proxy server request, %v:%v failed:%v", bindAddress, bindPort, err)
		return
	}
	defer proxy_link.Close()
	//fmt.Println(bindAddress)
	ipBytes := []byte(net.ParseIP(bindAddress))[12:]
	//fmt.Println(ipBytes)
	prefixBytes := []byte{0x05, 0x00, 0x00, 0x01}

	intPort, err := strconv.Atoi(bindPort)
	iport := int16(intPort)
	portBytes := []byte{byte(iport), byte(iport >> 8)}

	responBytes := bytes.Join([][]byte{prefixBytes, ipBytes, portBytes}, []byte{})
	//fmt.Println(responBytes)
	conn.Write(responBytes)

	//proxy_link.Write(magicToken)

	//go io.Copy(proxy_link, conn)
	//_, err = io.Copy(conn, proxy_link)
	//go copy(proxy_link, conn, "reveive", true)
	//copy(conn, proxy_link, "send", false)
	//log.Printf("link down: %v", err)

	inputByteCnt := make(chan int64, 1)
	outputByteCnt := make(chan int64, 1)

	bsize := utils.GetBlockSize()
	encrypt := utils.AesEncrypt
	go utils.Forward(conn, proxy_link, outputByteCnt, bsize, encrypt)
	err = utils.Forward(proxy_link, conn, inputByteCnt, bsize, encrypt)
	//inBytes := utils.Int64ToKB(<-inputByteCnt)
	//outBytes := utils.Int64ToKB(<-outputByteCnt)

	inBytes := <-inputByteCnt
	outBytes := <-outputByteCnt
	linkStr := fmt.Sprintf("request to %v:%v, proxy by %v:%v", address, port, bindAddress, bindPort)
	//flowStr := fmt.Sprintf("flowout:%.2f KB flowin:%.2f KB", outBytes, inBytes)
	flowStr := fmt.Sprintf("flowout:%d B flowin:%d B", outBytes, inBytes)
	if err != nil {
		log.Printf("link down abnormal: %v  (%v)", err, linkStr)
	}
	log.Printf("link(%v) exit; flow: %v", linkStr, flowStr)

}

func copy(i io.Reader, o io.Writer, tag string, s bool) {
	b := make([]byte, 128)
	for {
		_, err := i.Read(b)
		o.Write(b)
		if s {
			log.Printf("%v--> %v", tag, b)
		}
		if err != nil {
			break
		}
	}
}
