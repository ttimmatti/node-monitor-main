package db

import (
	"context"
	"database/sql"
	"log"
	"net/url"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/ttimmatti/ironfish-node-tg/errror"
)

const USERS_T = "users"
const SERVERS_T = "servers"

var DB *sql.DB

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
		err = errror.FormatL(err)
		log.Fatalln("Could not open DB:", err)
	}
	//defer closeConn()

	if err := db.PingContext(context.Background()); err != nil {
		err = errror.FormatL(err)
		log.Fatalln("Could not ping DB:", err)
	}

	return db
}
