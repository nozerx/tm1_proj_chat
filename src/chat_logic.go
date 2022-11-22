package src

import (
	"context"
	"encoding/json"
	"fmt"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

type chatmessage struct {
	Message    string
	SenderID   string
	SenderName string
}

type Group struct {
	Host      *p2pHost
	Inbound   chan chatmessage
	Outbound  chan string
	GroupName string
	UserName  string
	pscntx    context.Context
	pscancel  context.CancelFunc
	pstopic   *pubsub.Topic
	psub      *pubsub.Subscription
	selfid    peer.ID
}

func JoinGroup(p2phost *p2pHost) (*Group, error) {
	topic, err := p2phost.PubSub.Join("ChatGroup")
	if err != nil {
		fmt.Println("Error while join the chat")
		return nil, err
	}
	sub, err := topic.Subscribe()
	if err != nil {
		fmt.Println("Error while subscribing", sub)
		return nil, err
	}
	username := "user1"
	groupname := "Group1"
	pubsubctx, cancel := context.WithCancel(context.Background())
	chatgroup := &Group{
		Host:      p2phost,
		Inbound:   make(chan chatmessage),
		Outbound:  make(chan string),
		pscntx:    pubsubctx,
		pscancel:  cancel,
		pstopic:   topic,
		psub:      sub,
		UserName:  username,
		GroupName: groupname,
		selfid:    p2phost.Host.ID(),
	}
	go chatgroup.PubLoop()
	go chatgroup.SubLoop()
	return chatgroup, nil

}

func (gr *Group) PubLoop() {
	for {
		select {
		case <-gr.pscntx.Done():
			return
		case message := <-gr.Outbound:
			m := chatmessage{
				Message:    message,
				SenderID:   string(gr.selfid),
				SenderName: gr.UserName,
			}
			messagebyte, err := json.Marshal(m)
			if err != nil {
				fmt.Println("Error in marshaling")
				continue
			}
			err = gr.pstopic.Publish(gr.pscntx, messagebyte)
			if err != nil {
				fmt.Println("Error in publishing the message")
				continue
			}

		}
	}
}

func (gr *Group) SubLoop() {
	for {
		select {
		case <-gr.pscntx.Done():
			return
		default:
			message, err := gr.psub.Next(gr.pscntx)
			if err != nil {
				close(gr.Inbound)
				fmt.Println("Error while trying to read a message from a subscription")
				return
			}

			if message.ReceivedFrom == gr.selfid {
				continue
			}
			cm := &chatmessage{}
			err = json.Unmarshal(message.Data, cm)
			if err != nil {
				fmt.Println("Error during unmarshalling")
				continue
			}
			gr.Inbound <- *cm
		}
	}
}

func (gr *Group) Exit() {
	defer gr.pscancel()
}
