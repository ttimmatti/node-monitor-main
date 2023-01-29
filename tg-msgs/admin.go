package msgs

import (
	"fmt"
	"strings"

	"github.com/ttimmatti/ironfish-node-tg/db"
)

const ADMIN_COMMANDS = `/allusers
/allservers
/banuser
/unbanuser
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
	case "/banuser":
		if err := handleBanUsername(msg); err != nil {
			return fmt.Errorf("%w", err)
		}
	case "/unbanuser":
		if err := handleUnbanUsername(msg); err != nil {
			return fmt.Errorf("%w", err)
		}
	case "/admincmd":
		if err := handleAdminCmd(); err != nil {
			return fmt.Errorf("%w", err)
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

func handleBanUsername(msg Msg) error {
	_, username, _, err := parseCmd(msg.Text)
	if err != nil {
		return fmt.Errorf("handle_ban_username: %w", err)
	}

	err = db.BanUser(username)
	if err != nil {
		return fmt.Errorf("admin_handle-ban-username: %w", err)
	}

	msgResp := &SendMsg{
		Chat_id:    ADMIN_ID,
		Text:       fmt.Sprintf("banned: %s", username),
		Parse_mode: "",
	}

	if err := sendMsg(msgResp); err != nil {
		return fmt.Errorf("handle-ban-username: %w", err)
	}

	return nil
}

func handleUnbanUsername(msg Msg) error {
	_, username, _, err := parseCmd(msg.Text)
	if err != nil {
		return fmt.Errorf("handle_ban_username: %w", err)
	}

	err = db.UnbanUser(username)
	if err != nil {
		return fmt.Errorf("admin_handle-unban-username: %w", err)
	}

	msgResp := &SendMsg{
		Chat_id:    ADMIN_ID,
		Text:       fmt.Sprintf("unbanned: %s", username),
		Parse_mode: "",
	}

	if err := sendMsg(msgResp); err != nil {
		return fmt.Errorf("handle-unban-username: %w", err)
	}

	return nil
}

func handleAllServers() error {
	result, err := db.ReadServers()
	if err != nil {
		return fmt.Errorf("admin_handle-all-servers: %w", err)
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
