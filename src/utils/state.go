package utils

type ProcessState struct {
	Name         string
	Balance      int
	SentMsgs     []AppMessage
	ReceivedMsgs []AppMessage
	//Busy         bool // Process is doing a snapshot
}

/*func (n ProcessState) String() string {
	var res string
	res += "Sent: [ "
	for node, m := range n.SentMsgs {
		res += fmt.Sprintf(" %s-> %s,", m.Body, node)
	}
	res += " ], "
	res += "Recv: [ "
	for _, m := range n.RecvMsg {
		res += fmt.Sprintf(" %s,", m.Body)
	}
	res += " ]"
	return res
}*/

type ChState struct {
	RecvMsgs  []AppMessage
	Recording bool
}

/*func (cs ChState) String() string {
	var res string
	if cs.Recording {
		res += "Channel is still recording!!!!!!!!!!!!"
	}
	for _, m := range cs.RecvMsg {
		res += fmt.Sprintf(" %s,", m.Body)
	}
	return res
}*/

type FullState struct {
	Process  ProcessState
	Channels map[string]ChState
	//RecvAllMarks bool
}

/*func (as FullState) String() string {
	res := fmt.Sprintf("\nState: %s", as.Node)
	res += fmt.Sprintf("\nChannels:\n")
	for chKey := range as.Channels {
		res += fmt.Sprintf("[ %s ] ==> %s;", chKey, as.Channels[chKey])
	}
	if !as.RecvAllMarks {
		res += "\n----------------NOT RECEIVED ALL MARKS---------"
	}
	return res
}*/

type GlobalState struct {
	GS []FullState
}

/*
func (gs GlobalState) String() string {
	res := "\n"
	for _, as := range gs.GS {
		res += "--------------------------------------"
		res += fmt.Sprintf("Node %s: %s\n", as.Node.Name, as)
	}
	return res
}*/
