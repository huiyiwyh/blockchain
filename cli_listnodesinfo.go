package main

import (
	"fmt"
)

func (cli *CLI) listNodesInfo() {
	nodesInfo := NewNodesInfo()

	fmt.Println("-----------Nodes----------")
	for node, nodeInfo := range nodesInfo {
		fmt.Println("=================================")
		fmt.Println(node)
		fmt.Println("---------------------------------")
		fmt.Println(MapToSlice(nodeInfo.Nodes))
	}

	fmt.Println("-----------Wallets----------")
	for node, nodeInfo := range nodesInfo {
		fmt.Println("=================================")
		fmt.Println(node)
		fmt.Println("---------------------------------")
		fmt.Println(MapToSlice(nodeInfo.Wallets))
	}
}

func MapToSlice(oldMap map[string]string) []string {
	slice := []string{}

	for _, v := range oldMap {
		slice = append(slice, v)
	}

	return slice
}
