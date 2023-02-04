package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ttimmatti/nodes-bot/ironfish/errror"
)

// 26jan. this page is set

func UpdateBlVServer(ip, v string, bl int64) error {
	block := fmt.Sprintf("%d", bl)

	sqlresult, err := DB.ExecContext(context.Background(),
		"update "+IRON_SERVERS_DB+" set block=$1,version=$2 where ip=$3",
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

func UpdateSyncUpdServer(ip string, synced, updated bool) error {
	tz, _ := time.LoadLocation("Europe/Moscow")
	tf := time.Now().In(tz).Format("Jan 2 15:04 MST")

	sqlresult, err := DB.ExecContext(context.Background(),
		"update "+IRON_SERVERS_DB+" set synced=$1,updated=$2,lastpong=$3 where ip=$4",
		synced, updated, tf, ip)
	if err != nil {
		// idk how it can happen
		return errror.WrapErrorF(err, errror.ErrorCodeFailure,
			"updateSU_server_exec_err (ip)", ip)
	}

	rows, err := sqlresult.RowsAffected()
	if err != nil {
		// IF THIS RETURNS ERROR CHECK CHANGEUSERSERVER ROWSAFFECTED TOO
		// AND EVERY OTHER ROWSAFFEVTED FUNC
		return errror.WrapErrorF(err, errror.ErrorCodeNotFound,
			"updateSU_server_rows_affected_error WHAT (ip)", ip)
	}

	if rows == 0 {
		return errror.NewErrorf(errror.ErrorCodeNotFound,
			"updateSU_server_rows_affected_0 (ip)", ip)
	}

	return nil
}

// id;;chat_id;;ip:;
func IronReadServers() (string, error) {
	rows, err := DB.QueryContext(context.Background(),
		"select id,chat_id,ip,block,version,synced,updated,lastpong from "+IRON_SERVERS_DB)
	if err != nil {
		return "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"admin_read_servers_query_err")
	}

	var row []string

	for i := 0; rows.Next(); i++ {
		var (
			id, chat_id, ip, block, version, synced, updated, lastPong string
		)

		_ = rows.Scan(&id, &chat_id, &ip, &block, &version, &synced, &updated, &lastPong)

		row = append(row, id+";;"+
			chat_id+";;"+
			ip+";;"+
			block+";;"+
			version+";;"+
			synced+";;"+
			updated+";;"+
			lastPong)
	}

	result := strings.Join(row, ":;")

	return result, nil
}
