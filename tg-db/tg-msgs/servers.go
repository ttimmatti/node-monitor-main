package msgs

import (
	"fmt"
	"strings"

	"github.com/ttimmatti/nodes-bot/tg-db/db"
	"github.com/ttimmatti/nodes-bot/tg-db/disk"
	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

//for servers health monitor. should be in the other package probably

func handleServersMsg(msg Msg) error {
	_, cmd2, cmd3, err := parseCmd(msg.Text)
	if err != nil {
		return fmt.Errorf("handleServersMsg: %w", err)
	}

	switch cmd2 {
	case "disk":
		if cmd3 == "" {
			if err := serversDiskStart(msg); err != nil {
				return fmt.Errorf("handleServersMsg: %w", err)
			}
			return nil
		}
		if err := serversDiskCheck(msg); err != nil {
			return fmt.Errorf("handleServersMsg: %w", err)
		}
	}
	return nil
}

func serversDiskStart(msg Msg) error {
	reply := ReplyWithInlKey(
		msg.From.Id,
		"Choose the database for the command / Выберите сервера для которых выполнить команду",
		"",
		DISK_CHOICE_INL_KEYBOARD,
	)
	if err := sendMsg(reply); err != nil {
		return fmt.Errorf("serversDisk: %w", err)
	}
	return nil
}

var DISK_CHOICE_INL_KEYBOARD = InlineKeyboardMarkup{
	Inline_Keyboard: [][]InlineKeyboardButton{
		{InlineKeyboardButton{Text: "Sui", Callback_data: "/servers_disk_sui"}},
		{InlineKeyboardButton{Text: "IronFish", Callback_data: "/servers_disk_ironfish"}},
	},
}

func serversDiskCheck(msg Msg) error {
	_, _, node_db, err := parseCmd(msg.Text)
	if err != nil {
		return fmt.Errorf("serversDisk: %w", err)
	}

	var s_ipsF string
	if node_db == "sui" {
		s_ipsF, err = db.SuiGetUserIps(msg.From.Id)
		if err != nil {
			return fmt.Errorf("serversDiskCheck: %w", err)
		}
	} else if node_db == "ironfish" {
		s_ipsF, err = db.IronGetUserIps(msg.From.Id)
		if err != nil {
			return fmt.Errorf("serversDiskCheck: %w", err)
		}
	} else {
		return errror.NewErrorf(
			errror.ErrorCodeInvalidArgument,
			"serversDiskCheck: cmd3 not sui nor ironfish",
		)
	}

	s_ips := strings.Split(s_ipsF, ";;")

	resp := disk.GetDiskSpaceForSs(s_ips)

	reply := DefaultReply(msg.Chat.Id, resp, PARSE_MODE)
	if err := sendMsg(reply); err != nil {
		return fmt.Errorf("serversDiskCheck: %w", err)
	}
	return nil
}
