//
// Copyright 2017 Yaroslav Pogrebnyak <yyyaroslav@gmail.com>
//

package format

//
// BodyFormat is interface for yamp body format
//
type BodyFormat interface {

	//
	// Serialize object to byte array
	//
	Serialize(interface{}) ([]byte, error)

	//
	// Parse byte array to object
	//
	Parse([]byte, interface{}) error

	//
	// Get format type as string
	//
	GetType() string
}
