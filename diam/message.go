package diam

import (
	"bytes"
	"fmt"
	"log"
)

type Message struct {
	header   *Header
	list_avp []*AVP
}

func NewMessage() *Message {
	return &Message{}
}

func (mess Message) GetHeader() Header {
	return *mess.header
}

func (mess *Message) SetHeader(header Header) {
	if mess.header == nil {
		mess.header = NewHeader()
	}
	*mess.header = header
}

func (mess Message) GetListAVP() []*AVP {
	result := make([]*AVP, len(mess.list_avp))
	CopyListAVP(result, mess.list_avp)
	return result
}

func (mess Message) GetAVPWithCode(code uint32) AVP {
	for _, avp := range mess.list_avp {
		if avp == nil {
			continue
		}
		if code == avp.Code {
			return *avp
		}
	}
	return *NewAVP()
}

func (mess *Message) SetListAVP(avps []*AVP) {
	mess.list_avp = make([]*AVP, len(avps))
	CopyListAVP(mess.list_avp, avps)
}

func (mess *Message) AddAVP(avp *AVP) {
	addElement(&mess.list_avp, avp)
}

func (mess *Message) RemoveAVPWithCode(code uint32) {
	mess.list_avp = removeElement(mess.list_avp, code)
}

func (mess Message) ToBytes() []byte {
	var mess_buff bytes.Buffer

	mess.ComputeMessageLength()

	n, err := mess_buff.Write(mess.header.ToBytes())
	if err != nil {
		log.Println(err)
	} else if n != HeaderLength {
		log.Println("Wanning: only ", n, "bytes dumped")
	}

	for _, avp := range mess.list_avp {
		if avp != nil {
			avp_buf, err := avp.ToBytes()
			if err != nil {
				log.Println(err)
			}

			_, err = mess_buff.Write(avp_buf)
			if err != nil {
				log.Println(err)
			}
		}
	}

	if mess.header.MessageLength != uint32(mess_buff.Len()) {
		log.Println("Error: Message dumped false!", mess.header.MessageLength, " != ", mess_buff.Len())
	}
	return mess_buff.Bytes()
}

func (mess *Message) ComputeMessageLength() {
	mess.header.MessageLength = HeaderLength

	for _, avp := range mess.list_avp {
		if avp != nil {
			mess.header.MessageLength += uint32(avp.LengthWithPadding())
		}
	}
}

func (mess *Message) FromBytes(bytes_array []byte) error {
	b_arr_len, n := len(bytes_array), (HeaderLength + 8)

	if b_arr_len < n {
		return fmt.Errorf("can't load the message with %d bytes, the message length must be %d bytes", b_arr_len, n)
	}

	if mess.header == nil {
		mess.header = NewHeader()
	}
	err := mess.header.FromBytes(bytes_array[:HeaderLength])
	if err != nil {
		return err
	}

	index := HeaderLength
	for (index + 5) < b_arr_len {
		length, start, end := existAVP(bytes_array[index:index+8], index, b_arr_len)
		if length == 0 {
			return fmt.Errorf("can't load AVP of the message")
		}

		avp := NewAVP()
		err = avp.FromBytes(bytes_array[start:end])
		if err != nil {
			return err
		}
		mess.AddAVP(avp)

		index = end
	}

	return nil
}

func (mess Message) ToJsonString() (string, error) {

	if mess.header == nil {
		return "", fmt.Errorf("no data in the message")
	}

	result, err := mess.header.ToJsonString()
	if err != nil {
		return result, err
	}
	results := "{\n \"Diameter_Header\": \n"
	results += result + ",\n \"Attribute-Value_Pairs\": [\n"

	for _, avp := range mess.list_avp {
		if avp == nil {
			continue
		}
		result, err = avp.ToJsonString()
		if err != nil {
			return result, err
		}
		results += result + ",\n"
	}

	results += "]}"

	return results, err
}

func removeElement(list_avp []*AVP, avp uint32) []*AVP {
	for i, a := range list_avp {
		if a == nil {
			continue
		}
		if a.Code == avp {
			return append(list_avp[:i], list_avp[i+1:]...)
		}
	}
	return list_avp
}

func addElement(list_avp *[]*AVP, avp *AVP) {
	item := NewAVP()
	*item = *avp
	*list_avp = append(*list_avp, item)
}

func CopyListAVP(dst []*AVP, src []*AVP) int {
	n := len(dst)
	for i := 0; i < n; i++ {
		if dst[i] == nil {
			dst[i] = NewAVP()
		}
		*dst[i] = *src[i]
	}
	return n
}

func existAVP(bytes_array []byte, index int, len_arr int) (length int, start int, end int) {
	leng := int(Uint24to32(bytes_array[5:8]))
	len_mod_4 := leng % 4
	padding := 4 - len_mod_4

	if len_mod_4 == 0 {
		padding = 0
	}
	end = index + leng + padding
	if end > len_arr {
		return 0, 0, 0
	}

	return leng, index, end
}
