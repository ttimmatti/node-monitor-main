package msgs

import (
	"fmt"
	"log"

	"github.com/ttimmatti/nodes-bot/tg-db/db"
	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

const WELCOME = `<b>Fack Nodes Monitor Bot</b>

Welcome!

To view commands on Ironfish send "/ironfish"

----------------------------------------------

Добро пожаловать!

Для просмотра команд для Ironfish отправьте "/ironfish"`
const WELCOMEBACK = `<b>Fack Nodes Monitor Bot</b>

Welcome back!

To view commands on Ironfish send "/ironfish"

----------------------------------------------

Рады снова видеть!

Для просмотра команд по Ironfish отправьте "/ironfish"`
const ADDSERVERSCSS = `Your server was succesfully added / Ваш сервер был успешно добавлен`
const PARSE_MODE = "MarkdownV2"

var ADMIN_ID int64

func HandleMsg(msg Msg) {
	text := msg.Text

	log.Printf("HandleMsg: Received: %s --> %s", msg.From.Username, msg.Text)

	//check that the message is from a private chat, if not skip it
	if msg.Chat.Id != msg.From.Id {
		log.Printf("User from group tried accessing the bot: User - %s, ChatId - %d",
			msg.From.Username, msg.Chat.Id)
		return
	}

	cmd, _, _, err := parseCmd(text)
	if err != nil {
		//TODO: return me error
		log.Println("HandleMsg:", err)
		return
	}

	if err := handleCmd(IsAdmin(msg.From.Id), cmd, msg); err != nil {
		handleError(err, msg)
	}
}

func handleCmd(isAdmin bool, cmd string, msg Msg) error {
	isUser, isBanned := db.UserExistIsBanned(msg.From.Id)
	if isBanned {
		return errror.NewErrorf(
			errror.ErrorCodeIsBanned,
			"handle_cmd_user-banned: (username)", msg.From.Username,
		)
	}

	switch cmd {
	case "/start":
		if err := handleStart(isUser, msg); err != nil {
			return fmt.Errorf("handleCmd: %w", err)
		}
	case "/ironfish":
		if err := handleIronfishMsg(isUser, msg); err != nil {
			return fmt.Errorf("handleCmd: %w", err)
		}
	}
	//TODO: if user send wrong cmd return him the error

	if isAdmin {
		if err := adminCmd(cmd, msg); err != nil {
			return errror.WrapErrorF(err,
				errror.ErrorCodeFailure,
				"admin_cmd_execution:")
		}
	}

	return nil
}

// ready
func handleStart(isUser bool, msg Msg) error {
	// check if the user exist in db
	// if yes send "Oh, welcome back / Давно не виделись"
	// and startingGuide in another msg
	// if no Starting guide

	if isUser {
		msg2 := ReplyWithInlKey(msg.From.Id, WELCOMEBACK, "HTML", START_INL_KEYBOARD)
		if err := sendMsg(msg2); err != nil {
			log.Println(err)
		}
	} else {
		err := db.AddUser(msg.From.Id,
			msg.From.Username,
			msg.From.First_name,
			msg.Date)
		if err != nil {
			log.Println(err)
		}

		msg2 := ReplyWithInlKey(msg.From.Id, WELCOME, "HTML", START_INL_KEYBOARD)
		if err := sendMsg(msg2); err != nil {
			log.Println(err)
		}
	}

	return nil
}

var START_INL_KEYBOARD InlineKeyboardMarkup = InlineKeyboardMarkup{
	Inline_Keyboard: [][]InlineKeyboardButton{
		{InlineKeyboardButton{
			Text: "/ironfish", Callback_data: "/ironfish",
		}},
	},
}

var START_RPL_KEYBOARD ReplyKeyboardMarkup = ReplyKeyboardMarkup{
	Keyboard: [][]KeyBoardButton{{
		KeyBoardButton{Text: "/ironfish"}}},
	Resize_keyboard:         true,
	Input_field_placeholder: "Choose the command",
}

func IsAdmin(id int64) bool {
	return id == ADMIN_ID
}
