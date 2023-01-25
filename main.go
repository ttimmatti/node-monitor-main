package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	db "github.com/ttimmatti/ironfish-node-tg/db"
	msgs "github.com/ttimmatti/ironfish-node-tg/tg-msgs"
)

const TG_API = "https://api.telegram.org/bot5864005496:AAFYPu4VK53PD8rjmrMyFfIpnyaiCnQASeo"

const UPDATE_FREQ = 2 // every $int seconds

var LAST_MSG_INDEX int64

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln("Couldnt get Environment")
	}

	onExit()

	db.DB = db.SetConn(getDbEnv())
	msgs.ADMIN_ID = getAdminIdEnv()

	msgs.StartReceiving(TG_API, UPDATE_FREQ)

	// rules for error. if it's bad, send to admin
	// if it's ok - log, if user - user

	//TODO: when handling errors, if error starts with ! send it to admin

	//TODO: error handler. all msgs go to errhandler, if it's
	// fatal(prefix !) send to admin, if it's for user send back to user
	// and log, if not bad > log

	//implement admin methods. so i could view full tables in some formatting
}

func onExit() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		err := db.DB.Close()
		if err != nil {
			log.Println("couldnt close connection", err)
		}
		log.Println("connection was closed")
		os.Exit(1)
	}()
}
