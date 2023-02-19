package diam

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

// Header is the header representation of a Diameter message.
type Header struct {
	Version       uint8
	MessageLength uint32
	CommandFlags  uint8
	CommandCode   uint32
	ApplicationID uint32
	HopByHopID    uint32
	EndToEndID    uint32
}

// HeaderLength is the length of a Diameter header data structure.
const HeaderLength = 20

// Command flags.
const (
	RequestFlag       = 1 << 7
	ProxiableFlag     = 1 << 6
	ErrorFlag         = 1 << 5
	RetransmittedFlag = 1 << 4
)

func NewHeader() *Header {
	return &Header{}
}

func (header *Header) FromBytes(bytes_array []byte) error {

	if n := len(bytes_array); n < HeaderLength {
		return fmt.Errorf("can't load the header with %d bytes, the header length must be %d bytes", n, HeaderLength)
	}

	header.Version = bytes_array[0]
	header.MessageLength = Uint24to32(bytes_array[1:4])
	header.CommandFlags = bytes_array[4]
	header.CommandCode = Uint24to32(bytes_array[5:8])
	header.ApplicationID = binary.BigEndian.Uint32(bytes_array[8:12])
	header.HopByHopID = binary.BigEndian.Uint32(bytes_array[12:16])
	header.EndToEndID = binary.BigEndian.Uint32(bytes_array[16:20])

	return nil
}

func (header Header) ToBytes() []byte {
	buffer := make([]byte, HeaderLength)

	buffer[0] = header.Version
	copy(buffer[1:4], Uint32to24(header.MessageLength))
	buffer[4] = header.CommandFlags
	copy(buffer[5:8], Uint32to24(header.CommandCode))
	binary.BigEndian.PutUint32(buffer[8:12], header.ApplicationID)
	binary.BigEndian.PutUint32(buffer[12:16], header.HopByHopID)
	binary.BigEndian.PutUint32(buffer[16:20], header.EndToEndID)

	return buffer
}

func (header Header) ToJsonString() (string, error) {
	jsonBytes, err := json.MarshalIndent(header, "", "    ")
	return string(jsonBytes), err
}

// uint24to32 converts b from []byte in network byte order to uint32.
func Uint24to32(b []byte) uint32 {
	if len(b) != 3 {
		return 0
	}
	return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2])
}

// uint32to24 converts b from uint32 to []byte in network byte order.
func Uint32to24(n uint32) []byte {
	return []byte{uint8(n >> 16), uint8(n >> 8), uint8(n)}
}
