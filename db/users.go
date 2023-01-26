package db

import (
	"context"
	"fmt"
	"time"

	"github.com/ttimmatti/ironfish-node-tg/errror"
)

// 26jan. this file is ready

// ///////////////////////////////////////////////////////////
// METHODS for users
func UserExistIsBanned(chat_id int64) (bool, bool) {
	var username string
	var isBanned string

	id := fmt.Sprintf("%d", chat_id)

	row := DB.QueryRowContext(context.Background(),
		"select username,banned from users where chat_id=$1",
		id)

	if err := row.Scan(&username, &isBanned); err != nil {
		// user doesnt exist
		return false, false
	}

	if isBanned == "t" {
		return true, true
	}

	return true, false
}

func AddUser(chat_id int64, username, first_name string, date int64) error {
	id := fmt.Sprintf("%d", chat_id)
	time := time.Unix(date, 0)

	sqlresult, err := DB.ExecContext(context.Background(),
		"insert into users(chat_id,username,name,started) values($1,$2,$3,$4)",
		id, username, first_name, time)
	if err != nil {
		// the user is a duplicate(which can't ne as i always check if user IsUser)
		// or some fields are empty or corrupted
		return errror.WrapErrorF(err, errror.ErrorCodeFailure,
			"add_user_query_err_pk_or_corrupt_field (id,username,name):",
			id, username, first_name)
	}

	rows, _ := sqlresult.RowsAffected()
	if rows == 0 {
		return errror.NewErrorf(errror.ErrorCodeFailure,
			"add_user_rows_affected_0 (id,username,name):",
			id, username, first_name)
	}

	return nil
}

// METHODS for users
/////////////////////////////////////////////////////////////////
