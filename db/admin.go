package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/ttimmatti/ironfish-node-tg/errror"
)

// for admin only

// returns string serial;;username;;name;;started
//
// joined with "&("
func ReadUser(chat_id int64) (string, error) {
	id := fmt.Sprintf("%d", chat_id)

	row := DB.QueryRowContext(context.Background(),
		"select * from users where chat_id=$1",
		id)

	var (
		serial, username, name, started string
	)

	err := row.Scan(&serial, &id, &username, &name, &started)
	if err != nil {
		return "", errror.WrapErrorF(err,
			errror.ErrorCodeNotFound,
			"admin_read_user_scan_err (id)", id)
	}

	result := serial + ";;" +
		id + ";;" +
		username + ";;" +
		name + ";;" +
		started

	return result, nil
}

// returns string username;;...;;started:;username
func ReadUsers() (string, error) {
	rows, err := DB.QueryContext(context.Background(),
		"select id,chat_id,username,name,started from users")
	if err != nil {
		return "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"admin_read_users_query_err")
	}

	var row []string

	for i := 0; rows.Next(); i++ {
		var (
			serial, id, username, name, started string
		)

		_ = rows.Scan(&serial, &id, &username, &name, &started)

		row = append(row, serial+";;"+
			id+";;"+
			username+";;"+
			name+";;"+
			started)
	}

	result := strings.Join(row, ":;")

	return result, nil
}

func ReadServers() (string, error) {
	rows, err := DB.QueryContext(context.Background(),
		"select id,chat_id,ip from servers")
	if err != nil {
		return "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"admin_read_servers_query_err")
	}

	var row []string

	for i := 0; rows.Next(); i++ {
		var (
			id, chat_id, ip string
		)

		_ = rows.Scan(&id, &chat_id, &ip)

		row = append(row, id+";;"+
			chat_id+";;"+
			ip)
	}

	result := strings.Join(row, ":;")

	return result, nil
}

//implement get, delete users for admin
