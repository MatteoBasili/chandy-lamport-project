package utils

import (
	"log"
	"net/rpc"
	"os"
	"os/exec"
)

func RunCommand(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func RunRPCCommand(method string, conn *rpc.Client, content interface{}, resp int, chResp chan int) {
	call := conn.Go(method, &content, nil, nil)
	call = <-call.Done
	if call.Error != nil {
		panic(call.Error.Error())
	}
	chResp <- resp
}
