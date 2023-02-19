package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/tangnguyendeveloper/go-diameter/diam"
	"github.com/tangnguyendeveloper/go-diameter/diam/datatype"
)

const host = "127.0.0.1"
const port = 3868

func main() {

	client := SendMessage()

	defer client.Close()
	ReceiveResponse(client)

	fmt.Println("Done!")

}

func SendMessage() net.Conn {

	client, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to ", host, "port ", port)

	message := CreateMessage().ToBytes()
	nb, err := client.Write(message)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Sent a message in ", nb, "bytes")

	return client
}

func CreateMessage() *diam.Message {
	header := diam.NewHeader()

	header.Version = 1
	header.CommandFlags = diam.RequestFlag
	header.CommandCode = 275
	header.ApplicationID = 16777216
	header.HopByHopID = 10
	header.EndToEndID = 20

	avp1 := diam.NewAVP()
	avp1.Flags = 0
	avp1.Code = 269
	avp1.Data = datatype.UTF8String("Nokia 1280 Chào Các Bạn 123456").Serialize()

	var list_avp [2]*diam.AVP

	list_avp[0] = diam.NewAVP()
	list_avp[0].Flags = diam.Mbit
	list_avp[0].Code = 483
	list_avp[0].Data = datatype.Enumerated(8888).Serialize()
	list_avp[1] = diam.NewAVP()
	list_avp[1].Flags = diam.Vbit
	list_avp[1].Code = 258
	list_avp[1].VendorID = 10415
	list_avp[1].Data = datatype.Unsigned32(999).Serialize()

	mess := diam.NewMessage()
	mess.SetHeader(*header)
	mess.SetListAVP(list_avp[:])
	mess.AddAVP(avp1)

	return mess
}

func ReceiveResponse(client net.Conn) {

	before := make([]byte, 4)
	nb, err := client.Read(before)
	if err != nil {
		log.Fatal(err)
	}
	length_mess := diam.Uint24to32(before[1:nb])

	if length_mess > uint32(nb) {
		after := make([]byte, length_mess-uint32(nb))
		_, err = client.Read(after)
		if err != nil {
			log.Fatal(err)
		}

		var response_buffer bytes.Buffer
		_, err = response_buffer.Write(before)
		if err != nil {
			log.Fatal(err)
		}
		_, err = response_buffer.Write(after)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Received message in ", response_buffer.Len(), " bytes")

		message_response := diam.NewMessage()
		message_response.FromBytes(response_buffer.Bytes())

		js_str, _ := message_response.ToJsonString()
		log.Printf("\n%s", js_str)

		file, err := os.Create("message.json")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		_, err = file.Write([]byte(js_str))
		if err != nil {
			log.Fatal(err)
		}

		for _, avp := range message_response.GetListAVP() {
			log.Println("AVP code: ", avp.Code, ", Attribute value: ", DecodeData(&avp.Data, avp.Code))
		}
	}
}

func DecodeData(data *[]byte, code uint32) datatype.Type {
	switch code {
	case 269:
		out, _ := datatype.DecodeUTF8String(*data)
		return out
	case 483:
		out, _ := datatype.DecodeEnumerated(*data)
		return out
	case 258:
		out, _ := datatype.DecodeUnsigned32(*data)
		return out
	default:
		return datatype.Unknown{}
	}
}
