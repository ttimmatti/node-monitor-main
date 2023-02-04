package db

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

func IronAddUserServer(chat_id int64, server_ip string) error {
	id := fmt.Sprintf("%d", chat_id)

	sqlresult, err := DB.ExecContext(context.Background(),
		"insert into "+IRON_SERVERS_DB+"(chat_id, ip) values($1,$2)",
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

// func ChangeUserServer(chat_id int64, server_ip1, server_ip2 string) error {
// 	id := fmt.Sprintf("%d", chat_id)

// 	sqlresult, err := DB.ExecContext(context.Background(),
// 		"update "+IRON_SERVERS_DB+" set ip=$1 where chat_id=$2 and ip=$3",
// 		server_ip2, id, server_ip1)
// 	if err != nil {
// 		// in case of duplicates or wrong id
// 		return errror.WrapErrorF(err,
// 			errror.ErrorCodeInvalidArgument,
// 			"change_user_server_duplicate_or_pk (id,ip):",
// 			id, server_ip1, server_ip2)
// 	}

// 	//if rows == 0 bad insert
// 	rowsAffected, _ := sqlresult.RowsAffected()
// 	if rowsAffected == 0 {
// 		//nothing changed
// 		return errror.NewErrorf(errror.ErrorCodeInvalidArgument,
// 			"change_user_server_rows_affected_0 (id,ip)",
// 			id, server_ip1, server_ip2)
// 	}

// 	return nil
// }

func IronGetUserServers(chat_id int64) (string, error) {
	id := fmt.Sprintf("%d", chat_id)

	rows, err := DB.QueryContext(context.Background(),
		"select ip,block,version,synced,updated,lastpong from "+IRON_SERVERS_DB+" where chat_id=$1",
		id)
	if err != nil {
		//he does not have any servers
		return "", errror.WrapErrorF(err, errror.ErrorCodeNotFound,
			"get_user_servers_query_error (id)", id)
	}
	defer rows.Close()

	var i int
	var result string
	for rows.Next() {
		i++
		counter := fmt.Sprintf("%d", i)
		var ip, block, version, synced, updated, lastPong, pong string
		err := rows.Scan(&ip, &block, &version, &synced, &updated, &lastPong)
		if err != nil {
			log.Print("GetUserServers: end of scan? SCAN err:", err)
		}

		ip = strings.Join(strings.Split(ip, "."), "\\.") + "\n"
		version = strings.Join(strings.Split(version, "."), "\\.")
		if lastPong != "" {
			pong = "; LastReponse: _" + lastPong + "_"
		}

		if block != "" {
			block = "_" + block + "_"
		}
		if version != "" {
			version = "_" + version + "_"
		}
		if synced != "" {
			synced = "_" + synced + "_"
		}
		if updated != "" {
			updated = "_" + updated + "_"
		}

		result += "\n' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' ' '\n" +
			counter + "\\. Server: " + ip + "Block: " + block +
			"; Version: " + version + "; Synced: " + synced +
			"; Updated: " + updated + pong
	}

	return result, nil
}

func IronDeleteUserServer(chat_id int64, server_ip string) error {
	id := fmt.Sprintf("%d", chat_id)

	sqlresult, err := DB.ExecContext(context.Background(),
		"delete from "+IRON_SERVERS_DB+" where chat_id=$1 and ip=$2",
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
