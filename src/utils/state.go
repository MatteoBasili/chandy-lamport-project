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
	res += "Balance: $" + strconv.Itoa(n.Balance) + ", "
	res += "Sent: [ "
	for node, m := range n.SentMsgs {
		res += fmt.Sprintf(" %s-> %s,", m.Msg.ID, node)
	}
	res += " ], "
	res += "Recv: [ "
	for _, m := range n.ReceivedMsgs {
		res += fmt.Sprintf(" %s,", m.Msg.ID)
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
		res += "Channel is still recording!!!!!!!!!!!!"
	}
	for _, m := range cs.RecvMsgs {
		res += fmt.Sprintf(" %s,", m.Msg.ID)
	}
	return res
}

type FullState struct {
	Node         NodeState
	Channels     map[int]ChState
	AllMarksRecv bool
}

func (as FullState) String() string {
	res := fmt.Sprintf("\nState: %s", as.Node)
	res += fmt.Sprintf("\nChannels:\n")
	for chKey := range as.Channels {
		res += fmt.Sprintf("[ %d ] ==> %s;", chKey, as.Channels[chKey])
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
