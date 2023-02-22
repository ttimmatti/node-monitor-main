package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/ttimmatti/nodes-bot/ironfish/errror"
)

const IRON_MBS_DB = "iron_mbs"

//id serial,
//chat_id varchar not null references users(chat_id),
//ip text primary key,
//mint text,
//burn text,
//send text, t_mbs text, msb_done boolean

func UpdateMBSTime(ip, t_mbs int64) error {
	t := fmt.Sprintf("%d", t_mbs)

	sqlresult, err := DB.ExecContext(context.Background(),
		"update "+IRON_MBS_DB+" set t_mbs=$1 where ip=$2",
		t, ip)
	if err != nil {
		// idk how it can happen
		return errror.WrapErrorF(err, errror.ErrorCodeFailure,
			"UpdateMBSTime_exec_err (ip)", ip)
	}

	rows, err := sqlresult.RowsAffected()
	if err != nil {
		// IF THIS RETURNS ERROR CHECK CHANGEUSERSERVER ROWSAFFECTED TOO
		// AND EVERY OTHER ROWSAFFEVTED FUNC
		return errror.WrapErrorF(err, errror.ErrorCodeNotFound,
			"UpdateMBSTime_rows_affected_error (ip)", ip)
	}

	if rows == 0 {
		return errror.NewErrorf(errror.ErrorCodeNotFound,
			"UpdateMBSTime_rows_affected_0 (ip)", ip)
	}

	return nil
}

func UpdateMBSstatus(ip string, mint, burn, send, string, t_mbs int64, mbs_done bool) error {
	t := fmt.Sprintf("%d", t_mbs)

	sqlresult, err := DB.ExecContext(context.Background(),
		"update "+IRON_MBS_DB+" set mint=$1,burn=$2,send=$3,t_mbs=$4,mbs_done=$5 where ip=$6",
		mint, burn, send, t, mbs_done, ip)
	if err != nil {
		// idk how it can happen
		return errror.WrapErrorF(err, errror.ErrorCodeFailure,
			"UpdateMBSstatus_exec_err (ip)", ip)
	}

	rows, err := sqlresult.RowsAffected()
	if err != nil {
		// IF THIS RETURNS ERROR CHECK CHANGEUSERSERVER ROWSAFFECTED TOO
		// AND EVERY OTHER ROWSAFFEVTED FUNC
		return errror.WrapErrorF(err, errror.ErrorCodeNotFound,
			"UpdateMBSstatus_rows_affected_error WHAT (ip)", ip)
	}

	if rows == 0 {
		return errror.NewErrorf(errror.ErrorCodeNotFound,
			"UpdateMBSstatus_rows_affected_0 (ip)", ip)
	}

	return nil
}

// id;;chat_id;;ip:;
func IronMBSReadServers() (string, error) {
	rows, err := DB.QueryContext(context.Background(),
		"select id,chat_id,ip,mint,burn,send,t_mbs,mbs_done from "+IRON_MBS_DB)
	if err != nil {
		return "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"IronMBSReadServers_query_err")
	}

	var row []string

	for i := 0; rows.Next(); i++ {
		var (
			id, chat_id, ip, mint, burn, send string
			t_mbs                             int64
			mbs_done                          bool
		)

		_ = rows.Scan(&id, &chat_id, &ip, &mint, &burn, &send, &t_mbs, &mbs_done)

		row = append(row, id+";;"+
			chat_id+";;"+
			ip+";;"+
			mint+";;"+
			burn+";;"+
			send+";;"+
			fmt.Sprintf("%d", t_mbs)+";;"+
			fmt.Sprintf("%v", mbs_done))
	}

	result := strings.Join(row, ":;")

	return result, nil
}

func IronMBSAddServer(chat_id, ip, mint, burn, send string, t_mbs int64, mbs_done bool) error {
	sqlresult, err := DB.ExecContext(context.Background(),
		"insert into "+IRON_MBS_DB+"(chat_id,ip,mint,burn,send,t_mbs,mbs_done) values($1,$2,$3,$4,$5,$6,$7)",
	)
	if err != nil {
		// idk how it can happen
		return errror.WrapErrorF(err, errror.ErrorCodeFailure,
			"IronMBSAddServer_exec_err (ip)", ip)
	}

	rows, err := sqlresult.RowsAffected()
	if err != nil {
		// IF THIS RETURNS ERROR CHECK CHANGEUSERSERVER ROWSAFFECTED TOO
		// AND EVERY OTHER ROWSAFFEVTED FUNC
		return errror.WrapErrorF(err, errror.ErrorCodeNotFound,
			"IronMBSAddServer_rows_affected_error WHAT (ip)", ip)
	}

	if rows == 0 {
		return errror.NewErrorf(errror.ErrorCodeNotFound,
			"IronMBSAddServer_rows_affected_0 (ip)", ip)
	}

	return nil
}
