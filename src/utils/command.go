package utils

import (
	"log"
	"net/rpc"
	"os"
	"os/exec"
)

// Result represents the result of the RPC call
type Result struct {
	Result int
}

func RunRPCSnapshot(conn *rpc.Client, chResp chan GlobalState) {
	go func() {
		var gs GlobalState
		err := conn.Call("NodeApp.MakeSnapshot", nil, &gs)
		if err != nil {
			panic(err)
		}
		chResp <- gs
	}()
}

func RunRPCCommand(method string, conn *rpc.Client, content interface{}, resp int, chResp chan int) {
	go func() {
		err := conn.Call(method, &content, nil)
		if err != nil {
			panic(err)
		}
		if chResp == nil {
			chResp <- resp
		}
	}()
}

func RunPromptCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
