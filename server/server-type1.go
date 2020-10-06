package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"web/of/science/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

var localIP string

func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func init() {
	log.SetPrefix("Server of wos : ")
	data := GetOutboundIP()
	localIP = net.IPv4(data[0], data[1], data[2], data[3]).String()
	fmt.Sprintf("local ip is: %v", localIP)
}

type serverType1 struct {
	pb.UnimplementedWebOfScienceServer
	pubPath, keyPath string
	address, port    string
}

func proxyWorker(lis net.Listener, magic []byte, remoteConn net.Conn) {
	defer lis.Close()
	defer remoteConn.Close()
	conn, err := lis.Accept()
	if err != nil {
		log.Print(err)
		return
	}
	log.Printf("client tcp in !")
	defer conn.Close()
	go io.Copy(conn, remoteConn)
	io.Copy(remoteConn, conn)
}

//Request on server handle client Request
func (s *serverType1) Request(ctx context.Context, cr *pb.ConnectRequest) (*pb.ConnectResponse, error) {
	address := cr.GetAddress()
	port := cr.GetPort()
	aType := cr.GetAddressType()
	log.Printf("new connection Request %v:%v, %v", address, port, aType)

	remoteConn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", address, port))
	if err != nil {
		log.Printf("%v:%v can not reach", address, port)
		return nil, fmt.Errorf("%v:%v can not reach", address, port)
	}

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Printf("listen failed for %v:%v", address, port)
		return nil, fmt.Errorf("error to open new listening")
	}
	portStr := fmt.Sprintf("%d", (listener.Addr().(*net.TCPAddr).Port))
	log.Printf("listen port %v", portStr)
	bindPort := fmt.Sprintf("%d", remoteConn.LocalAddr().(*net.TCPAddr).Port)
	go proxyWorker(listener, []byte{0x01}, remoteConn)
	r := pb.ConnectResponse{
		BindAddress: localIP,
		BindPort:    bindPort,
		AType:       pb.AddressType_Ipv4,
		MagicPort:   portStr,
		MagicToken:  []byte{0x22, 0x33, 0x33, 0x22},
	}
	if ctx.Err() == context.Canceled {
		log.Print("request is canceled")
		return nil, status.Error(codes.Canceled, "request is canceled")
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Print("deadline is exceeded")
		return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	}
	return &r, nil
}

func (s *serverType1) loadTLSCredentials() (credentials.TransportCredentials, error) {
	cerds, err := credentials.NewServerTLSFromFile(s.pubPath, s.keyPath)
	if err != nil {
		return nil, err
	}
	return cerds, nil
}

//Start make server working
func (s *serverType1) Start() {
	//tlsCredentials, err := s.loadTLSCredentials()
	//if err != nil {
	//	log.Fatal("cannot load TLS credentials: ", err)
	//}

	//grpcServer := grpc.NewServer(grpc.Creds(tlsCredentials))
	grpcServer := grpc.NewServer()
	pb.RegisterWebOfScienceServer(grpcServer, s)
	address := fmt.Sprintf("%s:%s", s.address, s.port)
	listener, err := net.Listen("tcp", address)
	defer listener.Close()
	if err != nil {
		log.Fatal("cannot listening the port: ", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}

//NewServerType1 return a ready run server
func NewServerType1(address, port, serverPubPath, serverKeyPath string) Server {
	return &serverType1{
		pubPath: serverPubPath,
		keyPath: serverKeyPath,
		address: address,
		port:    port,
	}
}
