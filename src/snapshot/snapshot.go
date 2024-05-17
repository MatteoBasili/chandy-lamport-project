package snapshot

import (
	"fmt"
	"github.com/DistributedClocks/GoVector/govec"
	"sdccProject/src/utils"
)

type SnapNode struct {
	nodeIdx  int
	NetNodes []utils.Node

	NodeState      utils.NodeState
	ChannelsStates map[int]utils.ChState

	CurrentStateCh chan utils.FullState
	RecvStateCh    chan utils.FullState
	SendMarkCh     chan utils.AppMessage
	RecvMarkMsgCh  chan utils.AppMessage
	SendMsgCh      chan utils.AppMessage
	AppGSCh        chan utils.GlobalState
	InternalGsCh   chan utils.GlobalState
	IsLauncher     bool
	Logger         *utils.Logger
}

func NewSnapNode(netIdx int, currentStateCh chan utils.FullState, recvStateCh chan utils.FullState, sendMarkCh chan utils.AppMessage, recvMarkCh chan utils.AppMessage, sendMsgCh chan utils.AppMessage, netLayout *utils.NetLayout, logger *utils.Logger) *SnapNode {
	var myNode = netLayout.Nodes[netIdx]

	// Initialize channels state
	chsState := make(map[int]utils.ChState)
	for idx, node := range netLayout.Nodes {
		if idx != netIdx {
			chsState[node.Idx] = utils.ChState{
				RecvMsgs:  make([]utils.AppMessage, 0),
				Recording: false,
			}
		}
	}

	snapNode := &SnapNode{
		nodeIdx:  netIdx,
		NetNodes: netLayout.Nodes,
		NodeState: utils.NodeState{
			NodeName:     myNode.Name,
			SentMsgs:     make(map[string]utils.AppMessage),
			ReceivedMsgs: make([]utils.AppMessage, 0),
		},
		ChannelsStates: chsState,
		CurrentStateCh: currentStateCh,
		RecvStateCh:    recvStateCh,
		SendMarkCh:     sendMarkCh,
		RecvMarkMsgCh:  recvMarkCh,
		SendMsgCh:      sendMsgCh,
		InternalGsCh:   make(chan utils.GlobalState),
		IsLauncher:     false,
		Logger:         logger,
	}
	go snapNode.waitForSnapshot()
	return snapNode
}

func (n *SnapNode) MakeSnapshot() utils.GlobalState {
	n.NodeState.Busy = true // While Busy cannot send new msg
	n.Logger.Info.Println("Initializing snapshot...")
	// Save node state, all prerecording msg (sent btw | prev-state ---- mark | are store on n.NodeState.SendAppMsg
	n.IsLauncher = true

	n.Logger.Info.Println("Saving state...")
	// Save state
	n.CurrentStateCh <- utils.FullState{
		Node:         n.NodeState,
		Channels:     n.ChannelsStates,
		AllMarksRecv: false,
	}

	// Send markers
	n.Logger.Info.Println("Sending first Mark...")
	n.SendMarkCh <- utils.NewMarkMsg(n.nodeIdx)

	// Start channels recording
	for chKey := range n.ChannelsStates {
		n.ChannelsStates[chKey] = utils.ChState{
			RecvMsgs:  make([]utils.AppMessage, 0),
			Recording: true,
		}
	}

	gs := <-n.InternalGsCh
	return gs
}

func (n *SnapNode) recvMsgMark(nMarks *int8, msg utils.AppMessage) bool {
	// Recv a mark
	if msg.IsMarker {
		*nMarks = *nMarks + 1
		mark := msg
		// First mark recv, save process state
		if !n.NodeState.Busy {
			n.NodeState.Busy = true // While Busy cannot send new msg
			n.Logger.Info.Printf("Recv first MARK from %s\n", n.NetNodes[mark.From].Name)

			n.Logger.Info.Println("Saving state...")
			// Save state
			n.CurrentStateCh <- utils.FullState{
				Node:         n.NodeState,
				Channels:     n.ChannelsStates,
				AllMarksRecv: false,
			}

			// Send broadcast marks
			n.Logger.Info.Printf("Send broadcast Markers\n")
			n.SendMarkCh <- utils.NewMarkMsg(n.nodeIdx)

			// Start channels recording
			for chKey := range n.ChannelsStates {
				if chKey != n.NetNodes[mark.From].Idx {
					n.ChannelsStates[chKey] = utils.ChState{
						RecvMsgs:  make([]utils.AppMessage, 0),
						Recording: true,
					}
				} else {
					n.ChannelsStates[chKey] = utils.ChState{
						RecvMsgs:  make([]utils.AppMessage, 0),
						Recording: false,
					}
				}
			}
		} else {
			// NOT First mark recv, stop recording channel
			n.Logger.Info.Printf("Recv another MARK from %s\n", n.NetNodes[mark.From].Name)
			tempChState := n.ChannelsStates[n.NetNodes[mark.From].Idx]
			tempChState.Recording = false
			n.ChannelsStates[n.NetNodes[mark.From].Idx] = tempChState
		}

		if *nMarks == int8(len(n.NetNodes)) {
			// Send current state to all
			n.Logger.Info.Printf("Recv all MARKs\n")
			n.Logger.GoVector.LogLocalEvent("Recv all MARKs", govec.GetDefaultLogOptions())
			n.Logger.Info.Println("Sending my state to all")
			n.CurrentStateCh <- utils.FullState{
				Node:         n.NodeState,
				Channels:     n.ChannelsStates,
				AllMarksRecv: true,
			}
			return true
		}
	} else { // Recv a msg
		if n.NodeState.Busy {
			// Save msg as post-recording
			chState := n.ChannelsStates[n.NetNodes[msg.From].Idx]
			chState.RecvMsgs = append(chState.RecvMsgs, msg)
			n.ChannelsStates[n.NetNodes[msg.From].Idx] = chState
		} else { // Save msg on node state
			n.NodeState.ReceivedMsgs = append(n.NodeState.ReceivedMsgs, msg)
		}
	}
	return false
}

func (n *SnapNode) recvAllMarks() {
	// Receive all marks
	var nMarks int8 = 1 // my mark
	for {
		var msg utils.AppMessage
		select {
		case msg = <-n.RecvMarkMsgCh: // Recv mark or msg
			if n.recvMsgMark(&nMarks, msg) {
				return
			}
		case detMsg := <-n.SendMsgCh: // node send msg
			n.NodeState.SentMsgs[n.NetNodes[detMsg.To].Name] = detMsg
		}
	}
}

func (n *SnapNode) waitForSnapshot() {
	for {
		n.recvAllMarks()

		// Gather global status and send to app
		n.Logger.Info.Println("Beginning to gather states...")
		var gs utils.GlobalState
		gs.GS = append(gs.GS, utils.FullState{
			Node:         n.NodeState,
			Channels:     n.ChannelsStates,
			AllMarksRecv: true,
		})
		for i := 0; i < len(n.NetNodes)-1; i++ {
			indState := <-n.RecvStateCh
			n.Logger.Info.Printf("Recv State from: %s\n", indState.Node.NodeName)
			gs.GS = append(gs.GS, indState)
			fmt.Println("$", gs.GS[i].Node.Balance)
		}
		n.Logger.Info.Println("All states gathered")

		// Restore process state
		n.NodeState.Busy = false
		n.NodeState.SentMsgs = make(map[string]utils.AppMessage)
		n.NodeState.ReceivedMsgs = make([]utils.AppMessage, 0)

		// Inform node to continue receiving msg
		n.CurrentStateCh <- utils.FullState{
			Node:         n.NodeState,
			Channels:     n.ChannelsStates,
			AllMarksRecv: false,
		}

		// Send gs to launcher
		if n.IsLauncher {
			n.InternalGsCh <- gs
			n.IsLauncher = false
		}
	}
}
