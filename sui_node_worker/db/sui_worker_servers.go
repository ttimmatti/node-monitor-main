package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ttimmatti/nodes-bot/sui/errror"
)

// 26jan. this page is set

func UpdateTxVServer(ip, v string, bl int64) error {
	block := fmt.Sprintf("%d", bl)

	sqlresult, err := DB.ExecContext(context.Background(),
		"update "+SUI_SERVERS_DB+" set tx_id=$1,version=$2 where ip=$3",
		block, v, ip)
	if err != nil {
		// idk how it can happen
		return errror.WrapErrorF(err, errror.ErrorCodeFailure,
			"updateTxV_server_exec_err (ip)", ip)
	}

	rows, err := sqlresult.RowsAffected()
	if err != nil {
		// IF THIS RETURNS ERROR CHECK CHANGEUSERSERVER ROWSAFFECTED TOO
		// AND EVERY OTHER ROWSAFFEVTED FUNC
		return errror.WrapErrorF(err, errror.ErrorCodeNotFound,
			"updateTxV_server_rows_affected_error (ip)", ip)
	}

	if rows == 0 {
		return errror.NewErrorf(errror.ErrorCodeNotFound,
			"updateTxV_server_rows_affected_0 (ip)", ip)
	}

	return nil
}

func UpdateSyncUpdServer(ip, status string, synced, updated bool) error {
	t := time.Now().Unix()

	sqlresult, err := DB.ExecContext(context.Background(),
		"update "+SUI_SERVERS_DB+" set synced=$1,updated=$2,status=$3,lastpong=$4 where ip=$5",
		synced, updated, status, fmt.Sprintf("%d", t), ip)
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
func SuiReadServers() (string, error) {
	rows, err := DB.QueryContext(context.Background(),
		"select id,chat_id,ip,tx_id,version,synced,updated,status,lastpong from "+SUI_SERVERS_DB)
	if err != nil {
		return "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"admin_read_servers_query_err")
	}

	var row []string

	for i := 0; rows.Next(); i++ {
		var (
			id, chat_id, ip, tx_id, version, synced, updated, status, lastPong string
		)

		_ = rows.Scan(&id, &chat_id, &ip, &tx_id, &version, &synced, &updated, status, &lastPong)

		row = append(row, id+";;"+
			chat_id+";;"+
			ip+";;"+
			tx_id+";;"+
			version+";;"+
			synced+";;"+
			updated+";;"+
			lastPong+";;"+
			status)
	}

	result := strings.Join(row, ":;")

	return result, nil
}
