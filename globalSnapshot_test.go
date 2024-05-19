package main_test

import (
	"fmt"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
	"sdccProject/src/utils"
	"strconv"
	"testing"
	"time"
)

const (
	nodeMainDir      = "src/main/"
	nodeAppName      = "node_app"
	sendMsgMethod    = "NodeApp.SendAppMsg"
	lowerBoundAmount = 1
	upperBoundAmount = 100
)

var tempConfigFile string
var RPCConn map[string]*rpc.Client

// Connect and initialize RPC nodes
func TestMain(m *testing.M) {
	fmt.Println("Starting tests for Global Snapshot...")
	setupNetwork()
	fmt.Println("Execute the rest of the tests...")
	m.Run()
	fmt.Println("Global Snapshot tests finished. Closing...")
	time.Sleep(1 * time.Second)
	terminate()
}

func setupNetwork() {
	// Before running the test, create a temporary configuration file
	// with known data to be read by the ReadConfig function
	testConfigData := `{
        "nodes": [
            {
                "idx": 0,
                "name": "P0",
                "ip": "127.0.0.1",
                "port": 16090,
                "appPort": 26090
            },
            {
                "idx": 1,
                "name": "P1",
                "ip": "127.0.0.1",
                "port": 16091,
                "appPort": 26091
            },
            {
                "idx": 2,
                "name": "P2",
                "ip": "127.0.0.1",
                "port": 16092,
                "appPort": 26092
            }
        ],
        "initialBalance": 3000,
        "sendAttempts": 10
    }`

	// Write the data to the temporary file for the test
	tempConfigFile = "temp_net_config.json"
	err := os.WriteFile(tempConfigFile, []byte(testConfigData), 0644)
	if err != nil {
		panic("Error creating temporary config file")
	}

	var netLayout utils.NetLayout
	netLayout = utils.ReadConfig(tempConfigFile)
	if len(netLayout.Nodes) < 2 {
		panic("At least 2 processes are needed")
	}
	fmt.Printf("Net layout: %v\n", netLayout.Nodes)

	RPCConn = make(map[string]*rpc.Client)

	for idx, node := range netLayout.Nodes {
		// Initialize RPC node
		go utils.RunPromptCmd("go", "run", nodeMainDir+nodeAppName+".go", strconv.Itoa(idx), strconv.Itoa(node.AppPort), tempConfigFile)

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
	for _, conn := range RPCConn {
		_ = conn.Close()
	}
	fmt.Println("Connections terminated")
	time.Sleep(1 * time.Second)

	// Remove the temporary file
	err := os.Remove(tempConfigFile)
	if err != nil {
		panic("Error removing temporary config file")
	}
	fmt.Println("Temporary config file removed")

	// Terminate processes
	utils.RunPromptCmd("taskkill", "/IM", nodeAppName+".exe", "/F")
}

func TestEmptySnapshot(t *testing.T) {
	fmt.Println("Testing empty snapshot...")
	respSnapCh := make(chan utils.GlobalState, 1)
	defer close(respSnapCh)
	process := "P1"
	utils.RunRPCSnapshot(RPCConn[process], respSnapCh)
	gs := <-respSnapCh
	fmt.Printf("Empty Snapshot completed from process %s\n", process)

	for i := 0; i < 3; i++ {
		if gs.GS[i].Node.Balance != 3000 {
			t.Errorf("Expected %d, but got %d", 3000, gs.GS[i].Node.Balance)
		}
		if len(gs.GS[i].Node.SentMsgs) != 0 {
			t.Errorf("Expected %d, but got %q", 0, len(gs.GS[i].Node.SentMsgs))
		}
		if len(gs.GS[i].Node.ReceivedMsgs) != 0 {
			t.Errorf("Expected %d, but got %d", 0, len(gs.GS[i].Node.ReceivedMsgs))
		}
		for j := range gs.GS[i].Channels {
			if len(gs.GS[i].Channels[j].RecvMsgs) != 0 {
				t.Errorf("Expected %d, but got %d", 0, len(gs.GS[i].Channels))
			}
		}
		if gs.GS[i].AllMarksRecv != true {
			t.Errorf("Expected %t, but got %t", true, gs.GS[i].AllMarksRecv)
		}
	}
}

/*
func TestSameTimeSnapshot(t *testing.T) {
	nMsgs := 4
	respMsgCh := make(chan int, nMsgs)
	respSnapCh := make(chan utils.GlobalState, 1)

	msg1 := utils.NewAppMsg("MS1", 55, 0, 1)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P0"], msg1, 1, respMsgCh)
	fmt.Println("Test: ordered 1st msg")

	msg2 := utils.NewAppMsg("MS2", 23, 2, 1)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P2"], msg2, 2, respMsgCh)
	fmt.Println("Test: ordered 2nd msg")

	time.Sleep(2 * time.Second)
	utils.RunRPCSnapshot(RPCConn["P0"], respSnapCh)
	fmt.Println("Test: ordered GS")

	msg3 := utils.NewAppMsg("MS3", 98, 1, 0)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P1"], msg3, 3, respMsgCh)
	fmt.Println("Test: ordered 3rd msg")

	msg4 := utils.NewAppMsg("MS4", 45, 1, 2)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P1"], msg4, 4, respMsgCh)
	fmt.Println("Test: ordered 4th msg")

	for i := 0; i < nMsgs; i++ {
		msgN := <-respMsgCh
		fmt.Printf("Msg nº: %d sent\n", msgN)
	}
	fmt.Println("All messages sent.")

	gs := <-respSnapCh
	fmt.Printf("Snapshot completed: %s\n", gs.String())

	msg5 := utils.NewAppMsg("MS5 - last", genRandAmount(lowerBoundAmount, upperBoundAmount), 0, 2)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P0"], msg5, 5, respMsgCh)
	fmt.Println("Test: ordered 5th msg")

	time.Sleep(2 * time.Second)
	fmt.Println("Test: ordered last GS")
	utils.RunRPCSnapshot(RPCConn["P1"], respSnapCh)
	gs = <-respSnapCh
	fmt.Printf("Snapshot completed: %s\n", gs.String())
}
*/
