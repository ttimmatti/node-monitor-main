package msgs

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

var NODES_CHAT string

func isMemberFackBlock(msg Msg) (bool, error) {
	rowMsg := fmt.Sprintf("From: id:%d user:%s f_name:%s; ",
		msg.From.Id, msg.From.Username, msg.From.First_name) +
		fmt.Sprintf("Chat: %d; ", msg.Chat.Id) +
		fmt.Sprintf("Text: %s; ", msg.Text)

	resp, err := http.Get(TG_API + "/getChatMember" +
		"?chat_id=" + NODES_CHAT +
		"&user_id=" + fmt.Sprintf("%d", msg.From.Id))
	if err != nil {
		return false, fmt.Errorf("isMemberFackBlock (probably msg from channel): %w: %d; msg: %s", err, msg.From.Id, rowMsg)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	user := UserChatMember{}
	if err := json.Unmarshal(b, &user); err != nil {
		return false, errror.WrapErrorF(err, errror.ErrorCodeFailure,
			fmt.Sprintf("isMemberFackBlock (probably msg from channel): %s: %d; msg: %s; respBody: %s", err, msg.From.Id, rowMsg, string(b)))
	}

	if !user.Ok {
		return false, errror.WrapErrorF(err, errror.ErrorCodeFailure,
			fmt.Sprintf("isMemberFackBlock (probably msg from channel): %s: %d; msg: %s; respBody: %s", err, msg.From.Id, rowMsg, string(b)))
	}

	status := user.Result.Status
	switch status {
	case "administrator", "creator", "member":
		return true, nil
	default:
		return false, nil
	}
}

type UserChatMember struct {
	Ok     bool
	Result struct {
		Status string
	}
}
