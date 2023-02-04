package msgs

import (
	"errors"
	"log"

	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

func handleError(err error, msg Msg) {
	var ierr *errror.Error
	if !errors.As(err, &ierr) {
		log.Println("error wosn't of type Error: ", err)
		return
	}

	log.Println(err)

	switch ierr.Code() {
	case errror.ErrorCodeFailure,
		errror.ErrorCodeUnknown:
		msg := &SendMsg{
			Text:       "errFailure: " + ierr.Error(),
			Chat_id:    ADMIN_ID,
			Parse_mode: "",
		}
		if err := sendMsg(msg); err != nil {
			log.Println("handleError: Couldn't send note about fatal error to admin!!!")
		}
	case errror.ErrorCodeInvalidArgument:
		msg := &SendMsg{
			Text:       "Error. Code: Invalid Argument",
			Chat_id:    msg.From.Id,
			Parse_mode: "",
		}
		if err := sendMsg(msg); err != nil {
			log.Println("handleError: Couldn't send error msg to user!!!")
		}
	case errror.ErrorCodeNotFound:
		msg := &SendMsg{
			Text:       "Error. Code: Not Found",
			Chat_id:    msg.From.Id,
			Parse_mode: "",
		}
		if err := sendMsg(msg); err != nil {
			log.Println("handleError: Couldn't send error msg to user!!!")
		}
	case errror.ErrorCodePongFalse:
		msg := &SendMsg{
			Text:       TEXT_NOPONG,
			Chat_id:    msg.From.Id,
			Parse_mode: "",
		}
		if err := sendMsg(msg); err != nil {
			log.Println("handleError: Couldn't send error msg to user!!!")
		}
	}
}

const TEXT_NOPONG = `Error. Code: Server didn't respond
-----------------------------		
Скорее всего, вы не установили софт на сервер`
