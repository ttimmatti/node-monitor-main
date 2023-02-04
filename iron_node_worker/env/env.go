package env

import (
	"log"
	"os"
	"strconv"
)

// ////////////////////////////////////////////////////////////////////////
// ENVIRONMENT
func GetDbEnv() (string, string, string, string) {
	port, user, pass, dbname := os.Getenv("PORT"),
		os.Getenv("USERNAME"),
		os.Getenv("PASS"),
		os.Getenv("DB_NAME")

	if port == "" || user == "" || pass == "" || dbname == "" {
		log.Fatalln("env: couldnt get one of the variables(4) for db from env",
			port, user, pass, dbname)
	}

	return port, user, pass, dbname
}

func GetAdminIdEnv() int64 {
	adminId := os.Getenv("ADMIN_ID")
	id, err := strconv.ParseInt(adminId, 10, 64)
	if err != nil {
		log.Fatalln("env: admin_id error parsing to int")
	}

	return id
}

func GetTGApiEnv() string {
	tgApi := os.Getenv("TG_API")
	if len(tgApi) < 2 {
		log.Fatalln("env: tg_api empty")
	}
	return tgApi
}

func GetGitTokenEnv() string {
	token := os.Getenv("GITHUB_TOKEN")
	if len(token) < 2 {
		log.Fatalln("env: tg_api empty")
	}
	return token
}
