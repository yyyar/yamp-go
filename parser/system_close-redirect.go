//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

const SYSTEM_CLOSE_REDIRECT FrameType = 0x04

//
// SystemCloseRedirect frame
//
type SystemCloseRedirect struct {
	Url string
}

func (this *SystemCloseRedirect) GetType() FrameType {
	return SYSTEM_CLOSE_REDIRECT
}

func (this *SystemCloseRedirect) Parse(buffer io.Reader) error {

	// size of Url
	var size uint16
	if err := utils.Parse(buffer, &size); err != nil {
		return nil
	}

	// Url
	url := make([]uint8, size)
	if err := utils.Parse(buffer, &url); err != nil {
		return nil
	}
	this.Url = string(url[:])

	return nil
}

func (this *SystemCloseRedirect) Serialize(writer io.Writer) error {

	utils.Serialize(writer, this.GetType())

	utils.Serialize(writer, uint16(len(this.Url)))
	utils.Serialize(writer, []byte(this.Url))

	return nil
}
