package diam

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

// AVP Flags. See section 4.1 of RFC 6733.
const (
	Pbit = 1 << 5 // The 'P' bit, reserved for future use.
	Mbit = 1 << 6 // The 'M' bit, known as the Mandatory bit.
	Vbit = 1 << 7 // The 'V' bit, known as the Vendor-Specific bit.
)

// AVP is a Diameter attribute-value-pair.
type AVP struct {
	Code     uint32 // Code of this AVP
	Flags    uint8  // Flags of this AVP
	Length   int    // Length of this AVP's payload
	VendorID uint32 // VendorId of this AVP
	Data     []byte // Data of this AVP (payload)
}

func NewAVP() *AVP {
	return &AVP{}
}

func (avp *AVP) FromBytes(bytes_array []byte) error {
	var headerlen int = 8

	if n := len(bytes_array); n < headerlen {
		return fmt.Errorf("can't load the avp with %d bytes, the avp length must be greater than %d bytes", n, headerlen)
	}

	avp.Code = binary.BigEndian.Uint32(bytes_array[0:4])
	avp.Flags = bytes_array[4]
	avp.Length = int(Uint24to32(bytes_array[5:8]))

	if Vbit&avp.Flags == Vbit {

		headerlen = 12
		if n := len(bytes_array); n < headerlen {
			return fmt.Errorf("can't load the avp with %d bytes, the avp length must be greater than %d bytes because the vendor-specific bit have set", n, headerlen)
		} else {
			avp.VendorID = binary.BigEndian.Uint32(bytes_array[8:12])
			avp.Data = bytes_array[12:avp.Length]
		}

	} else {
		avp.Data = bytes_array[8:avp.Length]
	}

	bodyLen := avp.Length - headerlen
	if n := len(avp.Data); n < bodyLen {
		return fmt.Errorf(
			"not enough data to decode AVP: %d != %d",
			headerlen, n,
		)
	}

	return nil
}

func (avp AVP) ToBytes() ([]byte, error) {

	err := avp.ComputeLength()
	if err != nil {
		return []byte{}, err
	}

	buffer_size := avp.LengthWithPadding()

	buffer := make([]byte, buffer_size)

	binary.BigEndian.PutUint32(buffer[0:4], avp.Code)
	buffer[4] = avp.Flags
	copy(buffer[5:8], Uint32to24(uint32(avp.Length)))

	if avp.Flags&Vbit == Vbit {
		binary.BigEndian.PutUint32(buffer[8:12], avp.VendorID)
	}

	copy(buffer[avp.headerLength():avp.Length], avp.Data)

	for index := avp.Length; index < buffer_size; index++ {
		buffer[index] = 0
	}

	return buffer, nil
}

func (avp AVP) ToJsonString() (string, error) {
	jsonBytes, err := json.MarshalIndent(avp, "", "    ")
	return string(jsonBytes), err
}

func (avp AVP) headerLength() int {
	if avp.Flags&Vbit == Vbit {
		return 12
	}
	return 8
}

func (avp *AVP) ComputeLength() error {
	if len(avp.Data) == 0 {
		return fmt.Errorf("failed to serialize AVP: Data is nil")
	}
	avp.Length = avp.headerLength() + len(avp.Data)
	return nil
}

func (avp *AVP) LengthWithPadding() int {
	avp.ComputeLength()
	len_mod_4 := avp.Length % 4
	padding := 4 - len_mod_4

	if len_mod_4 == 0 {
		return avp.Length
	}
	return avp.Length + padding
}
