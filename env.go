package main

import (
	"log"
	"os"
	"strconv"
)

// ////////////////////////////////////////////////////////////////////////
// ENVIRONMENT
func getDbEnv() (string, string, string, string) {
	port, user, pass, dbname := os.Getenv("PORT"),
		os.Getenv("USERNAME"),
		os.Getenv("PASS"),
		os.Getenv("DB_NAME")

	if port == "" || user == "" || pass == "" || dbname == "" {
		log.Fatalln("couldnt get one of the variables(4) for db from env",
			port, user, pass, dbname)
	}

	return port, user, pass, dbname
}

func getAdminIdEnv() int64 {
	adminId := os.Getenv("ADMIN_ID")
	id, err := strconv.ParseInt(adminId, 10, 64)
	if err != nil {
		log.Fatalln("admin_id error parsing to int")
	}

	return id
}
