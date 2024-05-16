package utils

import (
	"log"
	"net/rpc"
	"os"
	"os/exec"
)

func RunRPCSnapshot(conn *rpc.Client, chResp chan GlobalState) {
	go func() {
		var gs GlobalState
		call := conn.Go("NodeApp.MakeSnapshot", nil, &gs, nil)
		call = <-call.Done
		if call.Error != nil {
			panic(call.Error.Error())
		}
		chResp <- gs
	}()
}

func RunRPCCommand(method string, conn *rpc.Client, content interface{}, resp int, chResp chan int) {
	call := conn.Go(method, &content, nil, nil)
	call = <-call.Done
	if call.Error != nil {
		panic(call.Error.Error())
	}
	chResp <- resp
}

func RunCommand(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
