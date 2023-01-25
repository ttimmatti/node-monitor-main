package msgs

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var TG_API string
var LAST_MSG_INDEX int64

func StartReceiving(tg_api string, update_sec int64) {
	TG_API = tg_api
	for {
		// update every $n secs
		time.Sleep(time.Duration(update_sec) * time.Second)

		// get messages
		msgs := checkMsgs(LAST_MSG_INDEX) //TODO: what if error sending response and etc. bulletproof
		if len(msgs) < 1 {
			// if empty then sleep
			continue
		}

		for _, msg_m := range msgs {
			msg := msg_m.Message
			HandleMsg(msg)
		}
	}
}

func checkMsgs(index int64) []MsgUpd {
	response, err := http.Get(TG_API + "/getUpdates" + "?offset=" + fmt.Sprintf("%d", index+1))
	if err != nil {
		log.Println((err))
	}
	defer response.Body.Close()

	rBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Println(err)
	}

	update := &UpdatesResponse{}
	json.Unmarshal(rBody, update)

	messages := update.Result

	msgs_ln := len(messages)
	if msgs_ln > 0 {
		last_message := messages[msgs_ln-1]
		LAST_MSG_INDEX = last_message.Update_id
	}

	return messages
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
	Update_id int64
	Message   Msg
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
