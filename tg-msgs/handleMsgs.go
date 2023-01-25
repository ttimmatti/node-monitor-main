package msgs

import (
	"fmt"
	"log"

	"github.com/ttimmatti/ironfish-node-tg/db"
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

	//check that the message is from a private chat, if not skip it
	//TODO:

	cmd, _, _, err := parseCmd(text)
	if err != nil {
		//TODO: return me error
		log.Println(err)
		return
	}

	//TODO: make separate handler for admin

	switch cmd {
	case "/start":
		if err := handleStart(msg); err != nil {
			log.Println(err)
		}
	case "/add":
		if err := handleAddServer(msg); err != nil {
			log.Println(err)
		}
	case "/servers":
		if err := handleServers(msg); err != nil {
			log.Println(err)
		}
	case "/change":
		if err := handleChange(msg); err != nil {
			log.Println(err)
		}
	case "/delete":
		handleDelete(msg)
	}
	//TODO: if user send wrong cmd return him the error
}

func handleDelete(msg Msg) error {
	log.Println("not implemendted")
	return nil
}

func handleChange(msg Msg) error {
	isUser := db.UserExist(msg.From.Id)
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

func handleServers(msg Msg) error {
	isUser := db.UserExist(msg.From.Id)
	if isUser {
		servers, err := db.GetServers(msg.From.Id)
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
func handleAddServer(msg Msg) error {
	isUser := db.UserExist(msg.From.Id)

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
func handleStart(msg Msg) error {
	// check if the user exist in db
	// if yes send "Oh, welcome back / Давно не виделись"
	// and startingGuide in another msg
	// if no Starting guide

	isUser := db.UserExist(msg.From.Id)

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
