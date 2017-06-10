//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"github.com/yyyar/yamp-go/format"
	"io"
	"log"
	"net"
	"sync"
	"testing"
)

//
// MockConnection useful for testing with io.Pipe()
//
type MockConnection struct {
	io.ReadCloser
	io.Writer
}

//
// Test basic request-response operation
//
func TestReqRes(t *testing.T) {

	var wg sync.WaitGroup
	wg.Add(3)

	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	// Run responder
	go (func() {

		client := NewConnection(&MockConnection{r1, w2}, &format.JsonBodyFormat{})

		client.OnRequest("sum", func(body []int, respond func(int)) {
			log.Println("OnRequest 'sum': ", body)
			respond(body[0] + body[1])
		})

	})()

	// Run requester
	go (func() {

		server := NewConnection(&MockConnection{r2, w1}, &format.JsonBodyFormat{})

		log.Println("SendRequest 'sum' [1 2]")
		server.SendRequest("sum", []int{1, 2}, func(body int) {
			log.Println("OnResponse 'sum' [1 2]: ", body)
			wg.Done()
		})

		log.Println("SendRequest 'sum' [3 4]")
		server.SendRequest("sum", []int{3, 4}, func(body int) {
			log.Println("OnResponse 'sum' [3 4]: ", body)
			wg.Done()
		})

		log.Println("SendRequest 'sum' [5 6]")
		server.SendRequest("sum", []int{5, 6}, func(body int) {
			log.Println("OnResponse 'sum' [5 6]: ", body)
			wg.Done()
		})
	})()

	wg.Wait()
}

//
// Test basic pub/sub events operation
//
func TestEvents(t *testing.T) {

	var wg sync.WaitGroup
	wg.Add(1)

	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	client := NewConnection(&MockConnection{r1, w2}, &format.JsonBodyFormat{})
	server := NewConnection(&MockConnection{r2, w1}, &format.JsonBodyFormat{})

	client.OnEvent("foo", func(body struct{ Msg string }) {

		log.Println("OnEvent 'foo'", body)
		wg.Done()
	})

	log.Println("SendEvent 'foo'")
	server.SendEvent("foo", struct{ Msg string }{"Hello"})

	wg.Wait()
}

//
// Test tcp transport
//
func TestTcp(t *testing.T) {

	const (
		PROTOCOL = "tcp"
		ADDR     = "localhost:5555"
	)

	var wg sync.WaitGroup
	wg.Add(1)

	l, err := net.Listen(PROTOCOL, ADDR)
	if err != nil {
		log.Println(err)
		return
	}
	defer l.Close()

	go (func() {
		c, _ := l.Accept()
		conn := NewConnection(c, &format.JsonBodyFormat{})
		conn.OnRequest("echo", func(body string, respond func(string)) {
			log.Println("OnRequest 'echo': ", body)
			respond(body)
		})
	})()

	c, err := net.Dial(PROTOCOL, ADDR)
	conn := NewConnection(c, &format.JsonBodyFormat{})
	conn.SendRequest("echo", "hello world!", func(body string) {
		log.Println("OnResponse 'echo' \"hello world\": ", body)
		wg.Done()
	})

	wg.Wait()
}
