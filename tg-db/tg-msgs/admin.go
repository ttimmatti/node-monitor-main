package msgs

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/ttimmatti/nodes-bot/tg-db/db"
	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

var LAST_INVITE_CODE int64

const ADMIN_COMMANDS = `/allusers
/allservers
/banuser
/unbanuser
/getcode
/deleteuser
/admincmd`

func adminCmd(cmd string, msg Msg) []error {
	switch cmd {
	case "/allusers":
		if err := handleAllUsers(); err != nil {
			return []error{fmt.Errorf("%w", err)}
		}
	case "/allservers":
		if err := handleAllServers(); err != nil {
			return []error{fmt.Errorf("%w", err)}
		}
	case "/banuser":
		if err := handleBanUsername(msg); err != nil {
			return []error{fmt.Errorf("%w", err)}
		}
	case "/unbanuser":
		if err := handleUnbanUsername(msg); err != nil {
			return []error{fmt.Errorf("%w", err)}
		}
	case "/getcode":
		if err := handleGetCode(msg); err != nil {
			return []error{fmt.Errorf("%w", err)}
		}
	case "/deleteuser":
		if errs := handleDeleteUser(msg); errs != nil {
			return errs
		}
	case "/admincmd":
		if err := handleAdminCmd(); err != nil {
			return []error{fmt.Errorf("%w", err)}
		}

	}

	return nil
}

func handleDeleteUser(msg Msg) []error {
	errs := []error{}
	_, username, _, err := parseCmd(msg.Text)
	if err != nil {
		return []error{fmt.Errorf("handle_ban_username: %w", err)}
	}

	if strings.Contains(username, "@") {
		username = username[1:]
	}

	user, err := db.ReadUser(username)
	if err != nil {
		return []error{fmt.Errorf("admin_handleDeleteUser: %w", err)}
	}
	userFields := strings.Split(user, ";;")
	if len(userFields) < 2 {
		return []error{errror.NewErrorf(errror.ErrorCodeInvalidArgument, "admin_handleDeleteUser: no userFields", user)}
	}
	id := userFields[1]
	if len(id) < 2 {
		return []error{errror.NewErrorf(errror.ErrorCodeInvalidArgument, "admin_handleDeleteUser: id empty", id, user)}
	}
	chat_id, _ := strconv.ParseInt(id, 10, 64)

	err = db.DeleteAllIronServers(chat_id)
	if err != nil {
		errs = append(errs, fmt.Errorf("admin_handleDeleteUser: %w", err))
	}
	err = db.DeleteAllIronServers(chat_id)
	if err != nil {
		errs = append(errs, fmt.Errorf("admin_handleDeleteUser: %w", err))
	}
	err = db.DeleteUser(chat_id)
	if err != nil {
		errs = append(errs, fmt.Errorf("admin_handleDeleteUser: %w", err))
	}

	msgResp := &SendMsg{
		Chat_id:    ADMIN_ID,
		Text:       fmt.Sprintf("deleted user: %s, errs %s", username, errs),
		Parse_mode: "",
	}

	if err := sendMsg(msgResp); err != nil {
		errs = append(errs, fmt.Errorf("admin_handleDeleteUser: %w", err))
		return errs
	}

	return nil
}

func handleGetCode(msg Msg) error {
	code := rand.Int63n(10000)*rand.Int63n(10000) + 100000

	LAST_INVITE_CODE = code

	msgResp := &SendMsg{
		Chat_id:    ADMIN_ID,
		Text:       fmt.Sprintf("/code %d", LAST_INVITE_CODE),
		Parse_mode: "",
	}

	if err := sendMsg(msgResp); err != nil {
		return fmt.Errorf("handleGetCode: %w", err)
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

	if strings.Contains(username, "@") {
		username = username[1:]
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

	if strings.Contains(username, "@") {
		username = username[1:]
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
	result1, err := db.ReadIronServers()
	if err != nil {
		return fmt.Errorf("admin_handle-all-servers: %w", err)
	}
	result2, err := db.SuiReadServers()
	if err != nil {
		return fmt.Errorf("admin_handle-all-servers: %w", err)
	}

	response := "Ironfish:\n" + wrapDbResp(result1) + "\n\nSui:" + wrapDbResp(result2)

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
