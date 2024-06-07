package main

import (
	"fmt"
	"math/rand"
	"net/rpc"
	"net/rpc/jsonrpc"
	"chandy_lamport/src/utils"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	configFileName   = "net_config.json"
	sendMsgMethod    = "NodeApp.SendAppMsg"
	lowerBoundAmount = 1
	upperBoundAmount = 100
	execTime         = 10 // Seconds
)

type Process struct {
	Id   int
	Name string
}

var globalMsgID uint64

func setupNetwork() ([]*Process, map[string]*rpc.Client) {
	var netLayout utils.NetLayout
	netLayout = utils.ReadConfig(configFileName)
	if len(netLayout.Nodes) < 2 {
		panic("At least 2 processes are needed")
	}
	time.Sleep(2 * time.Second)
	fmt.Printf("Net layout:\n%s\n", utils.PrintNetwork(netLayout))
	time.Sleep(1 * time.Second)

	processes := make([]*Process, len(netLayout.Nodes))
	rpcConn := make(map[string]*rpc.Client)

	fmt.Println("Connecting to network...")
	for idx, node := range netLayout.Nodes {
		// Initialize RPC main
		processes[idx] = &Process{Id: node.Idx, Name: node.Name}

		// Connect via RPC to the server
		var clientRPC *rpc.Client
		var err error
		for i := 0; i < netLayout.SendAttempts; i++ {
			time.Sleep(3 * time.Second) // Wait for RPC initialization
			clientRPC, err = jsonrpc.Dial("tcp", node.Name+":"+strconv.Itoa(node.AppPort))
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
	var wg sync.WaitGroup

	// Every second, each process transfers some fund to another process
	fmt.Println("The exchange of messages begins")
	for _, p := range processes {
		wg.Add(1)
		go func(p *Process) {
			defer wg.Done()
			for {
				select {
				case <-stop:
					fmt.Println("Process ", p.Name, " received stop signal")
					return
				case <-time.After(1 * time.Second):
					msgID := generateMsgID()
					amount := genRandAmount(lowerBoundAmount, upperBoundAmount)
					endProcess := genRandProc(0, len(processes)-1, p.Id)
					msg := utils.NewAppMsg("MSG"+strconv.FormatUint(msgID, 10), amount, p.Id, endProcess)
					utils.RunRPCCommand(sendMsgMethod, rpcConn[p.Name], msg, -1, nil)
					fmt.Printf("App: ordered MSG%d (%s --- $%d ---> P%d)\n", msgID, p.Name, amount, endProcess)
				}
			}
		}(p)
	}

	// Random process takes a snapshot every two seconds
	wg.Add(1)
	respSnapCh := make(chan utils.GlobalState, 1)
	go func() {
		defer close(respSnapCh)
		defer wg.Done()
		process := getRandomKey(rpcConn)
		fmt.Printf("Process %s begins taking snapshots\n", process)
		idGs := 0
		for {
			select {
			case <-stop:
				return
			case <-time.After(2 * time.Second):
				idGs += 1
				utils.RunRPCSnapshot(rpcConn[process], respSnapCh)
				gs := <-respSnapCh
				fmt.Println()
				fmt.Println("##################################################")
				fmt.Printf("* Snapshot %d: %v", idGs, gs)
				fmt.Println("##################################################")
				fmt.Println()
			}
		}
	}()

	wg.Wait()
}

func getRandomKey(m map[string]*rpc.Client) string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	randomIndex := rand.Intn(len(keys))
	return keys[randomIndex]
}

func genRandAmount(min int, max int) int {
	randomInt := rand.Intn(max-min+1) + min
	return randomInt
}

func genRandProc(min int, max int, procIdx int) int {
	for {
		randomInt := rand.Intn(max-min+1) + min
		if randomInt != procIdx {
			return randomInt
		}
	}
}

func generateMsgID() uint64 {
	return atomic.AddUint64(&globalMsgID, 1)
}

func main() {
	time.Sleep(1 * time.Second)
	fmt.Println("Starting environment...")
	time.Sleep(2 * time.Second)
	processes, rpcConn := setupNetwork()

	rand.New(rand.NewSource(time.Now().UnixNano()))

	fmt.Println("Starting the application...")
	time.Sleep(3 * time.Second)

	stop := make(chan struct{})

	go func() {
		timer := time.NewTimer(execTime * time.Second)
		<-timer.C
		close(stop)
	}()

	runApp(processes, rpcConn, stop)
	time.Sleep(1 * time.Second)

	fmt.Println()
	fmt.Println("Terminating the application...")
	time.Sleep(3 * time.Second)
	fmt.Println("Bye!")

	return
}
