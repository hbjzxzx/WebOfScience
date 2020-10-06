package localserver

//LocalServer interface provide a run method
type LocalServer interface {
	Start(localAddress, localPort string)
}
