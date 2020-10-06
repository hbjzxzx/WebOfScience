package main

import (
	"flag"
	"web/of/science/client"
	localserver "web/of/science/localServer"
)

func main() {
	laddress := flag.String("laddress", "0.0.0.0", "local listen address default is 0.0.0.0")
	lport := flag.String("lport", "7891", "local listen port, default is 7891")
	address := flag.String("address", "0.0.0.0", "remote address default is 0.0.0.0")
	port := flag.String("port", "2233", "remote listen port, default is 2233")
	pubPath := flag.String("pub", ".", "the path to the client pub cert")
	flag.Parse()

	c := client.NewClientType1(*address, *port, *pubPath)
	localServer := localserver.NewLSocketServer(*laddress, *lport, c)
	localServer.Start()
}
