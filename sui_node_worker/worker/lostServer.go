package node_worker

import (
	"fmt"
	"strconv"
	"time"

	"github.com/ttimmatti/nodes-bot/sui/db"
)

const LOST_SERVER_TEXT = `#sui

<b>Your Sui node last responce was more than 12 hours ago.</b>

Consider checking your server to clarify what's the problem.

If you want to delete this server from the bot use the button below the message.

----------------------------------------------

<b>Ваш сервер Sui не отвечает уже более 12 часов.</b>

Советуем проверить сервер и попытаться найти проблему.

Если вы хотите удалить этот сервер из бота, используйте кнопку снизу.`

func filterLost() []error {
	errs := []error{}

	result, err := db.SuiReadServers()
	if err != nil {
		return []error{fmt.Errorf("node_worker_handleServers: %w", err)}
	}
	servers, err := GetServers(result)
	if err != nil {
		return []error{fmt.Errorf("node_worker_handleServers: %w", err)}
	}

	t := time.Now().Unix()
	for _, s := range servers {
		if s.LastPong < t-3*60*60 {
			//update db
			var updated bool
			if s.Version == LAST_VERSION {
				updated = true
			}
			if err := db.UpdateSyncUpdServer(s.Ip, s.Status, false, updated); err != nil {
				errs = append(errs, fmt.Errorf("server_update-in-db: %w", err))
			}

			//send Msg
			chat_id, _ := strconv.ParseInt(s.Owner_id, 10, 64)
			msg := ReplyWithInlKey(chat_id, LOST_SERVER_TEXT, "HTML",
				InlKeyboard_DeleteThisServer(s.Ip))
			err := sendAnyMsg(msg)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

func InlKeyboard_DeleteThisServer(ip string) InlineKeyboardMarkup {
	return InlineKeyboardMarkup{Inline_Keyboard: [][]InlineKeyboardButton{
		{InlineKeyboardButton{
			Text: "/sui delete " + ip, Callback_data: "/sui delete " + ip,
		}},
	}}
}
