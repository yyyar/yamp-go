//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

//
// UserHeader frame part
//
type UserHeader struct {
	Uid [16]byte
	Uri string
}

func ParseUserHeader(buffer io.Reader) (*UserHeader, error) {

	message := UserHeader{}

	// Uid
	if err := utils.Parse(buffer, &message.Uid); err != nil {
		return nil, err
	}

	// size of Uri
	var size uint8
	if err := utils.Parse(buffer, &size); err != nil {
		return nil, err
	}

	// Uri
	uri := make([]uint8, size)
	if err := utils.Parse(buffer, &uri); err != nil {
		return nil, err
	}
	message.Uri = string(uri[:])

	return &message, nil
}

func WriteUserHeader(writer io.Writer, message UserHeader) error {

	utils.Serialize(writer, message.Uid)
	utils.Serialize(writer, uint8(len(message.Uri)))
	utils.Serialize(writer, []byte(message.Uri))

	return nil
}
