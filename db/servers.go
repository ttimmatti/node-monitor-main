package db

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ttimmatti/ironfish-node-tg/errror"
)

// 26jan. this page is set

func AddUserServer(chat_id int64, server_ip string) error {
	id := fmt.Sprintf("%d", chat_id)

	sqlresult, err := DB.ExecContext(context.Background(),
		"insert into servers(chat_id, ip) values($1,$2)",
		id, server_ip)
	if err != nil {
		//in case of duplicated or wrong user
		return errror.WrapErrorF(err,
			errror.ErrorCodeInvalidArgument,
			"add_user_server_duplicate_or_pk (id,ip):", chat_id, server_ip)
	}

	//if rows == 0 bad insert
	rowsAffected, _ := sqlresult.RowsAffected()
	if rowsAffected == 0 {
		//nothing changed
		return errror.NewErrorf(errror.ErrorCodeInvalidArgument,
			"add_user_server_rows_affected_0 (id,ip)", chat_id, server_ip)
	}

	return nil
}

func ChangeUserServer(chat_id int64, server_ip1, server_ip2 string) error {
	id := fmt.Sprintf("%d", chat_id)

	sqlresult, err := DB.ExecContext(context.Background(),
		"update servers set ip=$1 where chat_id=$2 and ip=$3",
		server_ip2, id, server_ip1)
	if err != nil {
		// in case of duplicates or wrong id
		return errror.WrapErrorF(err,
			errror.ErrorCodeInvalidArgument,
			"change_user_server_duplicate_or_pk (id,ip):",
			id, server_ip1, server_ip2)
	}

	//if rows == 0 bad insert
	rowsAffected, _ := sqlresult.RowsAffected()
	if rowsAffected == 0 {
		//nothing changed
		return errror.NewErrorf(errror.ErrorCodeInvalidArgument,
			"change_user_server_rows_affected_0 (id,ip)",
			id, server_ip1, server_ip2)
	}

	return nil
}

func GetUserServers(chat_id int64) (string, error) {
	id := fmt.Sprintf("%d", chat_id)

	rows, err := DB.QueryContext(context.Background(),
		"select ip,block,version from servers where chat_id=$1",
		id)
	if err != nil {
		//he does not have any servers
		return "", errror.WrapErrorF(err, errror.ErrorCodeNotFound,
			"get_user_servers_query_error (id)", id)
	}
	defer rows.Close()

	//TODO: change this to easy to read output
	var result string
	for rows.Next() {
		var ip, block, version string
		err := rows.Scan(&ip, &block, &version)
		if err != nil {
			log.Println("GetUserServers: end of scan? SCAN err:", err)
		}

		ip = "\n" + strings.Join(strings.Split(ip, "."), "\\.")
		result += ip + "block:" + block + "version:" + version
	}

	return result, nil
}

func DeleteUserServer(chat_id int64, server_ip string) error {
	id := fmt.Sprintf("%d", chat_id)

	sqlresult, err := DB.ExecContext(context.Background(),
		"delete from servers where chat_id=$1 and ip=$2",
		id, server_ip)
	if err != nil {
		// idk how it can happen
		return errror.WrapErrorF(err, errror.ErrorCodeFailure,
			"delete_user_server_query_err (id,ip)", id, server_ip)
	}

	rows, err := sqlresult.RowsAffected()
	if err != nil {
		// IF THIS RETURNS ERROR CHECK CHANGEUSERSERVER ROWSAFFECTED TOO
		// AND EVERY OTHER ROWSAFFEVTED FUNC
		return errror.WrapErrorF(err, errror.ErrorCodeNotFound,
			"delete_user_server_rows_affected_error (id,ip)", chat_id, server_ip)
	}

	if rows == 0 {
		return errror.NewErrorf(errror.ErrorCodeNotFound,
			"delete_user_server_rows_affected_0 (id,ip)", chat_id, server_ip)
	}

	return nil
}

func UpdateBlVServer(ip, v string, bl int64) error {
	block := fmt.Sprintf("%d", bl)

	sqlresult, err := DB.ExecContext(context.Background(),
		"update servers set block=$1,version=$2 where ip=$3",
		block, v, ip)
	if err != nil {
		// idk how it can happen
		return errror.WrapErrorF(err, errror.ErrorCodeFailure,
			"updateBlV_server_exec_err (ip)", ip)
	}

	rows, err := sqlresult.RowsAffected()
	if err != nil {
		// IF THIS RETURNS ERROR CHECK CHANGEUSERSERVER ROWSAFFECTED TOO
		// AND EVERY OTHER ROWSAFFEVTED FUNC
		return errror.WrapErrorF(err, errror.ErrorCodeNotFound,
			"updateBlV_server_rows_affected_error (ip)", ip)
	}

	if rows == 0 {
		return errror.NewErrorf(errror.ErrorCodeNotFound,
			"updateBlV_server_rows_affected_0 (ip)", ip)
	}

	return nil
}
