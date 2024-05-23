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
	node     *process.Process
	snap     *snapshot.SnapNode
	appMsgCh utils.AppMsgChannels
	log      *utils.Logger
}

func NewNodeApp(netIdx int, configFile string) *NodeApp {
	var nodeApp NodeApp

	// Read Network Layout
	var network utils.NetLayout
	network = utils.ReadConfig(configFile)
	if len(network.Nodes) < netIdx+1 {
		panic("At least " + strconv.Itoa(netIdx+1) + " processes are needed")
	}

	// Create channels
	nodeApp.appMsgCh = utils.CreateAppMsgChans(10)
	markCh := utils.CreateMarkChans(10)
	statesCh := utils.CreateStatesChans(10)

	nodeApp.log = utils.InitLoggers(strconv.Itoa(netIdx))
	nodeApp.node = process.NewProcess(netIdx, nodeApp.appMsgCh, statesCh, markCh, network, nodeApp.log)
	nodeApp.snap = snapshot.NewSnapNode(netIdx, nodeApp.appMsgCh.SendToSnapCh, statesCh, markCh, &network, nodeApp.log)
	return &nodeApp
}

func (a *NodeApp) MakeSnapshot(_ *interface{}, resp *utils.GlobalState) error {
	*resp = a.snap.MakeSnapshot()
	a.log.Info.Printf("Received global state: %s\n", resp.String())
	return nil
}

func (a *NodeApp) SendAppMsg(rq *utils.AppMessage, resp *utils.Result) error {
	responseCh := make(chan utils.AppMessage)
	a.log.Info.Printf("Sending message %s [Amount: %d] to: %s...\n", rq.Msg.ID, rq.Msg.Body, a.node.NetLayout.Nodes[rq.To].Name)
	a.appMsgCh.SendToProcCh <- utils.RespMessage{AppMsg: *rq, RespCh: responseCh}
	res := <-responseCh
	if res.To != -1 {
		time.Sleep(1 * time.Second)
		_ = a.SendAppMsg(rq, resp)
	}
	return nil
}

func (a *NodeApp) recvAppMsg() {
	for {
		appMsg := <-a.appMsgCh.RecvCh
		a.log.Info.Printf("Message %s [Amount: %d] received from: %s. Current budget: $%d\n", appMsg.Msg.ID, appMsg.Msg.Body, a.node.NetLayout.Nodes[appMsg.From].Name, a.node.Balance)
	}
}

func main() {
	args := os.Args[1:]
	var err error
	var netIdx int
	var l net.Listener

	if len(args) != 3 {
		panic("Incorrect number of arguments. Usage: go run node_app.go <0-based node index> <node app RPC port> <net config filename>")
	}

	netIdx, err = strconv.Atoi(args[0])
	if err != nil {
		panic(fmt.Sprintf("Bad argument[0]: %s. Error: %s. Usage: go run node_app.go <0-based node index> <node app RPC port> <net config filename>", args[0], err))
	}
	_, err = strconv.Atoi(args[1])
	if err != nil {
		panic(fmt.Sprintf("Bad argument[1]: %s. Error: %s. Usage: go run node_app.go <0-based node index> <node app RPC port> <net config filename>", args[1], err))
	}

	fmt.Printf("Starting P%d...\n", netIdx)
	time.Sleep(1 * time.Second)
	myNodeApp := NewNodeApp(netIdx, args[2])
	time.Sleep(2 * time.Second)
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
