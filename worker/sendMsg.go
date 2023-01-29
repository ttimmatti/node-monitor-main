package node_worker

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/ttimmatti/ironfish-node-tg/errror"
)

const TG_API = "https://api.telegram.org/bot5864005496:AAFYPu4VK53PD8rjmrMyFfIpnyaiCnQASeo"

func sendMsg(msg *SendMsg) error {
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

type SendMsg struct {
	Chat_id    int64  `json:"chat_id"`
	Text       string `json:"text"`
	Parse_mode string `json:"parse_mode"`
}
