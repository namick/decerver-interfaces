package network

// Since most networking takes place inside the js runtime, this is
// not used a lot.
import (
)

// Websocket rpc
type WsSession interface {
	SessionId() uint32
	WriteJsonMsg(msg interface{})
	WriteCloseMsg()
}


// Webserver
type Server interface {
	RegisterDapp(dappId string)
}
