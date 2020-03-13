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

package iot_test

import (
	"fmt"
	"testing"

	"github.com/awootton/knotfreeiot/iot"
	"github.com/awootton/knotfreeiot/packets"
	"github.com/awootton/knotfreeiot/tokens"
)

func TestTwoLevel(t *testing.T) {

	tokens.LoadPublicKeys()

	got := ""
	want := ""
	ok := true
	var err error
	currentTime = starttime

	// set up
	guru0 := iot.NewExecutive(100, "guru0", getTime, true)
	iot.GuruNameToConfigMap["guru0"] = guru0

	aide1 := iot.NewExecutive(100, "aide1", getTime, false)
	aide2 := iot.NewExecutive(100, "aide2", getTime, false)

	aide1.Looker.NameResolver = testNameResolver
	aide2.Looker.NameResolver = testNameResolver
	// we have to tell aides to connect to guru
	names := []string{"guru0"}
	aide1.Looker.SetUpstreamNames(names, names)
	aide2.Looker.SetUpstreamNames(names, names)
	WaitForActions(guru0)
	WaitForActions(aide1)
	WaitForActions(aide2)
	WaitForActions(guru0)
	// make a contact
	contact1 := MakeTestContact(aide1.Config) //testContact{}
	//contact1.downMessages = make(chan packets.Interface, 1000)
	//iot.AddContactStruct(&contact1.ContactStruct, &contact1, aide1.Config)
	// another
	contact2 := MakeTestContact(aide2.Config) //testContact{}
	//contact2.downMessages = make(chan packets.Interface, 1000)
	//iot.AddContactStruct(&contact2.ContactStruct, &contact2, aide2.Config)
	// note that they are in *different* lookups so normally they could not communicate but here we have a guru.

	connect := packets.Connect{}
	connect.SetOption("token", []byte(tokens.SampleSmallToken))
	iot.Push(contact1, &connect)
	iot.Push(contact2, &connect)

	// subscribe
	subs := packets.Subscribe{}
	subs.Address = []byte("contact1 address")
	err = iot.Push(contact1, &subs)

	WaitForActions(guru0)
	WaitForActions(aide1)
	WaitForActions(aide2)
	WaitForActions(guru0)

	got = contact1.(*testContact).getResultAsString()
	want = "no message received"
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	val := readCounter(iot.TopicsAdded)
	got = fmt.Sprint("topics collected ", val)
	count, fract := guru0.GetSubsCount()
	_ = fract
	got = fmt.Sprint("topics collected ", count)
	want = "topics collected 2"
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	sendmessage := packets.Send{}
	sendmessage.Address = []byte("contact1 address")
	sendmessage.Source = []byte("contact2 address")
	sendmessage.Payload = []byte("can you hear me now?")

	iot.Push(contact2, &sendmessage)

	WaitForActions(guru0)
	WaitForActions(aide1)
	WaitForActions(aide2)
	WaitForActions(guru0)

	got = contact1.(*testContact).getResultAsString()
	want = `[P,"contact1 address",=ygRnE97Kfx0usxBqx5cygy4enA1eojeRWdV/XMwSGzw,"contact2 address",,"can you hear me now?"]`
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	sendmessage2 := packets.Send{}
	sendmessage2.Address = []byte("contact1 address")
	sendmessage2.Source = []byte("contact2 address")
	sendmessage2.Payload = []byte("how about now?")

	iot.Push(contact2, &sendmessage2)

	WaitForActions(guru0)
	WaitForActions(aide1)
	WaitForActions(aide2)
	WaitForActions(guru0)

	got = contact1.(*testContact).getResultAsString()
	want = `[P,"contact1 address",=ygRnE97Kfx0usxBqx5cygy4enA1eojeRWdV/XMwSGzw,"contact2 address",,"can you hear me now?"]`
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	_ = got
	_ = want
	_ = err
	_ = ok

	WaitForActions(guru0)
	WaitForActions(aide1)
	WaitForActions(aide2)
	WaitForActions(guru0)

}

func TestSend(t *testing.T) {

	tokens.LoadPublicKeys()

	got := ""
	want := ""
	ok := true
	var err error
	currentTime = starttime

	// set up
	guru := iot.NewExecutive(100, "guru", getTime, true)

	// make a contact
	contact1 := MakeTestContact(guru.Config) //testContact{}
	//contact1.downMessages = make(chan packets.Interface, 1000)
	//iot.AddContactStruct(&contact1.ContactStruct, &contact1, guru.Config)
	// another
	contact2 := MakeTestContact(guru.Config) //testContact{}
	//contact2.downMessages = make(chan packets.Interface, 1000)
	//iot.AddContactStruct(&contact2.ContactStruct, &contact2, guru.Config)

	connect := packets.Connect{}
	connect.SetOption("token", []byte(tokens.SampleSmallToken))
	iot.Push(contact1, &connect)
	iot.Push(contact2, &connect)

	// subscribe
	subs := packets.Subscribe{}
	subs.Address = []byte("contact1_address")
	err = iot.Push(contact1, &subs)
	subs = packets.Subscribe{}
	subs.Address = []byte("contact2_address")
	err = iot.Push(contact2, &subs)

	got = contact1.(*testContact).getResultAsString()
	want = "no message received"
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	WaitForActions(guru)

	val := readCounter(iot.TopicsAdded)
	got = fmt.Sprint("topics collected ", val)
	count, fract := guru.GetSubsCount()
	_ = fract
	got = fmt.Sprint("topics collected ", count)
	want = "topics collected 3" //
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	sendmessage := packets.Send{}
	sendmessage.Address = []byte("contact1_address")
	sendmessage.Source = []byte("contact2_address")
	sendmessage.Payload = []byte("hello, can you hear me")

	iot.Push(contact2, &sendmessage)

	WaitForActions(guru)

	got = contact1.(*testContact).getResultAsString()
	want = `[P,contact1_address,=zC7beEa1uwyGGqQpWw+CxYn8/A8IV3bhYkAfKKktWv4,contact2_address,,"hello, can you hear me"]`
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	// how do we test that it's there?
	lookmsg := packets.Lookup{}
	lookmsg.Address = []byte("contact1_address")
	lookmsg.Source = []byte("contact2_address")
	iot.Push(contact2, &lookmsg)

	WaitForActions(guru)

	got = contact2.(*testContact).getResultAsString()
	want = `[L,contact1_address,=zC7beEa1uwyGGqQpWw+CxYn8/A8IV3bhYkAfKKktWv4,contact2_address,,count,1]`
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	unsub := packets.Unsubscribe{}
	unsub.Address = []byte("contact1_address")
	err = iot.Push(contact1, &unsub)

	lookmsg = packets.Lookup{}
	lookmsg.Address = []byte("contact1_address")
	lookmsg.Source = []byte("contact2_address")
	iot.Push(contact2, &lookmsg)

	WaitForActions(guru)

	got = contact2.(*testContact).getResultAsString()
	// note that the count is ZERO
	want = `[L,contact1_address,=zC7beEa1uwyGGqQpWw+CxYn8/A8IV3bhYkAfKKktWv4,contact2_address,,count,0]`
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	_ = ok
	_ = err

}
