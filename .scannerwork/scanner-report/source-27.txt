package utils_test

import (
	"chandy_lamport/src/utils"
	"testing"
)

func TestNodeState_String(t *testing.T) {
	nodeState := utils.NodeState{
		NodeName: "P1",
		Balance:  100,
		SentMsgs: []utils.AppMessage{
			{Msg: utils.Message{ID: "msg1"}, IsMarker: false, From: 1, To: 2},
			{Msg: utils.Message{ID: "msg2"}, IsMarker: false, From: 1, To: 3},
		},
		ReceivedMsgs: []utils.AppMessage{
			{Msg: utils.Message{ID: "msg3"}, IsMarker: false, From: 2, To: 1},
			{Msg: utils.Message{ID: "msg4"}, IsMarker: false, From: 3, To: 1},
		},
		Busy: true,
	}

	expected := "- Balance: $100\n- Messages sent: [ msg1, msg2 ]\n- Messages received: [ msg3, msg4 ]"
	if nodeState.String() != expected {
		t.Errorf("Expected %q, but got %q", expected, nodeState.String())
	}
}

func TestChState_String(t *testing.T) {
	chState := utils.ChState{
		RecvMsgs: []utils.AppMessage{
			{Msg: utils.Message{ID: "msg1"}, IsMarker: false, From: 1, To: 2},
			{Msg: utils.Message{ID: "msg2"}, IsMarker: false, From: 2, To: 1},
		},
		Recording: true,
	}

	expected := "(Channel is still recording!!!!!!!!!!!!) "
	expected += "msg1, msg2"
	if chState.String() != expected {
		t.Errorf("Expected %q, but got %q", expected, chState.String())
	}
}

func TestFullState_String(t *testing.T) {
	fullState := utils.FullState{
		Node: utils.NodeState{
			NodeName: "P1",
			Balance:  100,
			SentMsgs: []utils.AppMessage{
				{Msg: utils.Message{ID: "msg1"}, IsMarker: false, From: 1, To: 2},
			},
			ReceivedMsgs: []utils.AppMessage{
				{Msg: utils.Message{ID: "msg2"}, IsMarker: false, From: 2, To: 1},
			},
			Busy: true,
		},
		Channels: map[int]utils.ChState{
			1: {RecvMsgs: []utils.AppMessage{{Msg: utils.Message{ID: "msg3"}, IsMarker: false, From: 3, To: 1}}},
			2: {RecvMsgs: []utils.AppMessage{{Msg: utils.Message{ID: "msg4"}, IsMarker: false, From: 1, To: 2}}},
		},
		AllMarksRecv: false,
	}

	expected := "\nState:\n- Balance: $100\n- Messages sent: [ msg1 ]\n- Messages received: [ msg2 ]\nChannels:\n[ 1 ] ==> msg3\n[ 2 ] ==> msg4\n----------------NOT RECEIVED ALL MARKS---------"
	if fullState.String() != expected {
		t.Errorf("Expected %q, but got %q", expected, fullState.String())
	}
}

func TestGlobalState_String(t *testing.T) {
	globalState := utils.GlobalState{
		GS: []utils.FullState{
			{
				Node: utils.NodeState{
					NodeName: "P1",
					Balance:  100,
				},
				Channels: map[int]utils.ChState{
					1: {RecvMsgs: []utils.AppMessage{{Msg: utils.Message{ID: "msg1"}, IsMarker: false, From: 2, To: 1}}},
				},
				AllMarksRecv: true,
			},
		},
	}

	expected := "\n-------------------------------------- Process P1: \nState:\n- Balance: $100\n- Messages sent: [ ]\n- Messages received: [ ]\nChannels:\n[ 1 ] ==> msg1\n"
	if globalState.String() != expected {
		t.Errorf("Expected %q, but got %q", expected, globalState.String())
	}
}
