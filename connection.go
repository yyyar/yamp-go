//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"github.com/yyyar/yamp-go/api"
	"github.com/yyyar/yamp-go/dealers"
	"github.com/yyyar/yamp-go/format"
	"github.com/yyyar/yamp-go/parser"
	"github.com/yyyar/yamp-go/transport"
	"log"
)

const (

	// Implemented Yamp Version
	YAMP_VERSION = 0x01
)

// Connection is Yamp connection abstraction supports
// events sending/handling and request/response
// processing
type Connection struct {

	// Indicates party role
	isClient bool

	// Transport connection adapter
	conn transport.Connection

	// Yamp protocol parser
	parser *parser.Parser

	// User frames body format parser/serializer
	bodyFormat format.BodyFormat

	// Channel for pushing frames that will be written to other party
	framesOut chan (parser.Frame)

	*dealers.EventDealer
	*dealers.RequestDealer
	*dealers.ResponseDealer
}

//
// NewConnection Creates new instance of Connection wrapping
// transport.Connection and immediately starting read/write loop
//
func NewConnection(isClient bool, conn transport.Connection, bodyFormat format.BodyFormat) (*Connection, error) {

	out := make(chan parser.Frame)

	connection := &Connection{

		isClient:   isClient,
		conn:       conn,
		parser:     parser.NewParser(conn),
		bodyFormat: bodyFormat,

		framesOut: out,

		EventDealer:    dealers.NewEventDealer(bodyFormat),
		RequestDealer:  dealers.NewRequestDealer(bodyFormat, out),
		ResponseDealer: dealers.NewResponseDealer(bodyFormat),
	}

	// Try handshake
	if err := connection.handshake(); err != nil {
		return nil, err
	}

	return connection, nil
}

//
// Perform initial system.handshake
//
func (c *Connection) handshake() error {

	if c.isClient {
		if err := c.handshakeClient(); err != nil {
			return err
		}
	} else {
		if err := c.handshakeServer(); err != nil {
			return err
		}
	}

	go c.readLoop()
	go c.writeLoop()

	return nil
}

//
// Initiate client-side handshake with server
//
func (c *Connection) handshakeClient() error {

	// Send system.handshake

	(&parser.SystemHandshake{
		Version: YAMP_VERSION,
	}).Serialize(c.conn)

	// Get response

	frame, ok := <-c.parser.Frames
	if !ok {
		err := <-c.parser.Error
		return err
	}

	// If got system.handshake back, then we're ok
	if frame.GetType() == parser.SYSTEM_HANDSHAKE {
		return nil
	}

	// Something bad happened
	if frame.GetType() == parser.SYSTEM_CLOSE {
		return errors.New(frame.(*parser.SystemClose).Message)
	}

	// Got unexpected message, close drop connection

	c.conn.Close()
	return errors.New("Unexpected event")
}

//
// Handle server party handshake
//
func (c *Connection) handshakeServer() error {

	// Wait for client to send system.handshake

	frame, ok := <-c.parser.Frames
	if !ok {
		err := <-c.parser.Error
		return err
	}

	// If client sent something else, close connection

	if frame.GetType() != parser.SYSTEM_HANDSHAKE {
		c.conn.Close()
		return errors.New("Unexpected frame")
	}

	// Check versions, and if we're satisfied
	// respond with the same system.handshake

	handshake := frame.(*parser.SystemHandshake)
	if handshake.Version != YAMP_VERSION {
		c.closeWithCode(parser.CLOSE_VERSION_NOT_SUPPORTED, "")
		return errors.New(fmt.Sprintf("Version not supported, client was with version %d", handshake.Version))
	}

	frame.Serialize(c.conn)

	return nil
}

//
// Serializing loop
//
func (c *Connection) writeLoop() {

	for {

		frame, ok := <-c.framesOut

		if !ok {
			return
		}

		frame.Serialize(c.conn)
	}

}

//
// Parsing loop
//
func (c *Connection) readLoop() {

	for {

		//
		// Got new parsed frame
		//
		frame, ok := <-c.parser.Frames

		if !ok {
			log.Println(<-c.parser.Error)
			return
		}

		//
		// Dispatch new frame
		//
		switch frame.GetType() {

		case parser.SYSTEM_CLOSE:

			close := frame.(*parser.SystemClose)
			log.Println(close.Code, close.Message)

		case parser.SYSTEM_PING:

			ping := frame.(*parser.SystemPing)

			// TODO:got response on our ping request
			// since we do not send it now, we don't need to
			// handle it either
			if ping.Ack {
				continue
			}

			// Respond with ping ack
			c.framesOut <- &parser.SystemPing{
				Ack:     true,
				Payload: ping.Payload,
			}

		case parser.EVENT:
			c.EventDealer.In <- *(frame).(*parser.Event)

		case parser.RESPONSE:
			c.ResponseDealer.In <- *(frame).(*parser.Response)

		case parser.REQUEST:
			c.RequestDealer.In <- *(frame).(*parser.Request)

		default:
			log.Println("Unhandled frame", frame.GetType(), frame)

		}

	}

}

func (c *Connection) closeWithCode(code parser.CloseCode, message string) {

	c.framesOut <- &parser.SystemClose{
		Code:    code,
		Message: message,
	}

	c.conn.Close()
}

//
// Drop connection and send close frame
//
func (c *Connection) Close(message string) {
	c.closeWithCode(parser.CLOSE_UNKNOWN, message)
}

//
// SendEvent
//
func (c *Connection) SendEvent(uri string, body interface{}) {

	uid := uuid.NewV1()
	b, _ := c.bodyFormat.Serialize(body)

	event := parser.Event{
		UserHeader: parser.UserHeader{
			Uid: uid,
			Uri: uri,
		},
		UserBody: parser.UserBody{
			Body: b,
		},
	}

	c.framesOut <- &event
}

//
// SendRequest
//
func (c *Connection) SendRequest(uri string, body interface{}, handler api.ResponseHandler) {

	uid := uuid.NewV1()
	b, _ := c.bodyFormat.Serialize(body)

	request := parser.Request{
		UserHeader: parser.UserHeader{
			Uid: uid,
			Uri: uri,
		},
		UserBody: parser.UserBody{
			Body: b,
		},
	}

	c.OnResponse(uid, handler)

	c.framesOut <- &request
}
