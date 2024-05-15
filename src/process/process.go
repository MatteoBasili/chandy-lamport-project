package process

import (
	"fmt"
	"github.com/DistributedClocks/GoVector/govec"
	"net"
	"sdccProject/src/utils"
	"strconv"
	"sync"
	"time"
)

const (
	Period = 2000 * time.Millisecond
)

type Process struct {
	Info         utils.Node
	NetLayout    utils.NetLayout
	NetIdx       int
	Balance      int
	FullState    utils.FullState
	Listener     net.Listener
	SendAppMsgCh chan utils.AppMsgWithResp
	RecvAppMsgCh chan utils.AppMessage
	Logger       *utils.Logger
	Mutex        sync.Mutex
}

func NewProcess(netIdx int, netLayout utils.NetLayout, sendAppMsgCh chan utils.AppMsgWithResp, recvAppMsgCh chan utils.AppMessage, logger *utils.Logger) *Process {
	var myNode = netLayout.Nodes[netIdx]

	// Create channels

	// Open Listener port
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(myNode.Port))
	if err != nil {
		panic(fmt.Sprintf("ERROR: unable to open port: %s. Error: %s.", strconv.Itoa(myNode.Port), err))
	}

	var myProcess = Process{
		Info:         myNode,
		NetLayout:    netLayout,
		NetIdx:       netIdx,
		Balance:      netLayout.InitialBalance,
		FullState:    utils.FullState{},
		Listener:     listener,
		SendAppMsgCh: sendAppMsgCh,
		RecvAppMsgCh: recvAppMsgCh,
		Logger:       logger,
		Mutex:        sync.Mutex{},
	}
	myProcess.Logger.Trace.Printf("Listening on port: %s. Initial balance : $%d", strconv.Itoa(myNode.Port), myProcess.Balance)

	go myProcess.sender()
	go myProcess.recvAppMsg()
	return &myProcess
}

func (p *Process) recvAppMsg() *utils.AppMessage {
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
		var recvMsg utils.AppMessage
		p.Logger.GoVector.UnpackReceive("Receiving App Message", recvData[0:nBytes], &recvMsg, govec.GetDefaultLogOptions())
		p.Logger.GoVector.LogLocalEvent(fmt.Sprintf("MSG %s, content: $%d, from [%s]", recvMsg.ID, recvMsg.Body, p.NetLayout.Nodes[recvMsg.From].Name), govec.GetDefaultLogOptions())
		p.UpdateBalance(recvMsg.Body, "received")
		p.Logger.Info.Printf("MSG %s [Amount: %d] received from: %s\n", recvMsg.ID, recvMsg.Body, p.NetLayout.Nodes[recvMsg.From].Name)
		p.RecvAppMsgCh <- recvMsg
	}
}

func (p *Process) sender() {
	opts := govec.GetDefaultLogOptions()
	var outBuf []byte
	outBuf = []byte{'A', 'B'}
	for {
		select {
		case msgWithResp := <-p.SendAppMsgCh:
			detMsg := msgWithResp.Msg
			responseCh := msgWithResp.RespCh

			if detMsg.Body > p.Balance {
				p.Logger.Error.Panicln("Cannot send app msg: not enough money!")
			}
			detMsg.From = p.NetIdx
			outBuf = p.Logger.GoVector.PrepareSend(fmt.Sprintf("Sending msg %s, content: $%d", detMsg.ID, detMsg.Body), detMsg, opts)
			if err := p.sendGroup(outBuf, &detMsg); err != nil {
				p.Logger.Error.Panicf("Cannot send app msg: %s", err)
			}
			p.UpdateBalance(detMsg.Body, "sent")
			responseCh <- utils.AppMessage{}
		}
	}
}

// Sends req to the group
func (p *Process) sendGroup(data []byte, appMsg *utils.AppMessage) error {
	node := p.NetLayout.Nodes[appMsg.To]
	if node.Name != p.Info.Name {
		go p.sendDirectMsg(data, node)
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
