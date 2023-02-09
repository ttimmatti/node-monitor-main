package msgs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ttimmatti/nodes-bot/tg-db/db"
	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

const PORTfish = ":6596"

func handleIronfishMsg(isUser bool, msg Msg) error {
	_, cmd2, _, _ := parseCmd(msg.Text)

	switch cmd2 {
	case "":
		if err := ironfishHelp(isUser, msg); err != nil {
			return fmt.Errorf("ironfishMsg: %w", err)
		}
	case "add":
		if err := ironUserAddServer(isUser, msg); err != nil {
			return fmt.Errorf("ironfishMsg: %w", err)
		}
	case "servers":
		if err := ironUserServers(isUser, msg); err != nil {
			return fmt.Errorf("ironfishMsg: %w", err)
		}
	case "delete":
		if err := ironUserDeleteServer(isUser, msg); err != nil {
			return fmt.Errorf("ironfishMsg: %w", err)
		}
	}
	return nil
}

func ironfishHelp(isUser bool, msg Msg) error {
	ironfish_help := strings.Join(strings.Split(IRONFISH_HELP0, "."), "\\.")
	ironfish_help = strings.Join(strings.Split(ironfish_help, "-"), "\\-")
	ironfish_help = strings.Join(strings.Split(ironfish_help, "("), "\\(")
	ironfish_help = strings.Join(strings.Split(ironfish_help, ")"), "\\)")

	if isUser {
		msg2 := ReplyWithInlKey(msg.From.Id, ironfish_help, PARSE_MODE, IRONFFISH_HELP_INL_KEYBOARD)

		if err := sendMsg(msg2); err != nil {
			return fmt.Errorf("ironfishHelp: %w", err)
		}
	} else {
		msg2 := DefaultReply(msg.From.Id, "You're not registered, pls start with a /start cmd", "HTML")

		if err := sendMsg(msg2); err != nil {
			return fmt.Errorf("ironfishHelp: %w", err)
		}
	}

	return nil
}

func ironUserDeleteServer(isUser bool, msg Msg) error {
	if !isUser {
		//TODO: sendmessage "you're not a user, please type start to start..."
		// and guide WELCOME
		// and return
		msg2 := DefaultReply(msg.From.Id, "You're not a user", "HTML")

		err := sendMsg(msg2)
		if err != nil {
			return fmt.Errorf("ironUserDeleteServer: %w", err)
		}

		return errror.NewErrorf(errror.ErrorCodeNotFound,
			"ironUserDeleteServer: user not registered")
	}

	//parse cmd
	_, _, server_ip1, err := parseCmd(msg.Text)
	if err != nil {
		return fmt.Errorf("ironUserDeleteServer: %w", err)
	}

	//check that the cmd is valid
	isValid1 := validServerIp(server_ip1)
	if !isValid1 {
		return errror.NewErrorf(errror.ErrorCodeInvalidArgument,
			"ironUserDeleteServer: ip not valid")
	}

	//it can be that the server is already in the db
	// tell the user to check servers and verify that it's not there
	// or contact me
	if err := db.IronDeleteUserServer(msg.From.Id, server_ip1); err != nil {
		return fmt.Errorf("ironUserDeleteServer: %w", err)
	}

	// if rows affected = 1 --> success
	// if rows affected = 0 --> no changes

	msg2 := DefaultReply(msg.From.Id, "succesfully deleted server", PARSE_MODE)

	if err := sendMsg(msg2); err != nil {
		return fmt.Errorf("ironUserDeleteServer: %w", err)
	}

	return nil
}

func ironUserServers(isUser bool, msg Msg) error {
	if isUser {
		servers, err := db.IronGetUserServers(msg.From.Id)
		if err != nil {
			return fmt.Errorf("ironUserServers: %w", err)
		}

		msg2 := ReplyWithReplyKeyboard(msg.From.Id,
			"*YOUR SERVERS:*"+servers,
			PARSE_MODE,
			SERVERS_RPL_KEYBOARD)

		if err := sendMsg(msg2); err != nil {
			return fmt.Errorf("ironUserServers: %w", err)
		}
	} else {
		msg2 := DefaultReply(msg.From.Id, "You're not registered, pls start with a /start cmd", "HTML")

		if err := sendMsg(msg2); err != nil {
			return fmt.Errorf("ironUserServers: %w", err)
		}
	}

	return nil
}

// ready
func ironUserAddServer(isUser bool, msg Msg) error {
	if !isUser {
		msg2 := DefaultReply(msg.From.Id, "You're not registered, pls start with a /start cmd", "HTML")
		if err := sendMsg(msg2); err != nil {
			return fmt.Errorf("ironUserServers: %w", err)
		}
		return nil
	}

	//parse cmd
	_, _, server_ip, err := parseCmd(msg.Text)
	if err != nil {
		return fmt.Errorf("ironUserAddServer: %w", err)
	}

	//check that the cmd is valid
	isValid := validServerIp(server_ip)
	if !isValid {
		return errror.NewErrorf(errror.ErrorCodeInvalidArgument,
			"ironUserAddServer: ip not valid")
	}

	pong := pingIronServer(server_ip)
	if !pong {
		return errror.NewErrorf(errror.ErrorCodePongFalse,
			"server does not respond to ping")
	}

	//it can be that the server is already in the db
	// tell the user to check servers and verify that it's not there
	// or contact me
	if err := db.IronAddUserServer(msg.From.Id, server_ip); err != nil {
		return fmt.Errorf("ironUserAddServer: %w", err)
	}

	respMsg := DefaultReply(msg.From.Id, "succesfully added server", PARSE_MODE)

	if err := sendMsg(respMsg); err != nil {
		return fmt.Errorf("ironUserAddServer: %w", err)
	}

	return nil
}

func pingIronServer(server_ip string) bool {
	//TODO: ping server somehow!!!
	// server := node_worker.Server{
	// 	Owner_id: "",
	// 	Ip:       server_ip,
	// }
	// if err := server.Ping(); err != nil {
	// 	return errror.WrapErrorF(err,
	// 		errror.ErrorCodeNotFound,
	// 		"handle-add-user-server:")
	// }
	uri := "http://" + server_ip + PORTfish + "/ping"

	ctx, _ := context.WithTimeout(context.Background(),
		2000*time.Millisecond)
	r, _ := http.NewRequestWithContext(ctx,
		http.MethodGet,
		uri, nil)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := make(map[string]string)
	if err := json.Unmarshal(body, &result); err != nil {
		return false
	}

	if result["result"] != "ok" {
		return false
	}

	return true
}

const IRONFISH_HELP0 = `*IronFish Bot*
			
_The bot monitors Sync status and Version of your Ironfish node.
If Your node looses sync or version gets outdated the bot will Notify you Once._

` +
	"Before adding the server you need to install additional software on it. Use the below command on your server:\n" +
	"`. <(wget -qO- https://nodes.fackblock.com/api/iron_tg.sh)`\n\n" +
	"1. Add server: `/ironfish add 123.123.123.123`\n" +
	"2. Delete server: `/ironfish delete 123.123.123.123`\n" +
	"3. Check ironfish servers: `/ironfish servers`\n\n" +
	"----------------------------------------------------------------------\n\n" +
	"_Бот мониторит синк статус и версию вашей ноды Ironfish.\n" +
	"Если ваша нода потеряет синхронизацию или будет работать на устаревшей версии, бот пришлет увеломление._\n\n" +
	"Перед тем, как добавить сервер, необходимо установить дополнительное приложение на него. Для этого используйте команду (копировать и вставить одной командой):\n" +
	"`. <(wget -qO- https://nodes.fackblock.com/api/iron_tg.sh)`\n\n" +
	"1. Добавить сервер: `/ironfish add 123.123.123.123`\n" +
	"2. Удалить сервер: `/ironfish delete 123.123.123.123`\n" +
	"3. Проверить серверы: `/ironfish servers`"

var IRONFFISH_HELP_INL_KEYBOARD InlineKeyboardMarkup = InlineKeyboardMarkup{
	Inline_Keyboard: [][]InlineKeyboardButton{
		{InlineKeyboardButton{
			Text: "/ironfish servers", Callback_data: "/ironfish servers",
		}},
	},
}
