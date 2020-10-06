package main

import (
	"flag"
	"web/of/science/server"
)

func main() {
	address := flag.String("address", "0.0.0.0", "listen address default is 0.0.0.0")
	port := flag.String("port", "2233", "listen port, default is 2233")
	pubPath := flag.String("spub", "", "the path to the server pub cert")
	keyPath := flag.String("skey", "", "the path the server private key")
	flag.Parse()
	s := server.NewServerType1(*address, *port, *pubPath, *keyPath)
	s.Start()
}
