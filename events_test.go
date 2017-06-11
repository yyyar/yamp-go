//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package yamp

import (
	"github.com/yyyar/yamp-go/api"
	"github.com/yyyar/yamp-go/format"
	"io"
	"log"
	"sync"
	"testing"
)

//
// Test basic pub/sub events operation
//
func TestEvents(t *testing.T) {

	var wg sync.WaitGroup
	wg.Add(2)

	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	client := NewConnection(&MockConnection{r1, w2}, &format.JsonBodyFormat{})
	server := NewConnection(&MockConnection{r2, w1}, &format.JsonBodyFormat{})

	client.OnEvent("foo", func(event *api.Event) {

		var body struct{ Msg string }
		event.ReadTo(&body)

		log.Println("OnEvent 'foo'", body)
		wg.Done()
	})

	log.Println("SendEvent 'foo'")
	server.SendEvent("foo", struct{ Msg string }{"Hello"})
	server.SendEvent("foo", struct{ Msg string }{"World"})

	wg.Wait()
}
