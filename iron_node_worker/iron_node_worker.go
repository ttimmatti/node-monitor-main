package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/ttimmatti/nodes-bot/ironfish/db"
	"github.com/ttimmatti/nodes-bot/ironfish/env"
	node_worker "github.com/ttimmatti/nodes-bot/ironfish/worker"
)

const WD = "/home/ttimmatti/my_scripts/go/ironfish_checker_tg_client/iron_node_worker/"

func main() {
	err := godotenv.Load(WD + ".env")
	if err != nil {
		log.Fatalln("Couldnt get Environment")
	}

	db.OnExit()

	db.DB = db.SetConn(env.GetDbEnv())

	tg_api := env.GetTGApiEnv()
	git_token := env.GetGitTokenEnv()
	admin_id := env.GetAdminIdEnv()

	node_worker.ADMIN_ID = admin_id
	node_worker.GITHUB_TOKEN = git_token
	node_worker.TG_API = tg_api
	node_worker.StartMBS()
}
