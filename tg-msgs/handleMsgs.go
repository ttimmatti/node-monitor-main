package msgs

import (
	"fmt"
	"log"

	"github.com/ttimmatti/ironfish-node-tg/db"
	"github.com/ttimmatti/ironfish-node-tg/errror"
)

const WELCOME = `*Fack Nodes IronFish Monitor Bot*`
const WELCOMEBACK = `Oh, welcome back / Давно не виделись
*Fack Nodes IronFish Monitor Bot*`
const ADDSERVERSCSS = `Your server was succesfully added / Ваш сервер был успешно добавлен`
const PARSE_MODE = "MarkdownV2"

var ADMIN_ID int64

func HandleMsg(msg Msg) {
	text := msg.Text

	log.Println("\nIncoming Message:\n", "User:", msg.From.Username,
		"\nText:", msg.Text)

	//TODO: add banned users table. if msg is from one, return err

	//check that the message is from a private chat, if not skip it
	//TODO:

	cmd, _, _, err := parseCmd(text)
	if err != nil {
		//TODO: return me error
		log.Println(err)
		return
	}

	handleCmd(IsAdmin(msg.From.Id), cmd, msg)

	//TODO: make separate handler for admin
}

func handleCmd(isAdmin bool, cmd string, msg Msg) error {
	if isAdmin {
		if err := adminCmd(cmd, msg); err != nil {
			return errror.WrapErrorF(err,
				errror.ErrorCodeFailure,
				"admin_cmd_execution")
		}
		return nil
	}

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
			log.Println(err)
		}
	case "/add":
		if err := handleUserAddServer(isUser, msg); err != nil {
			log.Println(err)
		}
	case "/servers":
		if err := handleUserServers(isUser, msg); err != nil {
			log.Println(err)
		}
	case "/change":
		if err := handleUserChangeServer(isUser, msg); err != nil {
			log.Println(err)
		}
	case "/delete":
		if err := handleUserDeleteServer(isUser, msg); err != nil {
			log.Println(err)
		}
	}
	//TODO: if user send wrong cmd return him the error

	return nil
}

func handleUserDeleteServer(isUser bool, msg Msg) error {
	if !isUser {
		//TODO: sendmessage "you're not a user, please type start to start..."
		// and guide WELCOME
		// and return
		respMsg := &SendMsg{
			msg.From.Id,
			"You're not a user",
			PARSE_MODE,
		}

		err := sendMsg(respMsg)
		if err != nil {
			log.Println("couldnt send msg")
			return err
		}

		return fmt.Errorf("user not registered")
	}

	//parse cmd
	_, server_ip1, _, err := parseCmd(msg.Text)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("not a valid cmd")
	}

	//check that the cmd is valid
	isValid1 := validServerIp(server_ip1)
	if !isValid1 {
		return fmt.Errorf("server ip not in valid format")
	}

	//it can be that the server is already in the db
	// tell the user to check servers and verify that it's not there
	// or contact me
	if err := db.DeleteUserServer(msg.From.Id, server_ip1); err != nil {
		return err
	}

	// if rows affected = 1 --> success
	// if rows affected = 0 --> no changes

	respMsg := &SendMsg{
		msg.From.Id,
		"succesfully deleted server ip", //TODO: not welcome back
		PARSE_MODE,
	}
	if err := sendMsg(respMsg); err != nil {
		return err
	}

	return nil
}

func handleUserChangeServer(isUser bool, msg Msg) error {
	if !isUser {
		//TODO: sendmessage "you're not a user, please type start to start..."
		// and guide WELCOME
		// and return
		respMsg := &SendMsg{
			msg.From.Id,
			"You're not a user",
			PARSE_MODE,
		}

		err := sendMsg(respMsg)
		if err != nil {
			log.Println("couldnt send msg")
			return err
		}

		return fmt.Errorf("user not registered")
	}

	//parse cmd
	_, server_ip1, server_ip2, err := parseCmd(msg.Text)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("not a valid cmd")
	}

	//check that the cmd is valid
	isValid1 := validServerIp(server_ip1)
	isValid2 := validServerIp(server_ip2)
	if !isValid1 || !isValid2 {
		return fmt.Errorf("server ip not in valid format")
	}

	//it can be that the server is already in the db
	// tell the user to check servers and verify that it's not there
	// or contact me
	if err := db.ChangeUserServer(msg.From.Id,
		server_ip1,
		server_ip2); err != nil {
		return err
	}

	respMsg := &SendMsg{
		msg.From.Id,
		"succesfully updated server ip", //TODO: not welcome back
		PARSE_MODE,
	}
	if err := sendMsg(respMsg); err != nil {
		return err
	}

	return nil
}

func handleUserServers(isUser bool, msg Msg) error {
	if isUser {
		servers, err := db.GetUserServers(msg.From.Id)
		if err != nil {
			log.Println(err)
			return err
		}

		msg := &SendMsg{
			msg.From.Id,
			"SERVERS: \n" + servers,
			PARSE_MODE,
		}

		if err := sendMsg(msg); err != nil {
			log.Println(err)
		}
	} else {
		msg := &SendMsg{
			msg.From.Id,
			"poshel nahui, ti ne zaregan",
			PARSE_MODE,
		}

		if err := sendMsg(msg); err != nil {
			log.Println(err)
		}
	}

	return nil
}

// ready
func handleUserAddServer(isUser bool, msg Msg) error {
	//parse cmd
	_, server_ip, _, err := parseCmd(msg.Text)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("not a valid cmd")
	}

	//check that the cmd is valid
	isValid := validServerIp(server_ip)
	if !isValid {
		return fmt.Errorf("server ip not in valid format")
	}

	if !isUser {
		err := db.AddUser(msg.From.Id,
			msg.From.Username,
			msg.From.First_name,
			msg.Date)
		if err != nil {
			log.Println(err)
		}
	}

	//it can be that the server is already in the db
	// tell the user to check servers and verify that it's not there
	// or contact me
	if err := db.AddUserServer(msg.From.Id, server_ip); err != nil {
		return err
	}

	respMsg := &SendMsg{
		msg.From.Id,
		"succesfully added server", //TODO: not welcome back
		PARSE_MODE,
	}
	if err := sendMsg(respMsg); err != nil {
		log.Println(err)
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
		msg := &SendMsg{
			msg.From.Id,
			WELCOMEBACK,
			PARSE_MODE,
		}
		if err := sendMsg(msg); err != nil {
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

		msg := &SendMsg{
			msg.From.Id,
			WELCOME,
			PARSE_MODE,
		}
		if err := sendMsg(msg); err != nil {
			log.Println(err)
		}
	}

	return nil
}

func IsAdmin(id int64) bool {
	if id != ADMIN_ID {
		return false
	}

	return true
}
