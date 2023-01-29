package node_worker

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ttimmatti/ironfish-node-tg/db"
	"github.com/ttimmatti/ironfish-node-tg/errror"
)

const REPEAT_MIN = 1

var SERVERS []Server
var WG_DialServers sync.WaitGroup

var LAST_BLOCK int64
var LAST_VERSION string

func Start() {
	//checks the list of servers that pong me back
	// save the list

	//get ironfish status response from every server
	//get block from explorer
	// compare latest blocks and versions

	//also check version through their github releases

	for i := 0; ; i++ {
		if err := handleServers(i); err != nil {
			log.Println("FATAL!!! START: ", err)
		}
		time.Sleep(REPEAT_MIN * time.Minute)
	}
}

func handleServers(i int) error {
	//it should be wrapped in cycle.
	// get initial list of servers
	result, err := db.ReadServers()
	if err != nil {
		return fmt.Errorf("node_worker_handleServers: %w", err)
	}

	servers, err := GetServers(result)
	if err != nil {
		return fmt.Errorf("node_worker_handleServers: %w", err)
	}

	last_block, err := GetLastNetworkBlock()
	if err == nil {
		LAST_BLOCK = last_block
	}
	last_version, err := GetLastNetworkVersion()
	if err == nil {
		LAST_VERSION = last_version
	}

	//check Ping
	servers = FilterPing(servers)

	//clear servers before append
	SERVERS = []Server{}
	// pushes results to SERVERS and db
	DialServers(servers)

	//compare SERVERS bl, v to last bl, v and get a list of servers to send notys
	//do not proceed to sending notys if either last bl/v is empty
	if LAST_BLOCK < 2 || len(LAST_VERSION) < 2 {
		return errror.NewErrorf(
			errror.ErrorCodeFailure,
			fmt.Sprintf("handle_servers: before send msgs: last bl/v are empty. cycle N%d", i),
		)
	}
	NotifyBadServers(SERVERS, LAST_BLOCK, LAST_VERSION)

	return nil
}

// calls 10 servers from SERVERS_TO_CALL starting from $index
// and updates their
func DialServers(servers []Server) {
	//this should be done asynchorously
	//for all pinged servers check versions and last blocks, write response to interface

	WG_DialServers.Add(len(servers))

	for i, s := range servers {
		log.Println("deploying go N", i)
		go UpdateSERVER(i, s)
	}

	WG_DialServers.Wait()
}

func UpdateSERVER(i int, s Server) {
	defer WG_DialServers.Done()

	server, err := s.Update()
	if err != nil {
		log.Println("goroutine N", i, "; error:", err)
	} else {
		//pushes server to SERVERS
		SERVERS = append(SERVERS, server)
		log.Println("updates server. go N", i, "finished succesully")
	}
}

func NotifyBadServers(servers []Server, last_block int64, last_version string) {
	for _, s := range servers {
		if s.Last_block < last_block-10 {
			//sendMsg to owner that server is not synced
			chat_id, _ := strconv.ParseInt(s.Owner_id, 10, 64)
			msg := &SendMsg{
				Chat_id: chat_id,
				Text: fmt.Sprintf(
					"#ironfish\n\nHey! Your server: %s is not synced.\n\nYour server last_block: %d, Network last_block: %d", s.Ip, s.Last_block, last_block,
				) + "\n\n-----------------------------------------------------------" + fmt.Sprintf(
					"\n\nЙоу! Твой сервер: %s отстал от сети.\n\nПоследний блок на этом сервере: %d, Последний блок в сети: %d", s.Ip, s.Last_block, last_block,
				),
				Parse_mode: "",
			}
			sendMsg(msg)
		}
		if s.Version != last_version {
			//sendMsg to owner that server is version behind and needs to update with cmd
			chat_id, _ := strconv.ParseInt(s.Owner_id, 10, 64)
			msg := &SendMsg{
				Chat_id: chat_id,
				Text: fmt.Sprintf(
					"#ironfish\n\nHey! Your server: %s has old IronFish version, there is an update.\n\nYour server version: %s, Last version: %s\n\nYou can update your node via our <a href=\"https://fackblock.com/YcwHs6EhOX-#72AU\">Guide</a>", s.Ip, s.Version, last_version,
				) + "\n\n-----------------------------------------------------------" + fmt.Sprintf(
					"\n\nЙоу! Версия Ironfish на твоем сервере: %s устарела.\n\nВерсия на этом сервере: %s, Последняя версия: %s\n\nМожешь обновиться командой из раздела гайда <a href=\"https://fackblock.com/YcwHs6EhOX-#72AU\">Обновление</a>", s.Ip, s.Version, last_version,
				),
				Parse_mode: "HTML",
			}
			sendMsg(msg)
		}
		//server is working fine
	}
}

var FILTERED_SERVERS []Server
var WG_FilterServers sync.WaitGroup

func FilterPing(servers []Server) []Server {
	FILTERED_SERVERS = []Server{}

	WG_FilterServers.Add(len(servers))

	log.Println("node_worker: ping servers")
	for i, server := range servers {
		go tryPingServer(i, server)
	}

	WG_FilterServers.Wait()

	return FILTERED_SERVERS
}

func tryPingServer(i int, s Server) {
	defer WG_FilterServers.Done()

	err := s.Ping()
	if err == nil {
		// append to new arr
		FILTERED_SERVERS = append(FILTERED_SERVERS, s)
		log.Printf("go: tryPing: ping to %s suceeded", s.Ip)
	} else {
		log.Printf("go: tryPing: ping to %s failed", s.Ip)
	}
}

//

func GetServers(input string) ([]Server, error) {
	lines := strings.Split(input, ":;")

	var servers []Server

	for _, line := range lines {
		serverFields := strings.Split(line, ";;")

		if len(serverFields) < 3 {
			return nil, errror.NewErrorf(
				errror.ErrorCodeFailure,
				"getservers_not-enough-args",
			)
		}

		servers = append(servers, Server{
			Owner_id:   serverFields[1],
			Ip:         serverFields[2],
			Last_block: 0,
			Version:    "",
		})
	}

	return servers, nil
}
