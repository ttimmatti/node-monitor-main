package db

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

const SUI_SERVERS_DB = "sui_servers"

func SuiAddUserServer(chat_id int64, server_ip string) error {
	id := fmt.Sprintf("%d", chat_id)

	sqlresult, err := DB.ExecContext(context.Background(),
		"insert into "+SUI_SERVERS_DB+"(chat_id, ip) values($1,$2)",
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
// 		"update "+SUI_SERVERS_DB+" set ip=$1 where chat_id=$2 and ip=$3",
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

func SuiGetUserServers(chat_id int64) (string, error) {
	c := make(chan struct {
		Tx  int64
		Err error
	})
	go SuiWorkerLastTx(c)

	id := fmt.Sprintf("%d", chat_id)

	rows, err := DB.QueryContext(context.Background(),
		"select ip,tx_id,version,synced,updated,status,lastpong from "+SUI_SERVERS_DB+" where chat_id=$1",
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
		var ip, tx_id, version, synced, updated, status, lastPong, pong string
		err := rows.Scan(&ip, &tx_id, &version, &synced, &updated, &status, &lastPong)
		if err != nil {
			log.Print("GetUserServers: end of scan? SCAN err:", err)
		}

		ip = strings.Join(strings.Split(ip, "."), "\\.")
		version = strings.Join(strings.Split(version, "."), "\\.")
		if lastPong != "" {
			t, _ := strconv.ParseInt(lastPong, 10, 64)
			tz, _ := time.LoadLocation("Europe/Moscow")
			tf := time.Unix(t, 0).In(tz).Format("Jan 2 15:04 MST")
			pong = "; LastReponse: _" + tf + "_"
		}

		if tx_id != "" {
			tx_id = "_" + tx_id + "_"
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
			counter + "\\. Server: " + ip + " \\(" + status + "\\)\n" +
			"Tx: " + tx_id +
			"; Version: " + version + "; Synced: " + synced +
			"; Updated: " + updated + pong
	}

	var result1 string

	tx_count := <-c
	if tx_count.Err != nil {
		log.Printf("SuiGetUserServers: %s", tx_count.Err)
		result1 = result
	} else {
		result1 = " \\(Last Network Tx: " + fmt.Sprintf("%d", tx_count.Tx) + "\\)" + result
	}

	return result1, nil
}

func SuiDeleteUserServer(chat_id int64, server_ip string) error {
	id := fmt.Sprintf("%d", chat_id)

	sqlresult, err := DB.ExecContext(context.Background(),
		"delete from "+SUI_SERVERS_DB+" where chat_id=$1 and ip=$2",
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

func SuiWorkerLastTx(c chan struct {
	Tx  int64
	Err error
}) {
	resp, err := http.Get("http://localhost:6591/tx_id")
	if err != nil {
		c <- struct {
			Tx  int64
			Err error
		}{
			0, fmt.Errorf("SuiWorkerLastTx: %w", err),
		}
		return
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	last_tx, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		c <- struct {
			Tx  int64
			Err error
		}{
			0, fmt.Errorf("SuiWorkerLastTx: %w", err),
		}
		return
	}

	c <- struct {
		Tx  int64
		Err error
	}{
		last_tx, nil,
	}
}
