package main

import (
	"log"

	node_worker "github.com/ttimmatti/nodes-bot/sui/worker"
)

func main() {
	v, err := node_worker.GetNetworkVersion()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(v)
	id, err := node_worker.GetNetworkTxId()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(id)
}
