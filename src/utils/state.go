package utils

import (
	"fmt"
	"sort"
	"strconv"
)

type NodeState struct {
	NodeName     string
	Balance      int
	SentMsgs     []AppMessage
	ReceivedMsgs []AppMessage
	Busy         bool // Process is doing a snapshot
}

func (n NodeState) String() string {
	var res string
	res += "- Balance: $" + strconv.Itoa(n.Balance) + "\n"
	res += "- Messages sent: ["
	for i, m := range n.SentMsgs {
		if i == len(n.SentMsgs)-1 {
			res += fmt.Sprintf(" %s", m.Msg.ID)
		} else {
			res += fmt.Sprintf(" %s,", m.Msg.ID)
		}
	}
	res += " ]\n"
	res += "- Messages received: ["
	for j, m := range n.ReceivedMsgs {
		if j == len(n.ReceivedMsgs)-1 {
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
	res := fmt.Sprintf("\nState:\n%s", as.Node.String())
	res += fmt.Sprintf("\nChannels:\n")
	var counter int
	channelsLen := len(as.Channels)

	// Extract the keys from the map
	keys := make([]int, 0, channelsLen)
	for key := range as.Channels {
		keys = append(keys, key)
	}
	// Order the keys
	sort.Ints(keys)
	// Iterate over the sorted keys and access the corresponding values
	for _, key := range keys {
		counter++
		res += fmt.Sprintf("[ %d ] ==> %s", key, as.Channels[key].String())
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
		res += "-------------------------------------- "
		res += fmt.Sprintf("Process %s: %s\n", as.Node.NodeName, as.String())
	}
	return res
}
