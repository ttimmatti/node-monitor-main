package node_worker

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/ttimmatti/nodes-bot/ironfish/errror"
)

var TG_API string

func sendMsg(msg *SendMsg) error {
	msg.Disable_web_page_preview = true

	respByte, err := json.Marshal(msg)
	if err != nil {
		return errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"sendMsg_json_marshal_err")
	}

	resp, err := http.Post(TG_API+"/sendMessage", "Content-Type: application/json", bytes.NewBuffer(respByte))
	if err != nil {
		return errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"sendMsg_post_msg")
	}

	defer resp.Body.Close()

	// remove for prod
	response, _ := io.ReadAll(resp.Body)
	log.Println("\nPosted to tg:", string(respByte))
	log.Println("\nResponse from tg:", string(response))

	//TODO: READ MSG, IF "OK":FALSE return error

	return nil
}

func sendAnyMsg(msg interface{}) error {
	respByte, err := json.Marshal(msg)
	if err != nil {
		return errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"sendMsg_json_marshal_err")
	}

	resp, err := http.Post(TG_API+"/sendMessage", "Content-Type: application/json", bytes.NewBuffer(respByte))
	if err != nil {
		return errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"sendMsg_post_msg")
	}

	defer resp.Body.Close()

	// remove for prod
	response, _ := io.ReadAll(resp.Body)
	log.Println("\nResponse from tg:", string(response))

	//TODO: READ MSG, IF "OK":FALSE return error

	return nil
}

type SendMsg struct {
	Chat_id                  int64  `json:"chat_id"`
	Text                     string `json:"text"`
	Parse_mode               string `json:"parse_mode"`
	Disable_web_page_preview bool   `json:"disable_web_page_preview"`
}

type SendMsgInlKeyboard struct {
	Chat_id      int64                `json:"chat_id"`
	Text         string               `json:"text"`
	Parse_mode   string               `json:"parse_mode"`
	Reply_markup InlineKeyboardMarkup `json:"reply_markup"`
}
type InlineKeyboardMarkup struct {
	Inline_Keyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}
type InlineKeyboardButton struct {
	Text string `json:"text"`
	// returned command
	Callback_data string `json:"callback_data"`
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
