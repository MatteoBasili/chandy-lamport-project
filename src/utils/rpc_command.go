package utils

import (
	"net/rpc"
)

// Result represents the result of the RPC call "SendAppMsg"
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
		if chResp != nil {
			chResp <- resp
		}
	}()
}
