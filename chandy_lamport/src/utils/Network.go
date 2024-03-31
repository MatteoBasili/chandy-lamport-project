package utils

type Node struct {
	ID    int `json:"ID"`
	Address      string `json:"Address"`
	Port    int    `json:"Port"`
}

type NetLayout struct {
	Nodes        []Node `json:"Nodes"`
	Balance float64 `json:"Balance"`
}