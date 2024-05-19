package main

import (
	"fmt"
	"math/rand"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sdccProject/src/utils"
	"strconv"
	"time"
)

const (
	nodeMainDir   = "src/main/"
	nodeAppName   = "node_app"
	sendMsgMethod = "NodeApp.SendAppMsg"
	lowerBound    = 0
	upperBound    = 100
)

var RPCConn map[string]*rpc.Client

// Connect and initialize RPC nodes
func main() {
	//time.Sleep(1 * time.Second)
	fmt.Println("Starting environment... ")
	//time.Sleep(2 * time.Second)
	setupNetwork()
	fmt.Println("Starting application...")
	//time.Sleep(3 * time.Second)
	runApp()
	terminate()
	return
}

func setupNetwork() {
	var netLayout utils.NetLayout
	netLayout = utils.ReadConfig("net_config.json")
	if len(netLayout.Nodes) < 2 {
		panic("At least 2 processes are needed")
	}
	fmt.Printf("Net layout: %v\n", netLayout.Nodes)

	RPCConn = make(map[string]*rpc.Client)

	for idx, node := range netLayout.Nodes {
		// Initialize RPC node
		go utils.RunPromptCmd("go", "run", nodeMainDir+nodeAppName+".go", strconv.Itoa(idx), strconv.Itoa(node.AppPort))

		// Connect via RPC to the server
		var clientRPC *rpc.Client
		var err error
		for i := 0; i < netLayout.SendAttempts; i++ {
			time.Sleep(3 * time.Second) // Wait for RPC initialization
			clientRPC, err = jsonrpc.Dial("tcp", node.IP+":"+strconv.Itoa(node.AppPort))
			if err == nil {
				break
			}
		}
		if err != nil {
			panic(err)
		}
		RPCConn[node.Name] = clientRPC

	}
}

func terminate() {
	fmt.Println("Done! Closing connections...")
	for _, conn := range RPCConn {
		_ = conn.Close()
	}
	fmt.Println("Connections terminated")
	fmt.Println("Terminating all processes...")
	utils.RunPromptCmd("taskkill", "/F", "/IM", nodeAppName+".exe")
}

func genCasNum(min int, max int) int {
	randomInt := rand.Intn(max-min+1) + min
	return randomInt
}

func runApp() {
	nMsgs := 4
	respMsgCh := make(chan int, nMsgs)
	respSnapCh := make(chan utils.GlobalState, 1)

	rand.New(rand.NewSource(time.Now().UnixNano()))

	msg1 := utils.NewAppMsg("MS1", genCasNum(lowerBound, upperBound), 0, 1)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P0"], msg1, 1, respMsgCh)
	fmt.Println("Test: ordered 1st msg")

	msg2 := utils.NewAppMsg("MS2", genCasNum(lowerBound, upperBound), 2, 1)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P2"], msg2, 2, respMsgCh)
	fmt.Println("Test: ordered 2nd msg")

	time.Sleep(2 * time.Second)
	utils.RunRPCSnapshot(RPCConn["P0"], respSnapCh)
	fmt.Println("Test: ordered GS")

	msg3 := utils.NewAppMsg("MS3", genCasNum(lowerBound, upperBound), 1, 0)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P1"], msg3, 3, respMsgCh)
	fmt.Println("Test: ordered 3rd msg")

	msg4 := utils.NewAppMsg("MS4", genCasNum(lowerBound, upperBound), 1, 2)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P1"], msg4, 4, respMsgCh)
	fmt.Println("Test: ordered 4th msg")

	for i := 0; i < nMsgs; i++ {
		msgN := <-respMsgCh
		fmt.Printf("Msg nº: %d sent\n", msgN)
	}
	fmt.Println("All messages sent.")

	gs := <-respSnapCh
	fmt.Printf("Snapshot completed: %v\n", gs)

	msg5 := utils.NewAppMsg("MS5 - last", genCasNum(lowerBound, upperBound), 0, 2)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P0"], msg5, 5, respMsgCh)
	fmt.Println("Test: ordered 5th msg")

	time.Sleep(2 * time.Second)
	fmt.Println("Test: ordered last GS")
	utils.RunRPCSnapshot(RPCConn["P1"], respSnapCh)
	gs = <-respSnapCh
	fmt.Printf("Snapshot completed: %v\n", gs)
}
