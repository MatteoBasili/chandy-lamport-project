package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

type Node struct {
	Idx     int    `json:"idx"`
	Name    string `json:"name"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	AppPort int    `json:"appPort"`
}

type NetLayout struct {
	Nodes          []Node `json:"nodes"`
	InitialBalance int    `json:"initialBalance"`
	SendAttempts   int    `json:"sendAttempts"`
}

func ReadConfig() NetLayout {
	// read file
	data, err := os.ReadFile("net_config.json")
	if err != nil {
		panic(err)
	}

	fmt.Println("Reading network configuration file...")

	var netCfg NetLayout
	// parse content of json file to Config struct
	err = json.Unmarshal(data, &netCfg)
	if err != nil {
		panic(err)
	}

	return netCfg
}
