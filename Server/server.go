package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/tangnguyendeveloper/go-diameter/diam"
	"github.com/tangnguyendeveloper/go-diameter/diam/datatype"
)

const port = 3868

type Connection struct {
	client  net.Conn
	message diam.Message
}

func main() {

	my_listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	log.Println("listening at ", my_listener.Addr().String())

	queue_receive := make(chan net.Conn, 5)
	queue_response := make(chan Connection, 5)

	go func() {
		for {
			conn, err := my_listener.Accept()
			if err != nil {
				log.Println(err)
				continue
			}

			queue_receive <- conn
			time.Sleep(time.Millisecond * 5)
		}
	}()

	defer close(queue_receive)
	defer close(queue_response)
	defer my_listener.Close()

	for {
		select {

		case conn, ok := <-queue_receive:
			if !ok {
				continue
			} else {
				queue_response <- ReceiveRequest(conn)
			}
		case connection, ok := <-queue_response:
			if !ok {
				continue
			} else {
				header := connection.message.GetHeader()
				header.CommandFlags = 0
				connection.message.SetHeader(header)
				connection.message.RemoveAVPWithCode(269)

				nb, err := connection.client.Write(
					connection.message.ToBytes(),
				)
				if err != nil {
					log.Println(err)
					continue
				}
				log.Println("The response are sent in ", nb, " bytes")
			}
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}

}

func ReceiveRequest(client net.Conn) Connection {

	before := make([]byte, 4)
	nb, err := client.Read(before)
	if err != nil {
		log.Fatal(err)
	}
	length_mess := diam.Uint24to32(before[1:nb])

	message_request := diam.NewMessage()

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

		message_request.FromBytes(response_buffer.Bytes())

		js_str, _ := message_request.ToJsonString()
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

		for _, avp := range message_request.GetListAVP() {
			log.Println("AVP code: ", avp.Code, ", Attribute value: ", DecodeData(&avp.Data, avp.Code))
		}
	}
	return Connection{
		client: client, message: *message_request,
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
