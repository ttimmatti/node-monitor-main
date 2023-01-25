package db

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ttimmatti/ironfish-node-tg/errror"
)

// 25jan. this page is error ready

func AddUserServer(chat_id int64, server_ip string) error {
	id := fmt.Sprintf("%d", chat_id)

	sqlresult, err := DB.ExecContext(context.Background(),
		"insert into servers(chat_id, ip) values($1,$2)",
		id, server_ip)
	if err != nil {
		return errror.FormatU(err)
	}

	//if rows == 0 bad insert
	rowsAffected, _ := sqlresult.RowsAffected()
	log.Println("rows affected:", rowsAffected)
	if rowsAffected == 0 {
		//nothing changed
		return errror.FormatsU("No changes were made.")
	}

	return nil
}

func ChangeUserServer(chat_id int64, server_ip1, server_ip2 string) error {
	id := fmt.Sprintf("%d", chat_id)

	sqlresult, err := DB.ExecContext(context.Background(),
		"update servers set ip=$1 where chat_id=$2 and ip=$3",
		server_ip2, id, server_ip1)
	if err != nil {
		return errror.FormatU(err)
	}

	//if rows == 0 bad insert
	rowsAffected, _ := sqlresult.RowsAffected()
	log.Println("rows affected:", rowsAffected)
	if rowsAffected == 0 {
		//nothing changed
		return errror.FormatsU("No changes were made.")
	}

	return nil
}

func GetServers(chat_id int64) (string, error) {
	id := fmt.Sprintf("%d", chat_id)

	rows, err := DB.QueryContext(context.Background(),
		"select ip from servers where chat_id=$1",
		id)
	if err != nil {
		log.Println("ERROR QUERY:", err)
		return "", err
	}

	log.Println("ROWS:", rows)

	var result string
	for rows.Next() {
		var ip string
		err := rows.Scan(&ip)
		if err != nil {
			log.Println("SCAN err:", err)
		}

		ip = "\n" + strings.Join(strings.Split(ip, "."), "\\.")
		result += ip
	}

	return result, nil
}

//when deleting dont forget to check that the user is the owner of the row
