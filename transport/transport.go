//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package transport

import (
	"io"
)

//
// Connection is generic transport connection interface
//
type Connection interface {
	io.ReadWriteCloser
}
