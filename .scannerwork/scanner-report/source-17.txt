package utils

import (
	"fmt"
	"strconv"
)

type NodeState struct {
	NodeName     string
	Balance      int
	SentMsgs     map[string]AppMessage
	ReceivedMsgs []AppMessage
	Busy         bool // Process is doing a snapshot
}

func (n NodeState) String() string {
	var res string
	res += "- Balance: $" + strconv.Itoa(n.Balance) + "\n"
	res += "- Messages sent: ["
	var lastNode string
	for key := range n.SentMsgs {
		lastNode = key
	}
	for node, m := range n.SentMsgs {
		if node == lastNode {
			res += fmt.Sprintf(" %s -> %s", m.Msg.ID, node)
		} else {
			res += fmt.Sprintf(" %s -> %s,", m.Msg.ID, node)
		}
	}
	res += " ]\n"
	res += "- Messages received: ["
	for i, m := range n.ReceivedMsgs {
		if i == len(n.ReceivedMsgs)-1 {
			res += fmt.Sprintf(" %s", m.Msg.ID)
		} else {
			res += fmt.Sprintf(" %s,", m.Msg.ID)
		}
	}
	res += " ]"
	return res
}

type ChState struct {
	RecvMsgs  []AppMessage
	Recording bool
}

func (cs ChState) String() string {
	var res string
	if cs.Recording {
		res += "(Channel is still recording!!!!!!!!!!!!) "
	}
	for i, m := range cs.RecvMsgs {
		if i == len(cs.RecvMsgs)-1 {
			res += fmt.Sprintf("%s", m.Msg.ID)
		} else {
			res += fmt.Sprintf("%s, ", m.Msg.ID)
		}
	}
	return res
}

type FullState struct {
	Node         NodeState
	Channels     map[int]ChState
	AllMarksRecv bool
}

func (as FullState) String() string {
	res := fmt.Sprintf("\nState:\n%s", as.Node)
	res += fmt.Sprintf("\nChannels:\n")
	var counter int
	channelsLen := len(as.Channels)

	for chKey, chVal := range as.Channels {
		counter++
		res += fmt.Sprintf("[ %d ] ==> %s", chKey, chVal)
		if counter != channelsLen {
			res += "\n"
		}
	}
	if !as.AllMarksRecv {
		res += "\n----------------NOT RECEIVED ALL MARKS---------"
	}
	return res
}

type GlobalState struct {
	GS []FullState
}

func (gs GlobalState) String() string {
	res := "\n"
	for _, as := range gs.GS {
		res += "--------------------------------------"
		res += fmt.Sprintf("Node %s: %s\n", as.Node.NodeName, as)
	}
	return res
}
