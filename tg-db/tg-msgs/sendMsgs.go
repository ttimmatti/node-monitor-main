package msgs

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

func sendMsg(msg interface{}) error {
	respByte, err := json.Marshal(msg)
	if err != nil {
		return errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"sendMsg_json_marshal_err")
	}

	resp, err := http.Post(TG_API+"/sendMessage", "Content-Type: application/json", bytes.NewBuffer(respByte))
	if err != nil {
		time.Sleep(2 * time.Second)
		resp, err = http.Post(TG_API+"/sendMessage", "Content-Type: application/json", bytes.NewBuffer(respByte))
		if err != nil {
			return errror.WrapErrorF(err,
				errror.ErrorCodeFailure,
				"sendMsg_post_msg")
		}
	}

	defer resp.Body.Close()

	// remove for prod
	response, _ := io.ReadAll(resp.Body)
	var result map[string]bool
	json.Unmarshal(response, &result)
	ok := result["ok"]
	log.Println("sendMsg: ok:", ok)

	if !ok {
		return errror.NewErrorf(errror.ErrorCodeUnknown,
			"sendMsg: ok return false. TG response: "+string(response)+"\nPosted: "+string(respByte))
	}

	return nil
}

func DefaultReply(chat_id int64, text, parse_mode string) *SendMsg {
	msg := &SendMsg{
		Chat_id:    chat_id,
		Text:       text,
		Parse_mode: parse_mode,
	}

	return msg
}

func ReplyWithInlKey(chat_id int64, text, parse_mode string, keyboard InlineKeyboardMarkup) *SendMsgInlKeyboard {
	msg := &SendMsgInlKeyboard{
		Chat_id:      chat_id,
		Text:         text,
		Parse_mode:   parse_mode,
		Reply_markup: keyboard,
	}

	return msg
}

func ReplyWithReplyKeyboard(chat_id int64, text, parse_mode string, keyboard ReplyKeyboardMarkup) *SendMsgRplKeyboard {
	msg := &SendMsgRplKeyboard{
		Chat_id:      chat_id,
		Text:         text,
		Parse_mode:   parse_mode,
		Reply_markup: keyboard,
	}

	return msg
}

type SendMsgInlKeyboard struct {
	Chat_id      int64                `json:"chat_id"`
	Text         string               `json:"text"`
	Parse_mode   string               `json:"parse_mode"`
	Reply_markup InlineKeyboardMarkup `json:"reply_markup"`
}

type SendMsg struct {
	Chat_id    int64  `json:"chat_id"`
	Text       string `json:"text"`
	Parse_mode string `json:"parse_mode"`
}

type SendMsgRplKeyboard struct {
	Chat_id      int64               `json:"chat_id"`
	Text         string              `json:"text"`
	Parse_mode   string              `json:"parse_mode"`
	Reply_markup ReplyKeyboardMarkup `json:"reply_markup"`
}

//TODO: implement func send error
