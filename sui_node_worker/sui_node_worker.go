package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/ttimmatti/nodes-bot/sui/db"
	"github.com/ttimmatti/nodes-bot/sui/env"
	node_worker "github.com/ttimmatti/nodes-bot/sui/worker"
)

const WD = "/home/ttimmatti/my_scripts/go/ironfish_checker_tg_client/sui_node_worker/"

func main() {
	err := godotenv.Load(WD + ".env")
	if err != nil {
		log.Fatalln("Couldnt get Environment")
	}

	db.OnExit()

	db.DB = db.SetConn(env.GetDbEnv())

	tg_api := env.GetTGApiEnv()

	// server with last tx count
	go node_worker.StartListening()

	node_worker.TG_API = tg_api
	node_worker.Start()
}
