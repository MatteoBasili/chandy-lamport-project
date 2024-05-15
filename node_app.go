package main

import (
	"encoding/gob"
	"fmt"
	"github.com/DistributedClocks/GoVector/govec"
	"github.com/DistributedClocks/GoVector/govec/vrpc"
	"net"
	"net/rpc"
	"os"
	"sdccProject/src/process"
	"sdccProject/src/utils"
	"strconv"
)

type NodeApp struct {
	node         *process.Process
	netLayout    utils.NetLayout
	sendAppMsgCh chan utils.AppMsgWithResp
	recvAppMsgCh chan utils.AppMessage
	log          *utils.Logger
}

func NewNodeApp(idxNet int) *NodeApp {
	var nodeApp NodeApp

	// Read Network Layout
	var network utils.NetLayout
	network = utils.ReadConfig()
	if len(network.Nodes) < idxNet+1 {
		panic("At least " + strconv.Itoa(idxNet+1) + " processes are needed")
	}
	nodeApp.netLayout = network

	nodeApp.sendAppMsgCh = make(chan utils.AppMsgWithResp, 10) // node <--    msg   --- app
	nodeApp.recvAppMsgCh = make(chan utils.AppMessage, 10)     // node ---    msg   --> app

	// Register struct
	gob.Register(utils.AppMessage{})
	nodeApp.log = utils.InitLoggers(strconv.Itoa(idxNet))
	nodeApp.node = process.NewProcess(idxNet, network, nodeApp.sendAppMsgCh, nodeApp.recvAppMsgCh, nodeApp.log)
	return &nodeApp
}

func (a *NodeApp) recvAppMsg() {
	for {
		appMsg := <-a.recvAppMsgCh
		a.log.Info.Printf("MSG %s [Amount: %d] received from: %s. Current budget: $%d\n", appMsg.ID, appMsg.Body, a.netLayout.Nodes[appMsg.From].Name, a.node.Balance)
	}
}

func (a *NodeApp) SendAppMsg(rq *utils.AppMessage, resp *interface{}) error {
	responseCh := make(chan utils.AppMessage)
	a.log.Info.Printf("Sending MSG %s [Amount: %d] to: %s...\n", rq.ID, rq.Body, a.netLayout.Nodes[rq.To].Name)
	a.sendAppMsgCh <- utils.AppMsgWithResp{Msg: *rq, RespCh: responseCh}
	_ = <-responseCh
	a.log.Info.Printf("MSG %s [Amount: %d] sent to: %s. Current budget: $%d\n", rq.ID, rq.Body, a.netLayout.Nodes[rq.To].Name, a.node.Balance)
	return nil
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
		panic(fmt.Sprintf("Bad argument[1]: %s. Error: %s. Usage: go run node_app.go <0-based node index> <RPC port>", args[1], err))
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
	rpc.HandleHTTP()

	l, err = net.Listen("tcp", ":"+args[1])
	if err != nil {
		panic(err)
	}
	options := govec.GetDefaultLogOptions()
	vrpc.ServeRPCConn(server, l, myNodeApp.log.GoVector, options)
	return
}
