package db

import (
	"context"
	"database/sql"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const USERS_DB = "users"
const SUI_SERVERS_DB = "sui_servers"

var DB *sql.DB

// 26jan. this file is set

func SetConn(port, user, pass, dbname string) *sql.DB {
	dsn := url.URL{
		Scheme: "postgres",
		Host:   "localhost:" + port,
		User:   url.UserPassword(user, pass),
		Path:   dbname,
	}

	q := dsn.Query()
	q.Add("sslmode", "disable")

	dsn.RawQuery = q.Encode()

	db, err := sql.Open("pgx", dsn.String())
	if err != nil {
		log.Fatalln("Could not open DB:", err)
	}
	//defer closeConn()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatalln("Could not ping DB:", err)
	}

	return db
}

func OnExit() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		err := DB.Close()
		if err != nil {
			log.Println("couldnt close connection", err)
		}
		log.Println("connection was closed")
		os.Exit(1)
	}()
}
