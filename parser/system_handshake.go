//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package parser

import (
	"github.com/yyyar/yamp-go/utils"
	"io"
)

const SYSTEM_HANDSHAKE FrameType = 0x00

//
// SystemHandshake frame
//
type SystemHandshake struct {
	Version uint16
}

func (this *SystemHandshake) GetType() FrameType {
	return SYSTEM_HANDSHAKE
}

func (this *SystemHandshake) Parse(buffer io.Reader) error {

	// Version
	if err := utils.Parse(buffer, &this.Version); err != nil {
		return err
	}

	return nil
}

func (this *SystemHandshake) Serialize(writer io.Writer) error {

	utils.Serialize(writer, this.GetType())

	utils.Serialize(writer, this.Version)

	return nil
}
