package process

import (
	"fmt"
	"github.com/DistributedClocks/GoVector/govec"
	"net"
	"sdccProject/src/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	Period = 2000 * time.Millisecond
)

type Process struct {
	Info           utils.Node
	NetLayout      utils.NetLayout
	NetIdx         int
	Balance        int
	FullState      utils.FullState
	Listener       net.Listener
	CurrentStateCh chan utils.FullState
	RecvStateCh    chan utils.FullState
	SendMarkCh     chan utils.AppMessage
	RecvMarkCh     chan utils.AppMessage
	SendMsgCh      chan utils.AppMessage
	SendAppMsgCh   chan utils.RespMessage
	RecvAppMsgCh   chan utils.AppMessage
	Logger         *utils.Logger
	Mutex          sync.Mutex
}

func NewProcess(netIdx int, currentStateCh chan utils.FullState, recvStateCh chan utils.FullState, sendMarkCh chan utils.AppMessage, recvMarkCh chan utils.AppMessage, sendMsgCh chan utils.AppMessage, sendAppMsgCh chan utils.RespMessage, recvAppMsgCh chan utils.AppMessage, netLayout utils.NetLayout, logger *utils.Logger) *Process {
	var myNode = netLayout.Nodes[netIdx]

	// Open Listener port
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(myNode.Port))
	if err != nil {
		panic(fmt.Sprintf("ERROR: unable to open port: %s. Error: %s.", strconv.Itoa(myNode.Port), err))
	}

	var myProcess = Process{
		Info:           myNode,
		NetLayout:      netLayout,
		NetIdx:         netIdx,
		Balance:        netLayout.InitialBalance,
		FullState:      utils.FullState{},
		Listener:       listener,
		CurrentStateCh: currentStateCh,
		RecvStateCh:    recvStateCh,
		SendMarkCh:     sendMarkCh,
		RecvMarkCh:     recvMarkCh,
		SendMsgCh:      sendMsgCh,
		SendAppMsgCh:   sendAppMsgCh,
		RecvAppMsgCh:   recvAppMsgCh,
		Logger:         logger,
		Mutex:          sync.Mutex{},
	}
	myProcess.Logger.Trace.Printf("Listening on port: %s. Initial balance : $%d", strconv.Itoa(myNode.Port), myProcess.Balance)

	go myProcess.sender()
	go myProcess.receiver()
	return &myProcess
}

func (p *Process) receiver() *utils.Message {
	var conn net.Conn
	var err error
	var recvData []byte
	recvData = make([]byte, 1024)

	for {
		p.Logger.Trace.Println("Waiting for connection accept...")
		if conn, err = p.Listener.Accept(); err != nil {
			p.Logger.Error.Panicf("Server accept connection error: %s", err)
		}
		nBytes, err := conn.Read(recvData[0:])
		if err != nil {
			p.Logger.Error.Panicf("Server accept connection error: %s", err)
		}
		if !strings.Contains(string(recvData[0:nBytes]), "Channels") {
			// Waiting for MSG or marks
			var recvMsg utils.AppMessage
			p.Logger.GoVector.UnpackReceive("Receiving Message", recvData[0:nBytes], &recvMsg, govec.GetDefaultLogOptions())
			// Send data to snapshot
			if recvMsg.IsMarker {
				p.RecvMarkCh <- recvMsg
				p.Logger.Info.Printf("MARK recv from: %s\n", p.NetLayout.Nodes[recvMsg.From].Name)
			} else {
				p.Logger.GoVector.LogLocalEvent(fmt.Sprintf("MSG %s, content: $%d, from [%s]", recvMsg.Msg.ID, recvMsg.Msg.Body, p.NetLayout.Nodes[recvMsg.From].Name), govec.GetDefaultLogOptions())
				p.UpdateBalance(recvMsg.Msg.Body, "received")
				p.Logger.Info.Printf("MSG %s [Amount: %d] received from: %s\n", recvMsg.Msg.ID, recvMsg.Msg.Body, p.NetLayout.Nodes[recvMsg.From].Name)
				p.RecvAppMsgCh <- recvMsg
			}
		} else {
			var tempState = utils.FullState{}
			p.Logger.GoVector.UnpackReceive("Receiving State", recvData[0:nBytes], &tempState, govec.GetDefaultLogOptions())
			p.Logger.Info.Println("State recv from: ", tempState.Node.NodeName)
			// Send state to snapshot
			p.RecvStateCh <- tempState
		}
	}
}

func (p *Process) sender() {
	opts := govec.GetDefaultLogOptions()
	var outBuf []byte
	outBuf = []byte{'A', 'B'}
	for {
		select {
		case respMsg := <-p.SendAppMsgCh:
			p.Mutex.Lock()
			locState := p.FullState
			p.Mutex.Unlock()
			if !locState.Node.Busy { // it is not performing a global snapshot
				detMsg := respMsg.AppMsg
				responseCh := respMsg.RespCh

				if detMsg.Msg.Body > p.Balance {
					p.Logger.Error.Panicln("Cannot send app msg: not enough money!")
				}
				detMsg.From = p.NetIdx
				outBuf = p.Logger.GoVector.PrepareSend(fmt.Sprintf("Sending msg %s, content: $%d", detMsg.Msg.ID, detMsg.Msg.Body), detMsg, opts)
				if err := p.sendGroup(outBuf, &detMsg); err != nil {
					p.Logger.Error.Panicf("Cannot send app msg: %s", err)
				}
				p.UpdateBalance(detMsg.Msg.Body, "sent")
				responseCh <- utils.NewAppMsg("", -1, -1, -1)
				p.Logger.Info.Printf("MSG %s [Amount: %d] sent to: %s. Current budget: $%d\n", detMsg.Msg.ID, detMsg.Msg.Body, p.NetLayout.Nodes[detMsg.To].Name, p.Balance)
				p.SendMsgCh <- detMsg
			} else {
				p.Logger.Warning.Println("Cannot send app msg while node is performing global snapshot")
				p.SendAppMsgCh <- respMsg
			}
		case <-p.SendMarkCh:
			// Send mark
			mark := utils.NewMarkMsg(p.NetIdx)
			outBuf := p.Logger.GoVector.PrepareSend("Sending mark", mark, opts)
			err := p.sendGroup(outBuf, &mark)
			if err != nil {
				p.Logger.Error.Panicf("Cannot send initial mark: %s", err)
			}
		case state := <-p.CurrentStateCh:
			p.Mutex.Lock()
			p.FullState = state
			p.Mutex.Unlock()
			p.Logger.Info.Println("Node state updated")
			if state.AllMarksRecv {
				outBuf := p.Logger.GoVector.PrepareSend("Sending my state to all", state, opts)
				if err := p.sendGroup(outBuf, nil); err != nil {
					p.Logger.Error.Panicf("Cannot send app msg: %s", err)
				}
			}
		}
	}
}

// Sends req to the group
func (p *Process) sendGroup(data []byte, msg *utils.AppMessage) error {
	if msg == nil { // sending state
		for _, node := range p.NetLayout.Nodes {
			if node.Name != p.Info.Name {
				p.Logger.Info.Printf("Sending state to: %s\n", node.Name)
				go p.sendDirectMsg(data, node)
			}
		}
	} else { // sending mark
		if msg.IsMarker {
			for _, node := range p.NetLayout.Nodes {
				if node.Name != p.Info.Name {
					go p.sendDirectMsg(data, node)
				}
			}
		} else { // sending msg
			node := p.NetLayout.Nodes[msg.To]
			if node.Name != p.Info.Name {
				go p.sendDirectMsg(data, node)
			}
		}
	}
	return nil
}

func (p *Process) sendDirectMsg(msg []byte, node utils.Node) {
	var conn net.Conn
	var err error

	netAddr := fmt.Sprint(node.IP + ":" + strconv.Itoa(node.Port))
	conn, err = net.Dial("tcp", netAddr)
	for i := 0; err != nil && i < p.NetLayout.SendAttempts; i++ {
		p.Logger.Warning.Printf("Client connection error: %s", err)
		time.Sleep(Period)
		conn, err = net.Dial("tcp", netAddr)
	}
	if err != nil || conn == nil {
		p.Logger.Error.Panicf("Client connection error: %v", err)
	}
	_, err = conn.Write(msg)
	if err != nil {
		p.Logger.Error.Panicf("Sending data error: %v", err)
	}
	err = conn.Close()
	if err != nil {
		p.Logger.Error.Panicf("Closing connection error: %v", err)
	}
}

func (p *Process) UpdateBalance(amount int, status string) {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	switch status {
	case "sent":
		p.Balance -= amount
	case "received":
		p.Balance += amount
	}
}
