package main

import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"sdccProject/src/process"
	"sdccProject/src/snapshot"
	"sdccProject/src/utils"
	"strconv"
	"time"
)

type NodeApp struct {
	node *process.Process
	snap *snapshot.SnapNode
	log  *utils.Logger
}

func NewNodeApp(netIdx int) *NodeApp {
	var nodeApp NodeApp

	// Read Network Layout
	var network utils.NetLayout
	network = utils.ReadConfig()
	if len(network.Nodes) < netIdx+1 {
		panic("At least " + strconv.Itoa(netIdx+1) + " processes are needed")
	}

	// Create channels
	currentStateCh := make(chan utils.FullState, 10) // node <-- FullState --- snap
	recvStateCh := make(chan utils.FullState, 10)    // node --- FullState --> snap
	sendMarkCh := make(chan utils.AppMessage, 10)    // node <-- SendMark --> snap
	recvMarkCh := make(chan utils.AppMessage, 10)    // node --- mark|msg --> snap
	sendMsgCh := make(chan utils.AppMessage, 10)     // node <-- SendMark --> snap

	nodeApp.log = utils.InitLoggers(strconv.Itoa(netIdx))
	nodeApp.node = process.NewProcess(netIdx, currentStateCh, recvStateCh, sendMarkCh, recvMarkCh, sendMsgCh, network, nodeApp.log)
	nodeApp.snap = snapshot.NewSnapNode(netIdx, currentStateCh, recvStateCh, sendMarkCh, recvMarkCh, sendMsgCh, &network, nodeApp.log)
	return &nodeApp
}

func (a *NodeApp) MakeSnapshot(_ *interface{}, resp *utils.GlobalState) error {
	*resp = a.snap.MakeSnapshot()
	a.log.Info.Printf("Received global state: %v\n", resp)
	a.log.Info.Printf("Balance: $%d\n", resp.GS[a.node.Info.Idx].Node.Balance)
	return nil
}

func (a *NodeApp) SendAppMsg(rq *utils.AppMessage, resp *utils.Result) error {
	responseCh := make(chan utils.AppMessage)
	a.log.Info.Printf("Sending MSG %s [Amount: %d] to: %s...\n", rq.Msg.ID, rq.Msg.Body, a.node.NetLayout.Nodes[rq.To].Name)
	a.node.SendAppMsgCh <- utils.RespMessage{AppMsg: *rq, RespCh: responseCh}
	res := <-responseCh
	if res.To != -1 {
		time.Sleep(1 * time.Second)
		_ = a.SendAppMsg(rq, resp)
	}
	return nil
}

func (a *NodeApp) recvAppMsg() {
	for {
		appMsg := <-a.node.RecvAppMsgCh
		a.log.Info.Printf("MSG %s [Amount: %d] received from: %s. Current budget: $%d\n", appMsg.Msg.ID, appMsg.Msg.Body, a.node.NetLayout.Nodes[appMsg.From].Name, a.node.Balance)
	}
}

func main() {
	args := os.Args[1:]
	var err error
	var netIdx int
	var l net.Listener

	if len(args) != 2 {
		panic("Incorrect number of arguments. Usage: go run node_app.go <0-based node index> <node app RPC port>")
	}

	netIdx, err = strconv.Atoi(args[0])
	if err != nil {
		panic(fmt.Sprintf("Bad argument[0]: %s. Error: %s. Usage: go run node_app.go <0-based node index> <node app RPC port>", args[0], err))
	}
	_, err = strconv.Atoi(args[1])
	if err != nil {
		panic(fmt.Sprintf("Bad argument[1]: %s. Error: %s. Usage: go run node_app.go <0-based node index> <node app RPC port>", args[1], err))
	}

	fmt.Printf("Starting P%d...\n", netIdx)
	myNodeApp := NewNodeApp(netIdx)
	fmt.Printf("Process P%d is ready\n", netIdx)
	go myNodeApp.recvAppMsg()

	// Register node app as RPC
	server := rpc.NewServer()
	err = server.Register(myNodeApp)
	if err != nil {
		panic(err)
	}

	l, err = net.Listen("tcp", ":"+args[1])
	if err != nil {
		panic(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
	return
}
