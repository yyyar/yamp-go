//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"errors"
	"github.com/satori/go.uuid"
	"github.com/yyyar/yamp-go/format"
	"github.com/yyyar/yamp-go/parser"
	"github.com/yyyar/yamp-go/transport"
	"log"
)

//
// EventHandler represents function for receiving event
//
type EventHandler func(*Event)

//
// RequestHandler represents function type for handling
// requests and writing responses
//
type RequestHandler func(*Request, *Response)

//
// ResponseHandler represents function for receiving
// response of previously sent request
//
type ResponseHandler func(*Response)

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
	// User frames body format parser/serializer
	//
	bodyFormat format.BodyFormat

	//
	// Yamp protocol parser
	//
	parser *parser.Parser

	//
	// Channel for pushing frames that will be written to other party
	//
	framesOut chan (parser.Frame)

	//
	// --------------- user callbacks storage ---------------
	//

	//
	// User callbacks for events
	//
	eventHandlers map[string][]EventHandler

	//
	// User callbacks for responses
	//
	responseHandlers map[uuid.UUID]ResponseHandler

	//
	// User callbacks for requests
	//
	requestHandlers map[string]RequestHandler
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

		framesOut: make(chan parser.Frame),

		eventHandlers:    make(map[string][]EventHandler),
		responseHandlers: make(map[uuid.UUID]ResponseHandler),
		requestHandlers:  make(map[string]RequestHandler),
	}

	go connection.loop()

	return connection
}

//
// loop is parseing / serializing loop. It works
// until it gets parser.Frames channel close event,
// i.e. while underlying reader is live
//
func (this *Connection) loop() {

	for {
		select {

		// Got new frame for sending
		case frame := <-this.framesOut:
			frame.Serialize(this.conn)

		// Got new parsed frame
		case frame, ok := <-this.parser.Frames:

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
func (this *Connection) handleEvent(event parser.Event) {

	handlers, ok := this.eventHandlers[event.Uri]

	if !ok {
		log.Println("No handlers for event uri " + event.Uri)
		return
	}

	for _, handler := range handlers {

		go handler(&Event{
			this.bodyFormat,
			event,
		})
	}

}

//
// Handles request frame
//
func (this *Connection) handleRequest(request parser.Request) {

	handler, ok := this.requestHandlers[request.Uri]

	if !ok {
		log.Println("No handlers for request", request.Uri)
		return
	}

	go handler(&Request{
		this.bodyFormat,
		request,
	}, &Response{
		this.bodyFormat,
		this.framesOut,
		&request,
		nil,
	})
}

//
// Handle response for previously sent request.
//
func (this *Connection) handleResponse(response parser.Response) {

	handler, ok := this.responseHandlers[response.RequestUid]

	if !ok {
		log.Println("No handlers for response with request uid ", response.RequestUid)
		return
	}

	delete(this.responseHandlers, response.RequestUid)

	go handler(&Response{this.bodyFormat, nil, nil, &response})
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
// Blocks until event is sent
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

	this.framesOut <- &event

}

//
// Subscribe for and event. Function will call callback with the event once it receive it
//
func (this *Connection) OnEvent(uri string, f EventHandler) {
	if _, ok := this.eventHandlers[uri]; !ok {
		this.eventHandlers[uri] = []EventHandler{}
	}
	this.eventHandlers[uri] = append(this.eventHandlers[uri], f)
}

//
// Sends request to other party. Callback with response or error would be called
// once implementation get it.
// Blocks until request is sent
//
func (this *Connection) SendRequest(uri string, body interface{}, f ResponseHandler) {

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

	this.framesOut <- &request

}

//
// Subscribes for request. Callback with request info would be called when request will come in.
// Callback should then call response in order to respond to a request
//
func (this *Connection) OnRequest(uri string, handler RequestHandler) error {

	if _, ok := this.requestHandlers[uri]; ok {
		return errors.New("Request handler on uri " + uri + " already exists")
	}

	this.requestHandlers[uri] = handler
	return nil
}
