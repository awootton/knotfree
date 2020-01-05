// Copyright 2019,2020 Alan Tracey Wootton
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package str2protocol

import (
	"errors"
	"knotfreeiot/iot"
	"knotfreeiot/iot/reporting"
)

// ServerOfStr2 - implement string messages
func ServerOfStr2(subscribeMgr iot.PubsubIntf, addr string) *iot.SockStructConfig {

	config := iot.NewSockStructConfig(subscribeMgr)

	ServerOfStr2Init(config)

	iot.ServeFactory(config, addr)

	return config
}

// ServerOfStr2Init is to set default callbacks.
func ServerOfStr2Init(config *iot.SockStructConfig) {

	config.SetCallback(str2ServeCallback)

	servererr := func(ss *iot.SockStruct, err error) {
		str2LogThing.Collect("server closing")
	}
	config.SetClosecb(servererr)

	//  the writer
	handleTopicPayload := func(ss *iot.SockStruct, topic []byte, payload []byte, returnAddress []byte) error {

		cmd := Send{}
		cmd.source = returnAddress
		cmd.destination = topic
		cmd.data = payload

		err := cmd.Write(ss.GetConn())
		if err != nil {
			str2LogThing.Collect("error in str2 writer") //, n, err, cmd)
			return err
		}
		return nil
	}

	config.SetWriter(handleTopicPayload)
}

// str2ServeCallback is the default callback which implements an api
// to the pub sub manager.
//
func str2ServeCallback(ss *iot.SockStruct) {

	for {
		packet, err := ReadPacket(ss.GetConn())
		if err != nil {
			dis := Disconnect{}
			dis.options["error"] = Bstr(err.Error())
			err2 := dis.Write(ss.GetConn())
			_ = err2
			ss.Close(err)
			return
		}
		// As much fun as it would be to make the following code into virtual methods
		// of the types involved (and I tried it) it's more annoying and harder to read
		// than just doing it all here.
		switch packet.(type) {

		case *Subscribe:
			p := packet.(*Subscribe)
			ss.SendSubscriptionMessage(p.destination)

		case *Unsubscribe:
			p := packet.(*Unsubscribe)
			ss.SendSubscriptionMessage(p.destination)

		case *Connect:
			p := packet.(*Connect)
			// TODO copy out the JWT
			_ = p

		case *Disconnect:
			p := packet.(*Disconnect)
			err := errors.New("exit") // TODO copy over options into json?
			ss.Close(err)
			_ = p

		default:
			dis := Disconnect{}
			dis.options["error"] = Bstr("error unknown command")
			err2 := dis.Write(ss.GetConn())
			_ = err2
		}

		// 	switch first {
		// 	case "exit":
		// 		ServerOfStringsWrite(ss, "exit")
		// 		err := errors.New("exit")
		// 		ss.Close(err)

		// 	case "sub":
		// 		topic := strings.Trim(remaining, " ")
		// 		if len(topic) <= 0 {
		// 			ServerOfStringsWrite(ss, "error say 'sub mytopic' and not "+text)
		// 		} else {
		// 			ss.SendSubscriptionMessage([]byte(topic))
		// 			ServerOfStringsWrite(ss, "ok sub "+topic)
		// 		}

		// 	case "add":
		// 		returnAddr := strings.Trim(remaining, " ")
		// 		if len(returnAddr) <= 0 {
		// 			ServerOfStringsWrite(ss, "error say 'add returnAddr' and not "+text)
		// 		} else {
		// 			ss.SetSelfAddress([]byte(returnAddr))
		// 			ss.SendSubscriptionMessage([]byte(returnAddr))
		// 			ServerOfStringsWrite(ss, "ok add "+returnAddr)
		// 		}

		// 	case "unsub":
		// 		topic := strings.Trim(remaining, " ")
		// 		if len(topic) <= 0 {
		// 			ServerOfStringsWrite(ss, "error say 'unsub mytopic' and not "+text)
		// 		} else {
		// 			ss.SendUnsubscribeMessage([]byte(topic))
		// 			ServerOfStringsWrite(ss, "ok unsub "+topic)
		// 		}

		// 	case "pub":
		// 		topic, payload := GetFirstWord(remaining)
		// 		if len(topic) <= 0 || len(payload) < 0 {
		// 			ServerOfStringsWrite(ss, "error say 'pub mytopic mymessage' and not "+text)
		// 		} else {
		// 			topicHash := iot.HashType{}
		// 			topicHash.FromString(topic)
		// 			bytes := []byte(payload)
		// 			ss.SendPublishMessage([]byte(topic), []byte(bytes), []byte("unknown")) // FIXME:
		// 			ServerOfStringsWrite(ss, "ok pub "+topic+" "+payload)
		// 		}

		// 	default:
		// 		ServerOfStringsWrite(ss, "error unknown command "+text)
		// 	}
	}
}

var str2LogThing = reporting.NewStringEventAccumulator(16)
