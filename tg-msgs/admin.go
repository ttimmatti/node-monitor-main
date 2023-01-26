package msgs

import (
	"fmt"
	"strings"

	"github.com/ttimmatti/ironfish-node-tg/db"
)

const ADMIN_COMMANDS = `/allusers
/allservers
/banid
/unbanid
/admincmd`

func adminCmd(cmd string, msg Msg) error {
	switch cmd {
	case "/allusers":
		if err := handleAllUsers(); err != nil {
			return fmt.Errorf("%w", err)
		}
	case "/allservers":
		if err := handleAllServers(); err != nil {
			return fmt.Errorf("%w", err)
		}
	case "/banid":
		if err := handleBanId(msg); err != nil {

		}
	case "/unbanid":
		if err := handleUnbanId(msg); err != nil {

		}
	case "/admincmd":
		if err := handleAdminCmd(); err != nil {

		}
	}

	return nil
}

func handleAdminCmd() error {
	msg := &SendMsg{
		Chat_id:    ADMIN_ID,
		Text:       ADMIN_COMMANDS,
		Parse_mode: "",
	}

	if err := sendMsg(msg); err != nil {
		return fmt.Errorf("handleAllUsers: %w", err)
	}

	return nil
}

func handleUnbanId(msg Msg) error {
	panic("unimplemented")
}

func handleBanId(msg Msg) error {
	panic("unimplemented")
}

func handleAllServers() error {
	result, err := db.ReadUsers()
	if err != nil {
		return fmt.Errorf("admin_handle-all-users: %w", err)
	}

	response := wrapDbResp(result)

	msg := &SendMsg{
		Chat_id:    ADMIN_ID,
		Text:       response,
		Parse_mode: "",
	}

	if err := sendMsg(msg); err != nil {
		return fmt.Errorf("handleAllServers: %w", err)
	}

	return nil
}

func handleAllUsers() error {
	result, err := db.ReadUsers()
	if err != nil {
		return fmt.Errorf("admin_handle-all-users: %w", err)
	}

	response := wrapDbResp(result)

	msg := &SendMsg{
		Chat_id:    ADMIN_ID,
		Text:       response,
		Parse_mode: "",
	}

	if err := sendMsg(msg); err != nil {
		return fmt.Errorf("handleAllUsers: %w", err)
	}

	return nil
}

func wrapDbResp(input string) string {
	var response string
	usersUF := strings.Split(input, ":;")

	response = strings.Join(usersUF, "\n")

	return response
}
