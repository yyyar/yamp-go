//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package utils

import (
	"encoding/binary"
	"io"
)

var ENDIANNESS binary.ByteOrder = binary.BigEndian

//
// Parse parses data from reader enough to fill object
//
func Parse(reader io.Reader, to interface{}) error {
	err := binary.Read(reader, ENDIANNESS, to)
	return err
}

//
// Serialize serializes data to writer
//
func Serialize(writer io.Writer, data interface{}) error {
	err := binary.Write(writer, ENDIANNESS, data)
	return err
}
