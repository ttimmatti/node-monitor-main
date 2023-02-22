package msgs

import (
	"fmt"
	"log"
	"strconv"

	"github.com/ttimmatti/nodes-bot/tg-db/db"
	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

const ADDSERVERSCSS = `Your server was succesfully added / Ваш сервер был успешно добавлен`
const PARSE_MODE = "MarkdownV2"

var ADMIN_ID int64

func HandleMsg(msg Msg) {
	//check that the message is from a private chat, if not skip it
	if msg.Chat.Id != msg.From.Id {
		return
	}

	if msg.Text == "" {
		return
	}

	isUser, isBanned := db.UserExistIsBanned(msg.From.Id)
	if isBanned {
		handleError(msg, errror.NewErrorf(
			errror.ErrorCodeIsBanned,
			"handle_cmd_user-banned: (username)", msg.From.Username,
		))
		return
	}
	if !isUser {
		if member, err := isMemberFackBlock(msg); err != nil || !member {
			if err != nil {
				handleError(msg, fmt.Errorf("HandleMsg: %w", err))
			} else {
				err := handleNotMember(msg)
				if err != nil {
					handleError(msg, fmt.Errorf("HandleMsg: %w", err))
				}
			}
			return
		}
	}

	text := msg.Text

	log.Printf("HandleMsg: Received: %s --> %s", msg.From.Username, msg.Text)

	cmd, _, _, err := parseCmd(text)
	if err != nil {
		handleError(msg, fmt.Errorf("HandleMsg: %w", err))
		return
	}

	if errs := handleCmd(IsAdmin(msg.From.Id), isUser, cmd, msg); errs != nil {
		for i := range errs {
			errs[i] = fmt.Errorf("HandleMsg: %w", errs[i])
			handleError(msg, errs[i])
		}
	}
}

func handleNotMember(msg Msg) error {
	cmd, code, _, err := parseCmd(msg.Text)
	if err != nil {
		return fmt.Errorf("handleNotMember: %w", err)
	}
	if cmd != "/code" {
		reply := SendMsg{
			Chat_id:    msg.From.Id,
			Text:       NOTMEMBERTEXT,
			Parse_mode: "HTML",
		}
		err := sendMsg(reply)
		if err != nil {
			return fmt.Errorf("handleNotMember: %w", err)
		}
	}

	if LAST_INVITE_CODE == 0 {
		return errror.NewErrorf(errror.ErrorCodeInvalidArgument,
			"handleNotMember: invite_code not activated", msg.From.Username, "--> "+msg.Text)
	}

	inv_code, _ := strconv.ParseInt(code, 10, 64)
	if inv_code == LAST_INVITE_CODE {
		LAST_INVITE_CODE = 0
		err := db.AddUser(
			msg.From.Id,
			msg.From.Username,
			msg.From.First_name,
			msg.Date,
		)
		if err != nil {
			return fmt.Errorf("handleNotMember: %w", err)
		}

		reply := SendMsg{
			Chat_id:    msg.From.Id,
			Text:       "<b>Code</b> activated. Welcome!",
			Parse_mode: "HTML",
		}
		if err := sendMsg(reply); err != nil {
			return fmt.Errorf("handleNotMember: %w", err)
		}
	}

	return nil
}

func handleCmd(isAdmin, isUser bool, cmd string, msg Msg) []error {
	switch cmd {
	case "/start":
		if err := handleStart(isUser, msg); err != nil {
			return []error{fmt.Errorf("handleCmd: %w", err)}
		}
	case "/servers":
		if err := handleServersMsg(msg); err != nil {
			return []error{fmt.Errorf("handleCmd: %w", err)}
		}
	case "/ironfish":
		if err := handleIronfishMsg(isUser, msg); err != nil {
			return []error{fmt.Errorf("handleCmd: %w", err)}
		}
	case "/sui":
		if err := handleSuiMsg(isUser, msg); err != nil {
			return []error{fmt.Errorf("handleCmd: %w", err)}
		}
	}
	//TODO: if user send wrong cmd return him the error

	if isAdmin {
		if errs := adminCmd(cmd, msg); errs != nil {
			for i := range errs {
				errs[i] = fmt.Errorf("handleCmd: %w", errs[i])
			}
			return errs
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
		{InlineKeyboardButton{
			Text: "/sui", Callback_data: "/sui",
		}},
	},
}

var START_RPL_KEYBOARD ReplyKeyboardMarkup = ReplyKeyboardMarkup{
	Keyboard: [][]KeyBoardButton{
		{KeyBoardButton{Text: "/ironfish"}},
		{KeyBoardButton{Text: "/sui"}}},
	Resize_keyboard:         true,
	Input_field_placeholder: "Choose the command",
}

func IsAdmin(id int64) bool {
	return id == ADMIN_ID
}

var SERVERS_RPL_KEYBOARD ReplyKeyboardMarkup = ReplyKeyboardMarkup{
	Keyboard: [][]KeyBoardButton{
		{KeyBoardButton{Text: "/start"}},
		{
			KeyBoardButton{Text: "/ironfish servers"},
			KeyBoardButton{Text: "/sui servers"},
		},
		{
			KeyBoardButton{Text: "/sui checker"},
			KeyBoardButton{Text: "/servers disk"},
		},
	},
	Resize_keyboard: true,
	Is_persistent:   true,
}

const WELCOME = `<b>Fack Nodes Monitor Bot</b>

Welcome!

To view commands on Ironfish send "/ironfish"
To view commands on Sui send "/sui"

----------------------------------------------

Добро пожаловать!

Для просмотра команд для Ironfish отправьте "/ironfish"
Для просмотра команд для Sui отправьте "/sui"`
const WELCOMEBACK = `<b>Fack Nodes Monitor Bot</b>

Welcome back!

To view commands on Ironfish send "/ironfish"
To view commands on Sui send "/sui"

----------------------------------------------

Рады снова видеть!

Для просмотра команд по Ironfish отправьте "/ironfish"
Для просмотра команд для Sui отправьте "/sui"`
const NOTMEMBERTEXT = `<b>You are not a member of FackBlock Nodes Channel</b>

<i>The bot is only free for the members of 9kdao</i>

<i>For further inquiries you can contact @ttimmatti</i>

-------------------------------------------------------------------------------

<b>Вы не являетесть членом группы FackBlock Nodes</b>

<i>Этот бот является бесплатным только для участников 9kdao</i>

<i>По дальнейшим вопросам вы можете связаться с @ttimmatti</i>`
