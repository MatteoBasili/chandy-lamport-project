package snapshot

import (
	"github.com/DistributedClocks/GoVector/govec"
	"sdccProject/src/utils"
)

type SnapNode struct {
	nodeIdx  int
	NetNodes []utils.Node

	NodeState      utils.NodeState
	ChannelsStates map[int]utils.ChState
	nMarks         int8

	SendMsgCh    chan utils.AppMessage
	StatesCh     utils.StatesChannels
	MarkCh       utils.MarkChannels
	InternalGsCh chan utils.GlobalState
	IsLauncher   bool
	Logger       *utils.Logger
}

func NewSnapNode(netIdx int, sendMsgCh chan utils.AppMessage, statesCh utils.StatesChannels, markCh utils.MarkChannels, netLayout *utils.NetLayout, logger *utils.Logger) *SnapNode {
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
			Balance:      -1,
			SentMsgs:     make([]utils.AppMessage, 0),
			ReceivedMsgs: make([]utils.AppMessage, 0),
		},
		nMarks:         1,
		ChannelsStates: chsState,
		StatesCh:       statesCh,
		MarkCh:         markCh,
		SendMsgCh:      sendMsgCh,
		InternalGsCh:   make(chan utils.GlobalState),
		IsLauncher:     false,
		Logger:         logger,
	}
	go snapNode.wait()
	return snapNode
}

func (n *SnapNode) MakeSnapshot() utils.GlobalState {
	n.NodeState.Busy = true // While Busy cannot send new msg
	n.Logger.Info.Println("Initializing snapshot...")
	// Save node state, all prerecording msg (sent btw | prev-state ---- mark | are store on n.NodeState.SendAppMsg
	n.IsLauncher = true

	n.Logger.Info.Println("Saving state...")
	// Save state
	n.saveProcState()

	// Send markers
	n.Logger.Info.Println("Sending broadcast MARKERs...")
	n.sendBroadMark()

	// Start channels recording
	n.startRecChs(n.nodeIdx)

	gs := <-n.InternalGsCh
	return gs
}

func (n *SnapNode) saveProcState() {
	n.StatesCh.CurrCh <- utils.FullState{
		Node:         n.NodeState,
		Channels:     n.ChannelsStates,
		AllMarksRecv: false,
	}
	state := <-n.StatesCh.SaveCh
	n.NodeState.Balance = state.Node.Balance
}

func (n *SnapNode) sendBroadMark() {
	n.MarkCh.SendCh <- utils.NewMarkMsg(n.nodeIdx)
}

func (n *SnapNode) startRecChs(nodeIdx int) {
	if nodeIdx == n.nodeIdx {
		for chKey := range n.ChannelsStates {
			n.ChannelsStates[chKey] = utils.ChState{
				RecvMsgs:  make([]utils.AppMessage, 0),
				Recording: true,
			}
		}
	} else {
		for chKey := range n.ChannelsStates {
			if chKey != nodeIdx {
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
	}
}

func (n *SnapNode) stopRecCh(nodeIdx int) {
	tempChState := n.ChannelsStates[nodeIdx]
	tempChState.Recording = false
	n.ChannelsStates[nodeIdx] = tempChState
}

func (n *SnapNode) allMarksRecv(lastMark utils.AppMessage) bool {
	// First mark recv, save process state
	if !n.NodeState.Busy {
		n.NodeState.Busy = true // While Busy cannot send new msg
		n.Logger.Info.Printf("Received first MARKER from %s\n", n.NetNodes[lastMark.From].Name)

		n.Logger.Info.Println("Saving state...")
		// Save state
		n.saveProcState()

		// Send broadcast marks
		n.sendBroadMark()
		n.Logger.Info.Printf("Sent broadcast MARKERs\n")

		// Start channels recording
		n.startRecChs(lastMark.From)
	} else {

		// NOT First mark recv, stop recording channel
		n.Logger.Info.Printf("Received MARKER from %s\n", n.NetNodes[lastMark.From].Name)
		n.stopRecCh(lastMark.From)
	}

	if n.nMarks == int8(len(n.NetNodes)) {
		// Send current state to all
		n.Logger.Info.Printf("Received all MARKERs\n")
		n.Logger.GoVector.LogLocalEvent("Received all MARKERs", govec.GetDefaultLogOptions())
		n.Logger.Info.Println("Sending my state to all...")
		n.StatesCh.CurrCh <- utils.FullState{
			Node:         n.NodeState,
			Channels:     n.ChannelsStates,
			AllMarksRecv: true,
		}
		return true
	}
	return false
}

func (n *SnapNode) manageRecvMsg(msg utils.AppMessage) {
	// Recv a mark
	if msg.IsMarker {
		n.nMarks += 1
		if n.allMarksRecv(msg) {
			n.endSnapshot()
		}
	} else { // Recv a msg
		if n.ChannelsStates[msg.From].Recording {
			// Save msg as post-recording
			chState := n.ChannelsStates[n.NetNodes[msg.From].Idx]
			chState.RecvMsgs = append(chState.RecvMsgs, msg)
			n.ChannelsStates[n.NetNodes[msg.From].Idx] = chState
		} else { // Save msg on node state
			n.NodeState.ReceivedMsgs = append(n.NodeState.ReceivedMsgs, msg)
		}
	}
}

func (n *SnapNode) endSnapshot() {
	// Gather global status and send to app
	n.Logger.Info.Println("Beginning to gather states...")
	var gs utils.GlobalState
	gs.GS = append(gs.GS, utils.FullState{
		Node:         n.NodeState,
		Channels:     n.ChannelsStates,
		AllMarksRecv: true,
	})
	for i := 0; i < len(n.NetNodes)-1; i++ {
		indState := <-n.StatesCh.RecvCh
		n.Logger.Info.Printf("Received state from: %s\n", indState.Node.NodeName)
		gs.GS = append(gs.GS, indState)
	}
	n.Logger.Info.Println("All states gathered")

	// Restore process state
	n.NodeState.Busy = false
	n.NodeState.Balance = -1
	n.NodeState.SentMsgs = make([]utils.AppMessage, 0)
	n.NodeState.ReceivedMsgs = n.takePendingMessages()
	n.nMarks = 1

	// Inform node to continue receiving msg
	n.StatesCh.CurrCh <- utils.FullState{
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

func (n *SnapNode) takePendingMessages() []utils.AppMessage {
	var recvMsgs []utils.AppMessage
	for _, val := range n.ChannelsStates {
		for i := 0; i < len(val.RecvMsgs); i++ {
			recvMsgs = append(recvMsgs, val.RecvMsgs[i])
		}
	}
	return recvMsgs
}

func (n *SnapNode) wait() {
	for {
		select {
		case msg := <-n.MarkCh.RecvCh: // Recv mark or msg
			n.manageRecvMsg(msg)
		case detMsg := <-n.SendMsgCh: // node send msg
			n.NodeState.SentMsgs = append(n.NodeState.SentMsgs, detMsg)
		}
	}
}
