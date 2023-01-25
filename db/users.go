package db

import (
	"context"
	"fmt"
	"time"

	"github.com/ttimmatti/ironfish-node-tg/errror"
)

// ///////////////////////////////////////////////////////////
// METHODS for users
func UserExist(chat_id int64) bool {
	var username string

	id := fmt.Sprintf("%d", chat_id)

	row := DB.QueryRowContext(context.Background(),
		"select username from users where chat_id=$1",
		id)

	if err := row.Scan(&username); err != nil {
		return false
	}

	return true
}

func AddUser(chat_idI int64, username, first_name string, date int64) error {
	chat_id := fmt.Sprintf("%d", chat_idI)
	time := time.Unix(date, 0)

	sqlresult, err := DB.ExecContext(context.Background(),
		"insert into users(chat_id,username,name,started) values($1,$2,$3,$4)",
		chat_id, username, first_name, time)
	if err != nil {
		return errror.FormatA(err)
	}

	rows, _ := sqlresult.RowsAffected()
	if rows == 0 {
		errror.FormatA(fmt.Errorf("Tried to insert user (Id: %d, Username: %s) but failed. --> Rows affected: 0",
			chat_idI, username))
	}

	return nil
}

// for admin only
func ReadUser(chat_id int64) (string, error) {
	id := fmt.Sprintf("%d", chat_id)

	row := DB.QueryRowContext(context.Background(),
		"select username from users where chat_id=$1",
		id)

	var username string
	err := row.Scan(&username)
	if err != nil {
		return "", fmt.Errorf("error in row scan")
	}

	return username, nil
}

// METHODS for users
/////////////////////////////////////////////////////////////////

//implement get, delete users for admin
