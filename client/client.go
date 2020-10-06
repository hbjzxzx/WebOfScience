package client

import (
	"net"
	"web/of/science/pb"
)

//Client provide a interface to handle local connection
type Client interface {
	OnRequest(address, port string, atype pb.AddressType, conn net.Conn)
}

