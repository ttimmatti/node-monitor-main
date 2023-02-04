package msgs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

var TG_API string
var LAST_MSG_INDEX int64

func StartReceiving(tg_api string, update_sec int64) {
	TG_API = tg_api
	for {
		// update every $n secs
		time.Sleep(time.Duration(update_sec) * time.Second)

		// get messages
		updates, err := checkMsgs(LAST_MSG_INDEX) //TODO: what if error sending response and etc. bulletproof
		if len(updates) < 1 || err != nil {
			if err != nil {
				log.Println(err)
			}
			// if empty then sleep
			continue
		}

		for _, update := range updates {
			//define if update is a message or a callback query
			updateType := updateType(update)
			msg := Msg{}
			if updateType == IsMessage {
				msg = update.Message
			} else {
				cbq := update.Callback_query
				msg = Msg{
					From: cbq.From,
					Chat: struct{ Id int64 }{cbq.From.Id},
					Text: cbq.Data,
				}
				answerCBQ(cbq.Id)
			}
			HandleMsg(msg)
		}
	}
}

func answerCBQ(i int64) {
	resp, err := http.Get(TG_API + "/answerCallbackQuery?callback_query_id=" + fmt.Sprintf("%d", i))
	if err != nil {
		log.Println("answerCallbackQuery: ", err)
	}
	defer resp.Body.Close()

	io.ReadAll(resp.Body)
}

func checkMsgs(index int64) ([]MsgUpd, error) {
	response, err := http.Get(TG_API + "/getUpdates" + "?offset=" + fmt.Sprintf("%d", index+1))
	if err != nil {
		return nil, errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"checkMsgs:")
	}
	defer response.Body.Close()

	rBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"checkMsgs: ")
	}

	update := &UpdatesResponse{}
	json.Unmarshal(rBody, update)

	updates := update.Result

	msgs_ln := len(updates)
	if msgs_ln > 0 {
		last_message := updates[msgs_ln-1]
		LAST_MSG_INDEX = last_message.Update_id
	}

	return updates, nil
}

func getLastMsgIndex() int64 {
	responseU, err := http.Get(TG_API + "/getUpdates")
	if err != nil {
		log.Println(err)
	}
	defer responseU.Body.Close()

	rBody, err := io.ReadAll(responseU.Body)
	if err != nil {
		log.Println(err)
	}

	update := &UpdatesResponse{}
	json.Unmarshal(rBody, update)

	messages := update.Result
	msgs_ln := len(messages)
	if msgs_ln > 0 {
		last_message := messages[msgs_ln-1]
		return last_message.Update_id
	}

	return 0
}

// //////////////////////////////////////////////////////////
// //////////////////////////////////////////////////////////
// required types
type UpdatesResponse struct {
	Ok     bool
	Result []MsgUpd
}
type MsgUpd struct {
	Update_id      int64
	Message        Msg
	Callback_query CBQ
}
type Msg struct {
	Message_id int64
	From       struct {
		Id         int64
		First_name string
		Username   string
	}
	Chat struct {
		Id int64
	}
	Date     int64
	Text     string
	Entities struct {
		Offset int64
		Length int64
		Type   string
	}
}
type CBQ struct {
	Id   int64
	From struct {
		Id         int64
		First_name string
		Username   string
	}
	Data string
}

type InlineKeyboardMarkup struct {
	Inline_Keyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}
type InlineKeyboardButton struct {
	Text string `json:"text"`
	// returned command
	Callback_data string `json:"callback_data"`
}

type ReplyKeyboardMarkup struct {
	Keyboard                [][]KeyBoardButton `json:"keyboard"`
	Resize_keyboard         bool               `json:"resize_keyboard"`
	One_time_keyboard       bool               `json:"one_time_keyboard"`
	Input_field_placeholder string             `json:"input_field_placeholder"`
}
type KeyBoardButton struct {
	Text string `json:"text"`
}

const (
	IsMessage = iota
	IsCBQ
)

// is it a message or a callback query
func updateType(update MsgUpd) int {
	if update.Callback_query.Data == "" {
		return IsMessage
	} else {
		return IsCBQ
	}
}
