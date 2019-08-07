// Copyright 2019 Alan Tracey Wootton

package iot

import (
	"fmt"

	"knotfreeiot/iot/reporting"
	"strconv"
)

// PubsubIntf is stuff that deals with pub/sub. The other part of this interface though
// is the Write function in the SocketStructConfig which is what the pubsubmgr uses to distribute
// messages that are published.
type PubsubIntf interface {
	SendSubscriptionMessage(topic []byte, ss *SockStruct)
	SendUnsubscribeMessage(topic []byte, ss *SockStruct)
	SendPublishMessage(topic []byte, ss *SockStruct, payload []byte)
	//SendOnlineQuery(topic string, ss *SockStruct )
	GetAllSubsCount() (int, int)
}

// NewPubsubManager makes a SubscriptionsIntf, usually a singleton.
// In the tests we call here and then use the result to init a server.
func NewPubsubManager(amt int) PubsubIntf {
	psMgr := pubSubManager{}
	psMgr.key.Random()
	fmt.Println("NewPubsubManager", psMgr.key.String())
	portion := amt / int(theBucketsSize)
	portion2 := amt >> theBucketsSizeLog2 // we can init the hash maps big
	if portion != portion2 {
		fmt.Println("theBucketsSizeLog2 != uint(math.Log2(float64(theBucketsSize)))")
	}
	psMgr.allTheSubscriptions = make([]subscribeBucket, theBucketsSize)
	for i := uint(0); i < theBucketsSize; i++ {
		psMgr.allTheSubscriptions[i].mySubscriptions = make(map[HashType]*watchedTopic, portion)
		tmp := make(chan interface{}, 32)
		psMgr.allTheSubscriptions[i].incoming = &tmp
		psMgr.allTheSubscriptions[i].subscriber = &psMgr
		go psMgr.allTheSubscriptions[i].processMessages(&psMgr)
	}

	subscribeEvents := reporting.NewStringEventAccumulator(12)
	subscribeEvents.SetQuiet(true)
	psMgr.subscribeEvents = subscribeEvents

	subscrFRepofrtFunct := func(seconds float32) []string {
		strlist := make([]string, 0, 5)
		count := 0
		for _, b := range psMgr.allTheSubscriptions {
			count += len(b.mySubscriptions)
		}
		strlist = append(strlist, "Topic count="+strconv.Itoa(count))
		return strlist
	}
	reporting.NewGenericEventAccumulator(subscrFRepofrtFunct)

	// parents := *iot.SockStructConfig

	return &psMgr
}

// SendSubscriptionMessage will create a message object, copy pointers to it so it'll own them now, and queue the message.
func (me *pubSubManager) SendSubscriptionMessage(realName []byte, ss *SockStruct) {
	topic := HashType{}
	topic.FromBytes(realName)
	ss.topicToName[HalfHash(topic.a)] = realName
	msg := subscriptionMessage{}
	msg.Topic = &topic
	msg.ss = ss
	i := topic.GetFractionalBits(theBucketsSizeLog2)
	b := me.allTheSubscriptions[i]
	*b.incoming <- msg
}

// SendUnsubscribeMessage will create a message object, copy pointers to it so it'll own them now, and queue the message.
func (me *pubSubManager) SendUnsubscribeMessage(realName []byte, ss *SockStruct) {

	topic := HashType{}
	topic.FromBytes(realName)

	delete(ss.topicToName, HalfHash(topic.a))

	msg := unsubscribeMessage{}
	msg.Topic = &topic
	msg.ss = ss
	//msg.ConnectionID = c.GetKey()
	i := topic.GetFractionalBits(theBucketsSizeLog2)
	b := me.allTheSubscriptions[i]
	*b.incoming <- msg
}

// SendPublishMessage will create a message object, copy pointers to it so it'll own them now, and queue the message.
func (me *pubSubManager) SendPublishMessage(realName []byte, ss *SockStruct, payload []byte) {

	topic := HashType{}
	topic.FromBytes(realName)

	msg := publishMessage{}
	msg.Topic = &topic
	msg.payload = payload
	msg.ss = ss
	msg.returnAddress = ss.GetSelfAddress()

	i := topic.GetFractionalBits(theBucketsSizeLog2)
	b := me.allTheSubscriptions[i]
	*b.incoming <- msg
}

// GetAllSubsCount returns the count of subscriptions and the
// average depth of the channels.

func (me *pubSubManager) GetAllSubsCount() (int, int) {
	count := 0
	qdepth := 0
	for _, b := range me.allTheSubscriptions {
		count += len(b.mySubscriptions)
		qdepth += (len(*b.incoming))
	}
	qdepth = qdepth / len(me.allTheSubscriptions)
	return count, qdepth
}

// now it gets private

// TODO: implement a pool of the incoming types.

// PrivateSendUnsubscribeMessage is only for use by SockStruct Close()
// func PrivateSendUnsubscribeMessage(me *pubSubManager, topic *HashType, ss *SockStruct) {

// 	msg := unsubscribeMessage{}
// 	msg.Topic = topic
// 	msg.ss = ss
// 	i := topic.GetFractionalBits(theBucketsSizeLog2)
// 	b := me.allTheSubscriptions[i]
// 	*b.incoming <- msg
// }

func (me *pubSubManager) checkForBadSS(badsock *SockStruct, pubstruct *watchedTopic) bool {

	forgetme := false
	if badsock.conn == nil {
		forgetme = true
	}
	if badsock.ele == nil {
		forgetme = true
	}
	if forgetme {
		for topic, realName := range badsock.topicToName {
			me.SendUnsubscribeMessage(realName, badsock)
			badsock.topicToName = nil
			_ = topic
		}
		delete(pubstruct.watchers, badsock.key)
		return true
	}
	return false
}

type parentsIntf struct {
	setPublishCallback (func(topic *HashType))
}

func (*parentsIntf) init(topics []HashType) {

}

func (*parentsIntf) subscibe(topics []HashType) {

}
func (*parentsIntf) unsubscibe(topics []HashType) {

}

func (bucket *subscribeBucket) processMessages(me *pubSubManager) {

	for {
		msg := <-*bucket.incoming // wait right here
		switch msg.(type) {

		case subscriptionMessage:
			submsg := msg.(subscriptionMessage)
			substruct := bucket.mySubscriptions[*submsg.Topic]
			if substruct == nil {
				substruct = &watchedTopic{}
				substruct.name.FromHashType(submsg.Topic)
				substruct.watchers = make(map[HalfHash]*SockStruct, 0)
				bucket.mySubscriptions[*submsg.Topic] = substruct
				bucket.subscriber.subscribeEvents.Collect("subscription")
			}
			// this is the important part:
			// add the caller to  the set
			//fmt.Println("pubsub ", bucket.subscriber.key.String(), " sub ", submsg.Topic.a&0x0FFFF)
			substruct.watchers[submsg.ss.key] = submsg.ss

			//parent.subscribe(submsg.Topic)

		case publishMessage:
			pubmsg := msg.(publishMessage)
			pubstruct, ok := bucket.mySubscriptions[*pubmsg.Topic]
			//fmt.Println("pubsub ", bucket.subscriber.key.String(), " pub  ", pubmsg.Topic.a&0x0FFFF)
			if ok == false {
				// no publish possible !
				// it's sad really when someone sends messages to nobody.
				// TODO: we need an Online function for topics.
			} else {
				for key, ss := range pubstruct.watchers {
					if key != pubmsg.ss.key {
						if me.checkForBadSS(ss, pubstruct) == false {
							realName := ss.topicToName[HalfHash(pubmsg.Topic.a)]

							ss.config.writer(ss, realName, pubmsg.payload, pubmsg.returnAddress)
						}
					}
				}
			}

		case unsubscribeMessage:

			unmsg := msg.(unsubscribeMessage)
			unstruct, ok := bucket.mySubscriptions[*unmsg.Topic]
			if ok == true {
				bucket.subscriber.subscribeEvents.Collect("unsubscribe")
				delete(unstruct.watchers, unmsg.ss.key)
				if len(unstruct.watchers) == 0 {
					// forget the entire topic
					delete(bucket.mySubscriptions, *unmsg.Topic)
				}
			}

		default:
			// no match. do nothing.
		}
	}
}

// theBucketsSize is 4 for debug and 1024 for prod
// it's just to keep the threads busy.
const theBucketsSize = uint(8) // uint(1024)
const theBucketsSizeLog2 = 3   // 10 // uint(math.Log2(float64(theBucketsSize)))

type subscriptionMessage struct {
	Topic *HashType // not my real name
	//	ConnectionID *HashType
	ss *SockStruct
}

// unsubscribeMessage for real
type unsubscribeMessage struct {
	subscriptionMessage
}

// publishMessage used here
type publishMessage struct {
	subscriptionMessage
	payload       []byte
	returnAddress []byte
}

// watchedTopic, this is private here
type watchedTopic struct {
	name     HashType                 // not my real name
	watchers map[HalfHash]*SockStruct // these are ID's for tcp Connection mgr
}

type subscribeBucket struct {
	mySubscriptions map[HashType]*watchedTopic
	incoming        *chan interface{} //SubscriptionMessage
	subscriber      *pubSubManager
}

// // this is the whole point:
// // implements SubscriptionsIntf
type pubSubManager struct {
	allTheSubscriptions []subscribeBucket
	subscribeEvents     *reporting.StringEventAccumulator
	key                 HashType

	parents *SockStructConfig
	clients *SockStructConfig
}