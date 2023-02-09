package main

import (
	"log"

	"github.com/joho/godotenv"
	db "github.com/ttimmatti/nodes-bot/tg-db/db"
	"github.com/ttimmatti/nodes-bot/tg-db/env"
	msgs "github.com/ttimmatti/nodes-bot/tg-db/tg-msgs"
)

const WD = "/home/ttimmatti/my_scripts/go/ironfish_checker_tg_client/tg-db/"

const UPDATE_FREQ = 2 // every $int seconds

var LAST_MSG_INDEX int64

func main() {
	err := godotenv.Load(WD + ".env")
	if err != nil {
		log.Fatalln("Couldnt get Environment")
	}

	db.OnExit()

	db.DB = db.SetConn(env.GetDbEnv())
	msgs.ADMIN_ID = env.GetAdminIdEnv()

	tg_api := env.GetTGApiEnv()

	msgs.NODES_CHAT = env.GetNodesChat()

	log.Println("Start receiving")
	msgs.StartReceiving(tg_api, UPDATE_FREQ)

	//TODO: make each sendMsg try twice

	// rules for error. if it's bad, send to admin
	// if it's ok - log, if user - user

	//TODO: when handling errors, if error starts with ! send it to admin

	//TODO: error handler. all msgs go to errhandler, if it's
	// fatal(prefix !) send to admin, if it's for user send back to user
	// and log, if not bad > log

	//implement admin methods. so i could view full tables in some formatting
}
