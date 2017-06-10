//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

const SYSTEM_CLOSE FrameType = 0x03

//
// SystemClose frame
//
type SystemClose struct {
	Reason string
}

func (this *SystemClose) GetType() FrameType {
	return SYSTEM_CLOSE
}

func (this *SystemClose) Parse(buffer io.Reader) error {

	// size of Reason
	var size uint16
	if err := utils.Parse(buffer, &size); err != nil {
		return err
	}

	// Reason
	reason := make([]uint8, size)
	if err := utils.Parse(buffer, &reason); err != nil {
		return err
	}
	this.Reason = string(reason[:])

	return nil
}

func (this *SystemClose) Serialize(writer io.Writer) error {

	utils.Serialize(writer, this.GetType())

	utils.Serialize(writer, uint16(len(this.Reason)))
	utils.Serialize(writer, []byte(this.Reason))

	return nil
}
