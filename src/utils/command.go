package utils

import (
	"github.com/vmihailenco/msgpack/v5"
	"log"
	"net/rpc"
	"os"
	"os/exec"
)

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
		data, err := msgpack.Marshal(content)
		if err != nil {
			//return nil, err
		}

		err = conn.Call(method, data, nil)
		if err != nil {
			panic(err)
		}
		chResp <- resp
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
