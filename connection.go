//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"github.com/satori/go.uuid"
	"github.com/yyyar/yamp-go/format"
	"github.com/yyyar/yamp-go/parser"
	"github.com/yyyar/yamp-go/transport"
	"log"
	"reflect"
)

//
// Callback for event and reqest/response handlers
//
type CallbackFunc interface{}

//
// Connection is Yamp connection abstraction supports
// events sending/handling and request/response
// processing
//
// TODO: Caution! Add synchronize access to handler maps.
//
type Connection struct {

	//
	// Transport connection adapter
	//
	conn transport.Connection

	//
	// User messages body format parser/serializer
	//
	bodyFormat format.BodyFormat

	//
	// Yamp protocol parser
	//
	parser *parser.Parser

	//
	// Channel for pushing frames that will be written to other party
	//
	messagesOut chan (parser.Frame)

	//
	// --------------- user callbacks storage ---------------
	//

	//
	// User callbacks for events
	//
	eventHandlers map[string][]CallbackFunc

	//
	// User callbacks for responses
	//
	responseHandlers map[uuid.UUID]CallbackFunc

	//
	// User callbacks for requests
	//
	requestHandlers map[string]CallbackFunc
}

//
// NewConnection Creates new instance of Connection wrapping
// transport.Connection and immediately starting read/write loop
//
func NewConnection(conn transport.Connection, bodyFormat format.BodyFormat) *Connection {

	connection := &Connection{

		conn:       conn,
		bodyFormat: bodyFormat,
		parser:     parser.NewParser(conn),

		messagesOut: make(chan parser.Frame),

		eventHandlers:    make(map[string][]CallbackFunc),
		responseHandlers: make(map[uuid.UUID]CallbackFunc),
		requestHandlers:  make(map[string]CallbackFunc),
	}

	go connection.loop()

	return connection
}

//
// loop is parseing / serializing loop. It works
// until it gets parser.Messages channel close event,
// i.e. while underlying reader is live
//
func (this *Connection) loop() {

	for {
		select {

		// Got new frame for sending
		case frame := <-this.messagesOut:
			frame.Serialize(this.conn)

		// Got new parsed frame
		case frame, ok := <-this.parser.Messages:

			if !ok {
				log.Println(<-this.parser.Error)
				return
			}

			// Dispatch new frame
			switch frame.GetType() {
			case parser.EVENT:
				this.handleEvent(*(frame).(*parser.Event))
			case parser.RESPONSE:
				this.handleResponse(*(frame).(*parser.Response))
			case parser.REQUEST:
				this.handleRequest(*(frame).(*parser.Request))
			default:
				log.Println("Unhandled frame", frame.GetType(), frame)

			}
		}

	}

}

//
// Handles event frame. Iterates over all event uri subscribers
// and executes callback in gourutines
//
func (this *Connection) handleEvent(message parser.Event) {

	handlers, ok := this.eventHandlers[message.Uri]

	if !ok {
		log.Println("No handlers for event uri " + message.Uri)
		return
	}

	for _, handler := range handlers {

		// infer callback first parameter type to cast
		// message body properly

		handlerType := reflect.TypeOf(handler)
		handlerArg := reflect.New(handlerType.In(0)).Interface()

		this.bodyFormat.Parse(message.Body, handlerArg)

		go reflect.ValueOf(handler).Call([]reflect.Value{reflect.Indirect(reflect.ValueOf(handlerArg))})
	}

}

//
// Handles request frame
//
func (this *Connection) handleRequest(message parser.Request) {

	handler, ok := this.requestHandlers[message.Uri]

	if !ok {
		log.Println("No handlers for request", message.Uri)
		return
	}

	// Response sender function
	// Made in generic way to fit any params signature
	responseSender := func(in []reflect.Value) []reflect.Value {
		resp := in[0].Interface()
		b, _ := this.bodyFormat.Serialize(resp)
		r := parser.Response{
			UserHeader: parser.UserHeader{
				Uid: uuid.NewV1(),
				Uri: message.Uri,
			},
			RequestUid: message.Uid,
			Type:       parser.RESPONSE_DONE,
			UserBody: parser.UserBody{
				Body: b,
			},
		}
		this.messagesOut <- &r
		return nil
	}

	handlerType := reflect.TypeOf(handler)

	responseSenderArg := reflect.New(handlerType.In(1))
	responseSenderWrap := reflect.MakeFunc(reflect.Indirect(responseSenderArg).Type(), responseSender)

	handlerArg := reflect.New(handlerType.In(0)).Interface()
	this.bodyFormat.Parse(message.Body, handlerArg)

	go reflect.ValueOf(handler).Call([]reflect.Value{reflect.Indirect(reflect.ValueOf(handlerArg)), responseSenderWrap})
}

//
// Handle response for previously sent request.
//
func (this *Connection) handleResponse(message parser.Response) {

	handler, ok := this.responseHandlers[message.RequestUid]

	if !ok {
		log.Println("No handlers for response with request uid ", message.RequestUid)
		return
	}

	delete(this.responseHandlers, message.RequestUid)

	handlerType := reflect.TypeOf(handler)
	handlerArg := reflect.New(handlerType.In(0)).Interface()

	this.bodyFormat.Parse(message.Body, handlerArg)

	go reflect.ValueOf(handler).Call([]reflect.Value{reflect.Indirect(reflect.ValueOf(handlerArg))})
}

/*                                                                                      */
/* ----------------------------------- Public Methods --------------------------------- */
/*                                                                                      */

//
// Drop connection instantly
//
func (this *Connection) Destroy() {
	this.conn.Close()
}

//
// Close connection gracefully sending reason
//
func (this *Connection) Close(reason string) {
	// TODO
}

// Close connection sending redirect uri
func (this *Connection) CloseRedirect(uri string) {
	// TODO
}

//
// Sends event with uri and body. Body would be serialized.
// Blocks until message is sent
//
func (this *Connection) SendEvent(uri string, body interface{}) {

	b, _ := this.bodyFormat.Serialize(body)

	event := parser.Event{
		UserHeader: parser.UserHeader{
			Uid: uuid.NewV1(),
			Uri: uri,
		},
		UserBody: parser.UserBody{
			Body: b,
		},
	}

	this.messagesOut <- &event

}

//
// Subscribe for and event. Function will call callback with the event once it receive it
//
func (this *Connection) OnEvent(uri string, f interface{}) {
	if _, ok := this.eventHandlers[uri]; !ok {
		this.eventHandlers[uri] = []CallbackFunc{}
	}
	this.eventHandlers[uri] = append(this.eventHandlers[uri], f)
}

//
// Sends request to other party. Callback with response or error would be called
// once implementation get it.
// Blocks until request is sent
//
func (this *Connection) SendRequest(uri string, body interface{}, f interface{}) {

	uid := uuid.NewV1()

	b, _ := this.bodyFormat.Serialize(body)

	request := parser.Request{
		UserHeader: parser.UserHeader{
			Uid: uid,
			Uri: uri,
		},
		Progressive: false,
		UserBody: parser.UserBody{
			Body: b,
		},
	}

	this.responseHandlers[uid] = f

	this.messagesOut <- &request

}

//
// Subscribes for request. Callback with request info would be called when request will come in.
// Callback should then call response in order to respond to a request
//
func (this *Connection) OnRequest(uri string, f CallbackFunc) {
	this.requestHandlers[uri] = f
}
