package msgs

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/ttimmatti/ironfish-node-tg/errror"
)

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

func sendStartMsg(id int64) error {
	response := &SendMsg{
		id,
		WELCOME,
		PARSE_MODE,
	}

	// response1 := map[string]string{
	// 	"chat_id":    fmt.Sprintf("%d", msg.From.Id),
	// 	"text":       START_MSG,
	// 	"parse_mode": PARSE_MODE,
	// }

	if err := sendMsg(response); err != nil {
		return err
	}

	return nil
}

//TODO: implement func send error

type SendMsg struct {
	Chat_id    int64  `json:"chat_id"`
	Text       string `json:"text"`
	Parse_mode string `json:"parse_mode"`
}
