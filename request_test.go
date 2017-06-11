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

	const N = 10

	var wg sync.WaitGroup
	wg.Add(N)

	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	// Run responder
	go (func() {

		client := NewConnection(&MockConnection{r1, w2}, &format.JsonBodyFormat{})

		client.OnRequest("sum", func(req *api.Request, res *api.Response) {

			var body []int
			//req.ReadTo(&body)
			(&format.JsonBodyFormat{}).Parse(req.RawBody(), &body)

			log.Println(res.RequestId(), "OnRequest   'sum': ", body)

			res.Done(body[0] + body[1])
		})

	})()

	// Run requester
	go (func() {

		server := NewConnection(&MockConnection{r2, w1}, &format.JsonBodyFormat{})

		for i := 0; i < N; i++ {

			server.SendRequest("sum", []int{i, i}, func(res *api.Response) {

				var body int
				res.ReadTo(&body)

				log.Println(res.RequestId(), "OnResponse  'sum': ", body)
				wg.Done()
			})

		}

	})()

	wg.Wait()
}
