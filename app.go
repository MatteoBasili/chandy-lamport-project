package main

import (
	"fmt"
	"math/rand"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sdccProject/src/utils"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	nodeMainDir      = "src/main/"
	nodeAppName      = "node_app"
	sendMsgMethod    = "NodeApp.SendAppMsg"
	lowerBoundAmount = 1
	upperBoundAmount = 100
)

type Process struct {
	Id   int
	Name string
}

var globalMsgID uint64

func setupNetwork() ([]*Process, map[string]*rpc.Client) {
	var netLayout utils.NetLayout
	netLayout = utils.ReadConfig("net_config.json")
	if len(netLayout.Nodes) < 2 {
		panic("At least 2 processes are needed")
	}
	fmt.Printf("Net layout: %v\n", netLayout.Nodes)

	processes := make([]*Process, len(netLayout.Nodes))
	rpcConn := make(map[string]*rpc.Client)

	for idx, node := range netLayout.Nodes {
		// Initialize RPC node
		go utils.RunPromptCmd("go", "run", nodeMainDir+nodeAppName+".go", strconv.Itoa(idx), strconv.Itoa(node.AppPort))
		processes[idx] = &Process{Id: node.Idx, Name: node.Name}

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
		rpcConn[node.Name] = clientRPC

	}

	return processes, rpcConn
}

func runApp(processes []*Process, rpcConn map[string]*rpc.Client, stop chan struct{}) {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	respSnapCh := make(chan utils.GlobalState, 1)

	var wg sync.WaitGroup

	// Every second, each process transfers some fund to another process
	for _, p := range processes {
		wg.Add(1)
		go func(p *Process) {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				case <-time.After(1 * time.Second):
					msgID := generateMsgID()
					msg := utils.NewAppMsg("MS"+strconv.FormatUint(msgID, 10), genCasAmount(lowerBoundAmount, upperBoundAmount), p.Id, genCasProc(0, len(processes)-1, p.Id))
					utils.RunAPP(sendMsgMethod, rpcConn[p.Name], msg)
					fmt.Printf("Test: ordered MS%d\n", msgID)
				}
			}
		}(p)
	}

	// Process P0 takes a snapshot every two seconds
	go func() {
		idGs := 0
		for {
			select {
			case <-stop:
				return
			case <-time.After(2 * time.Second):
				idGs += 1
				utils.RunRPCSnapshot(rpcConn["P0"], respSnapCh)
				gs := <-respSnapCh
				fmt.Printf("Snapshot %d: %v\n", idGs, gs)
			}
		}
	}()

	wg.Wait()

}

func genCasAmount(min int, max int) int {
	randomInt := rand.Intn(max) + min
	return randomInt
}

func genCasProc(min int, max int, procIdx int) int {
	for {
		randomInt := rand.Intn(max) + min
		if randomInt != procIdx {
			return randomInt
		}
	}
}

func generateMsgID() uint64 {
	return atomic.AddUint64(&globalMsgID, 1)
}

// Connect and initialize RPC nodes
func main() {
	fmt.Println("Starting environment... ")
	processes, rpcConn := setupNetwork()
	fmt.Println("Starting the application...")

	stop := make(chan struct{})

	go func() {
		timer := time.NewTimer(10 * time.Second)
		<-timer.C
		close(stop)
	}()

	runApp(processes, rpcConn, stop)

	fmt.Println("Terminating the application...")
	time.Sleep(3 * time.Second)
	fmt.Println("Bye!")
	return
}

/*
func runApp() {
	nMsgs := 4
	respMsgCh := make(chan int, nMsgs)
	respSnapCh := make(chan utils.GlobalState, 1)

	rand.New(rand.NewSource(time.Now().UnixNano()))

	msg1 := utils.NewAppMsg("MS1", genCasAmount(lowerBoundAmount, upperBoundAmount), 0, 1)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P0"], msg1, 1, respMsgCh)
	fmt.Println("Test: ordered 1st msg")

	msg2 := utils.NewAppMsg("MS2", genCasAmount(lowerBoundAmount, upperBoundAmount), 2, 1)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P2"], msg2, 2, respMsgCh)
	fmt.Println("Test: ordered 2nd msg")

	time.Sleep(2 * time.Second)
	utils.RunRPCSnapshot(RPCConn["P0"], respSnapCh)
	fmt.Println("Test: ordered GS")

	msg3 := utils.NewAppMsg("MS3", genCasAmount(lowerBoundAmount, upperBoundAmount), 1, 0)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P1"], msg3, 3, respMsgCh)
	fmt.Println("Test: ordered 3rd msg")

	msg4 := utils.NewAppMsg("MS4", genCasAmount(lowerBoundAmount, upperBoundAmount), 1, 2)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P1"], msg4, 4, respMsgCh)
	fmt.Println("Test: ordered 4th msg")

	for i := 0; i < nMsgs; i++ {
		msgN := <-respMsgCh
		fmt.Printf("Msg nº: %d sent\n", msgN)
	}
	fmt.Println("All messages sent.")

	gs := <-respSnapCh
	fmt.Printf("Snapshot completed: %s\n", gs.String())

	msg5 := utils.NewAppMsg("MS5 - last", genCasAmount(lowerBoundAmount, upperBoundAmount), 0, 2)
	utils.RunRPCCommand(sendMsgMethod, RPCConn["P0"], msg5, 5, respMsgCh)
	fmt.Println("Test: ordered 5th msg")

	time.Sleep(2 * time.Second)
	fmt.Println("Test: ordered last GS")
	utils.RunRPCSnapshot(RPCConn["P1"], respSnapCh)
	gs = <-respSnapCh
	fmt.Printf("Snapshot completed: %s\n", gs.String())
}
*/
