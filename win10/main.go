package main

import (
	"web/of/science/client"
	localserver "web/of/science/localServer"
)

func main() {
	c := client.NewClientType1("sync.contrary.ac.cn", "2233", ".")
	localServer := localserver.NewLSocketServer("0.0.0.0", "7891", c)
	localServer.Start()
}
