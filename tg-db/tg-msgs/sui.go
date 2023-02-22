package msgs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ttimmatti/nodes-bot/tg-db/db"
	"github.com/ttimmatti/nodes-bot/tg-db/errror"
	"github.com/ttimmatti/nodes-bot/tg-db/sui"
)

const PORTsui = ":9000"

func handleSuiMsg(isUser bool, msg Msg) error {
	_, cmd2, _, _ := parseCmd(msg.Text)

	switch cmd2 {
	case "":
		if err := suiHelp(isUser, msg); err != nil {
			return fmt.Errorf("suiMsg: %w", err)
		}
	case "add":
		if err := suiUserAddServer(isUser, msg); err != nil {
			return fmt.Errorf("suiMsg: %w", err)
		}
	case "servers":
		if err := suiUserServers(isUser, msg); err != nil {
			return fmt.Errorf("suiMsg: %w", err)
		}
	case "delete":
		if err := suiUserDeleteServer(isUser, msg); err != nil {
			return fmt.Errorf("suiMsg: %w", err)
		}
	case "checker":
		if err := suiServerChecker(msg); err != nil {
			return fmt.Errorf("suiMsg: %w", err)
		}
	case "check":
		if err := suiCheckTxServer(msg); err != nil {
			return fmt.Errorf("suiMsg: %w", err)
		}
	}
	return nil
}

func suiServerChecker(msg Msg) error {
	serversS, err := db.SuiGetUserIps(msg.From.Id)
	if err != nil {
		return fmt.Errorf("suiServerChecker: %w", err)
	}

	sui_checker_inl_keyb := suiCheckerInlKeyb(serversS)

	msg2 := ReplyWithInlKey(msg.From.Id, "Choose server id / Выберите ID сервера\n\n"+
		"Server ids are the same as in `/sui servers` output\n"+
		strings.Join(strings.Split("---------------------------------------------------------\n", "-"), "\\-")+
		"ID серверов те же, что и в выводе команды `/sui servers`", PARSE_MODE,
		sui_checker_inl_keyb)

	if err := sendMsg(msg2); err != nil {
		return fmt.Errorf("suiHelp: %w", err)
	}

	return nil
}

func diskSpaceInlKeyboard(input_ips string) InlineKeyboardMarkup {
	keyboard := [][]InlineKeyboardButton{}

	s_ips := strings.Join(strings.Split(input_ips, ";;"), ";")

	keyboard = append(keyboard, []InlineKeyboardButton{{
		Text: "Check Disk Space", Callback_data: "/servers_disk_" + s_ips,
	}})

	return InlineKeyboardMarkup{
		keyboard,
	}
}

func suiCheckerInlKeyb(input_ips string) InlineKeyboardMarkup {
	keyboard := [][]InlineKeyboardButton{}

	ips := strings.Split(input_ips, ";;")

	// for i := 0; i < int(math.Floor(float64(len(ips))/float64(5))); i = i + 5 {
	// 	keyboard = append(keyboard, []InlineKeyboardButton{})
	// 	for i2 := i; i2 < len(ips)-i; i2++ {
	// 		ip := ips[i2]
	// 		keyboard[i] = append(keyboard[i], InlineKeyboardButton{
	// 			Text:          fmt.Sprintf("%d", i),
	// 			Callback_data: "/sui_check_" + ip,
	// 		})
	// 	}
	// }

	i2 := -1
	for i, ip := range ips {
		if i%5 == 0 {
			i2++
			keyboard = append(keyboard, []InlineKeyboardButton{})
		}
		keyboard[i2] = append(keyboard[i2],
			InlineKeyboardButton{
				Text:          fmt.Sprintf("%d", i+1),
				Callback_data: "/sui_check_" + ip,
			})
	}

	return InlineKeyboardMarkup{
		keyboard,
	}
}

func suiCheckTxServer(msg Msg) error {
	_, _, server_ip, err := parseCmd(msg.Text)
	if err != nil {
		return fmt.Errorf("suiCheckTxServer: %w", err)
	}

	//check that the cmd is valid
	isValid1 := validServerIp(server_ip)
	if !isValid1 {
		return errror.NewErrorf(errror.ErrorCodeInvalidArgument,
			"suiCheckTxServer: ip not valid")
	}

	c := make(chan struct {
		Tx_id int64
		Err   error
	})
	go sui.GetTxId(server_ip, c)

	resp := <-c
	if resp.Err != nil {
		return fmt.Errorf("suiCheckTxServer: %w", resp.Err)
	}

	tx := resp.Tx_id

	// if rows affected = 1 --> success
	// if rows affected = 0 --> no changes

	server_ip = strings.Join(strings.Split(server_ip, "."), "\\.")
	msg2 := DefaultReply(msg.From.Id, "*"+server_ip+"*\nLast tx on server: "+fmt.Sprintf("%d", tx), PARSE_MODE)

	if err := sendMsg(msg2); err != nil {
		return fmt.Errorf("suiCheckTxServer: %w", err)
	}

	return nil
}

func suiHelp(isUser bool, msg Msg) error {
	sui_help := strings.Join(strings.Split(SUI_HELP, "."), "\\.")
	sui_help = strings.Join(strings.Split(sui_help, "-"), "\\-")
	sui_help = strings.Join(strings.Split(sui_help, "("), "\\(")
	sui_help = strings.Join(strings.Split(sui_help, ")"), "\\)")

	if isUser {
		msg2 := ReplyWithInlKey(msg.From.Id, sui_help, PARSE_MODE,
			SUI_HELP_INL_KEYBOARD)

		if err := sendMsg(msg2); err != nil {
			return fmt.Errorf("suiHelp: %w", err)
		}
	} else {
		msg2 := DefaultReply(msg.From.Id, "You're not registered, pls start with a /start cmd", "HTML")

		if err := sendMsg(msg2); err != nil {
			return fmt.Errorf("suiHelp: %w", err)
		}
	}

	return nil
}

func suiUserDeleteServer(isUser bool, msg Msg) error {
	if !isUser {
		//TODO: sendmessage "you're not a user, please type start to start..."
		// and guide WELCOME
		// and return
		msg2 := DefaultReply(msg.From.Id, "You're not a user", "HTML")

		err := sendMsg(msg2)
		if err != nil {
			return fmt.Errorf("suiUserDeleteServer: %w", err)
		}

		return errror.NewErrorf(errror.ErrorCodeNotFound,
			"suiUserDeleteServer: user not registered")
	}

	//parse cmd
	_, _, server_ip1, err := parseCmd(msg.Text)
	if err != nil {
		return fmt.Errorf("suiUserDeleteServer: %w", err)
	}

	//check that the cmd is valid
	isValid1 := validServerIp(server_ip1)
	if !isValid1 {
		return errror.NewErrorf(errror.ErrorCodeInvalidArgument,
			"suiUserDeleteServer: ip not valid")
	}

	if err := db.SuiDeleteUserServer(msg.From.Id, server_ip1); err != nil {
		return fmt.Errorf("suiUserDeleteServer: %w", err)
	}

	// if rows affected = 1 --> success
	// if rows affected = 0 --> no changes

	msg2 := DefaultReply(msg.From.Id, "succesfully deleted server", PARSE_MODE)

	if err := sendMsg(msg2); err != nil {
		return fmt.Errorf("suiUserDeleteServer: %w", err)
	}

	return nil
}

func suiUserServers(isUser bool, msg Msg) error {
	if isUser {
		servers, err := db.SuiGetUserServers(msg.From.Id)
		if err != nil {
			return fmt.Errorf("suiUserServers: %w", err)
		}

		msg2 := ReplyWithReplyKeyboard(msg.From.Id,
			"*YOUR SERVERS:*"+servers,
			PARSE_MODE,
			SERVERS_RPL_KEYBOARD)

		if err := sendMsg(msg2); err != nil {
			return fmt.Errorf("suiUserServers: %w", err)
		}
	} else {
		msg2 := DefaultReply(msg.From.Id, "You're not registered, pls start with a /start cmd", "HTML")

		if err := sendMsg(msg2); err != nil {
			return fmt.Errorf("suiUserServers: %w", err)
		}
	}

	return nil
}

// ready
func suiUserAddServer(isUser bool, msg Msg) error {
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
		return fmt.Errorf("suiUserAddServer: %w", err)
	}

	//check that the cmd is valid
	isValid := validServerIp(server_ip)
	if !isValid {
		return errror.NewErrorf(errror.ErrorCodeInvalidArgument,
			"suiUserAddServer: ip not valid")
	}

	pong := pingSuiServer(server_ip)
	if !pong {
		return errror.NewErrorf(errror.ErrorCodePongFalse,
			"server does not respond to ping")
	}

	//it can be that the server is already in the db
	// tell the user to check servers and verify that it's not there
	// or contact me
	if err := db.SuiAddUserServer(msg.From.Id, server_ip); err != nil {
		return fmt.Errorf("suiUserAddServer: %w", err)
	}

	respMsg := DefaultReply(msg.From.Id, "succesfully added server", PARSE_MODE)

	if err := sendMsg(respMsg); err != nil {
		return fmt.Errorf("suiUserAddServer: %w", err)
	}

	return nil
}

func pingSuiServer(server_ip string) bool {
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
	uri := "http://" + server_ip + PORTsui

	ctx, _ := context.WithTimeout(context.Background(),
		3000*time.Millisecond)
	r, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		uri, nil)
	r.Header["Content-Type"] = []string{"application/json"}
	if err != nil {
		return false
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := string(body)

	return strings.Contains(result, "HTTP")
}

const SUI_HELP = `*Sui Bot*
			
_The bot monitors Sync status and Version of your Sui node.
If Your node looses sync or version gets outdated the bot will Notify you Once._

` +
	"1. Add server: `/sui add 123.123.123.123`\n" +
	"2. Delete server: `/sui delete 123.123.123.123`\n" +
	"3. Check Sui servers: `/sui servers`\n\n" +
	"----------------------------------------------------------------------\n\n" +
	`_Бот мониторит синк статус и версию вашей ноды Sui.
Если ваша нода потеряет синхронизацию или будет работать на устаревшей версии, бот пришлет увеломление._

` +
	"1. Добавить сервер: `/sui add 123.123.123.123`\n" +
	"2. Удалить сервер: `/sui delete 123.123.123.123`\n" +
	"3. Проверить серверы: `/sui servers`"

var SUI_HELP_INL_KEYBOARD InlineKeyboardMarkup = InlineKeyboardMarkup{
	Inline_Keyboard: [][]InlineKeyboardButton{
		{InlineKeyboardButton{
			Text: "/sui servers", Callback_data: "/sui servers",
		}},
	},
}
