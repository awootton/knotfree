package protocolaa

import (
	"errors"
	"fmt"
	"knotfree/types"
	"net"
	"reflect"
	"strings"
	"time"
)

// ProtocolAa is a lame pub/sub protocol with a length byte followed by a string of len 0 to 255.
// The first char in the string
// indicates which type of message we're getting so a zero string is going to be an error.

// Push into 'west' chan headed for the east. Used by clients and not the server
func (me *Handler) Push(cmd interface{}) error {
	tmp := cmd
	aathing, ok := tmp.(aaInterface)
	if !ok {
		return errors.New("expected aaInterface{} got " + reflect.TypeOf(cmd).String())
	}
	select {
	case me.wire.west <- aathing:
	case <-time.After(10 * time.Millisecond):
		return errors.New("Aa Push slow")
	}
	return nil
}

// Pop blocks. Actually returns aaInterface. See above.
func (me *Handler) Pop() (interface{}, error) {
	select {
	case obj := <-me.wire.east:
		return obj, nil
	case <-time.After(21 * time.Minute):
		return nil, errors.New("Aa read too slow")
	}
	//return nil, errors.New("wtf")
}

type aaDuplexChannel struct {
	east chan aaInterface
	west chan aaInterface
}

var aaDefaultTimeout = 21 * time.Minute

func newAaDuplexChannel(capacity int, conn *net.TCPConn) aaDuplexChannel {
	adc := aaDuplexChannel{}
	adc.east = make(chan aaInterface, capacity)
	adc.west = make(chan aaInterface, capacity)
	// We'll put the socket in the east.
	go func() {
		for {
			str, err := readProtocolAstr(conn)
			//fmt.Println("sock to east:", str, err)
			if err != nil { // we're dead. its over.
				adc.east <- &Death{"Aa read err " + err.Error()}
				return
			}
			obj := unMarshalAa(str[:1], str[1:])
			adc.east <- obj
		}
	}()

	go func() {
		for {
			obj := <-adc.west
			str := obj.marshal()
			err := writeProtocolAaStr(conn, str)
			if err != nil { // we're dead. its over.
				death := Death{"Aa write err " + err.Error()}
				str = death.marshal()
				_ = writeProtocolAaStr(conn, str)
			}
		}
	}()

	return adc
}

func unMarshalAa(firstChar string, str string) aaInterface {
	switch firstChar[0] {
	case 's':
		return &Subscribe{str}
	case 't':
		return &SetTopic{str}
	case 'p':
		return &Publish{str}
	case 'd':
		return &Death{str}
	}
	return &Ping{}
}

// HandleWrite from ProtocolHandler interface
func (me *ServerHandler) HandleWrite(msg *types.IncomingMessage) error {

	realName, ok := me.c.GetRealTopicName(msg.Topic)
	if !ok {
		return errors.New("missing real name")
	}
	// TODO: optimize redundant SetTopic commands.
	select {
	case me.wire.west <- &SetTopic{realName}:
	case <-time.After(10 * time.Millisecond):
		return errors.New("Aa wr slow")
	}
	select {
	case me.wire.west <- &Publish{string(*msg.Message)}:
	case <-time.After(10 * time.Millisecond):
		return errors.New("Aa wr slow2")
	}
	return nil
}

// Serve implementing  ProtocolHandler interface
func (me *ServerHandler) Serve() error {
	select {
	case obj := <-me.wire.east:
		err := obj.execute(me)
		if err != nil {
			return err
		}
	case <-time.After(21 * time.Minute):
		return errors.New("Aa read slow")
	}
	return nil
}

// readProtocolAstr will block trying to get a string until the conn times out.
func readProtocolAstr(conn net.Conn) (string, error) {

	buffer := make([]byte, 256)
	ch := []byte{'a'}
	n, err := conn.Read(ch)
	if n != 1 {
		if err != nil { // probably timed out
			fmt.Println(err.Error())
			return string(buffer), err
		}
		return string(buffer), errors.New(" needed 1 bytes. got " + string(buffer))
	}
	msglen := int(ch[0]) & 0x00FF
	var sb strings.Builder
	for msglen > 0 {
		n, err = conn.Read(buffer[:msglen])
		s := string(buffer[:n])
		sb.WriteString(s)
		msglen -= n
		if err != nil {
			return "", err
		}
	}
	return sb.String(), nil
}

// writeProtocolAaStr writes our lame protocol to the conn
func writeProtocolAaStr(conn net.Conn, str string) error {

	strbytes := []byte(str)
	if len(strbytes) > 255 {
		return errors.New("WriteProtocolA string too long")
	}
	prefix := []byte{byte(len(strbytes))}
	n, err := conn.Write(prefix)
	if n != 1 || err != nil {
		return err
	}
	n, err = conn.Write(strbytes)
	if n != len(strbytes) || err != nil {
		return err
	}
	return nil
}