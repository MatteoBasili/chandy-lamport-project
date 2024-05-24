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

func ReadConfig(file string) NetLayout {
	// read file
	data, err := os.ReadFile(file)
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

func PrintNetwork(layout NetLayout) string {
	var res string
	for _, node := range layout.Nodes {
		res += "—————————————————————————————————— "
		res += fmt.Sprintf("Process %s:\n", node.Name)
		res += fmt.Sprintf("- Index: %d\n", node.Idx)
		res += fmt.Sprintf("- IP: %s\n", node.IP)
		res += fmt.Sprintf("- Port: %d\n", node.Port)
		res += fmt.Sprintf("- App port: %d\n", node.AppPort)
		res += fmt.Sprintf("- Initial Balance: $%d\n", layout.InitialBalance)
	}
	res += "**********************************************\n"

	return res
}
