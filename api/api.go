package api

// The deCerver API
//
// There are two ways to communicate with deCerver. The first way is stateless, via http-based
// rpc. The RPC protocol we use is JSON-RPC 2.0 (http://www.jsonrpc.org/specification), and
// the implementation we use is http://www.gorillatoolkit.org/pkg/rpc. The site contains a good
// example on how to write an RPC service. Basically, an rpc function takes three arguments -
// a pointer to a "net" http request, a pointer to an arguments object, and a pointer to a
// reply object. It must also return an 'error'.
//
//
// The second way is stateful, via a websocket connection. The RPC specification is a fork of
// JSON-RPC 2.0. An rpc function takes two arguments - a pointer to a Request object, and a
// pointer to a Response object, as per the type 'WsAPIMethod'
//
// TODO Write up the websocket rpc protocol and explain how to create a service.
//
// Stateful connections is the normal way to communicate between a UI and deCerver, as it allows
// two-way communication. Http can be used in cases when you only want to do some occasional polling.
import (
	"encoding/json"
	"github.com/robertkrimen/otto"
)

type ErrorCode int

const (
	E_PARSE       ErrorCode = -32700
	E_INVALID_REQ ErrorCode = -32600
	E_NO_METHOD   ErrorCode = -32601
	E_BAD_PARAMS  ErrorCode = -32602
	E_INTERNAL    ErrorCode = -32603
	E_SERVER      ErrorCode = -32000
)

// A JSON-RPC message received by the server.
type Request struct {
	// Should always be set to '2.0'
	JsonRPC string `json:"jsonrpc"`
	// A String containing the name of the method to be invoked.
	Method string `json:"method"`
	// An Array of objects to pass as arguments to the method.
	Params *json.RawMessage `json:"params"`
	// Id
	Id int `json:"id"`
	// Timestamp
	Timestamp int `json:"timestamp"`
}

// A JSON-RPC message sent by the server.
type Response struct {
	// The Object that was returned by the invoked method. This must be null
	// in case there was an error invoking the method.
	Result interface{} `json:"result"`
	// An Error object if there was an error invoking the method. It must be
	// null if there was no error.
	Error interface{} `json:"error"`
	// Timestamp
	Timestamp int `json:"timestamp"`
	// The name of the Object that was returned by the invoke method.
	Id string `json:"id"`
}

var null = json.RawMessage([]byte("null"))

// Errors sent in api call responses.
type Error struct {
	// A Number that indicates the error type that occurred.
	Code ErrorCode `json:"code"` /* required */
	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"` /* required */
	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data"` /* optional */
}

func (e *Error) Error() string {
	return e.Message
}

type WebSocketObj interface {
	SessionId() uint32
	WriteTextMsg(msg interface{})
	WriteCloseMsg()
}

// Function that handles socket based rpc requests
type WsAPIMethod func(*Request, *Response)

// A stateful (websocket-based) rpc service.
type WsAPIService interface {
	Name() string
	Init()
	HandleRPC(*Request) (*Response, error)
	SetConnection(wsConn WebSocketObj)
	Shutdown()
}

// A factory for creating SRPCServices of a given kind
type WsAPIServiceFactory interface {
	CreateService() WsAPIService
	Init()
	ServiceName() string
	Shutdown()
}

// A function that can be added to the otto vm.
type AteFunc func(otto.FunctionCall) otto.Value

type ApiRegistry interface {
	RegisterHttpServices(service ...interface{})
	RegisterWsServiceFactories(factory ...WsAPIServiceFactory)
}

type NetHandler interface {
	HandleInboundHttp(*Request) *Response
	HandleInboundWs(*Request)
	HandleOutboundWs(*Response)
	// TODO handlers for inbound/outbound stream
}
