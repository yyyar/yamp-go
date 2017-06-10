//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

//
// UserBody frame part
//
type UserBody struct {
	Body []byte
}

func ParseUserBody(buffer io.Reader) (*UserBody, error) {

	message := UserBody{}

	// size of Body
	var size uint32
	if err := utils.Parse(buffer, &size); err != nil {
		return nil, err
	}

	// Body
	message.Body = make([]byte, size)
	if err := utils.Parse(buffer, &message.Body); err != nil {
		return nil, err
	}

	return &message, nil
}

func WriteUserBody(writer io.Writer, message UserBody) error {

	utils.Serialize(writer, uint32(len(message.Body)))
	utils.Serialize(writer, message.Body)

	return nil
}
