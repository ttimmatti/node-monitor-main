package node_worker

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ttimmatti/nodes-bot/ironfish/db"
	"github.com/ttimmatti/nodes-bot/ironfish/errror"
)

const REPEAT_MIN = 5

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
		if errs := handleServers(i); errs != nil {
			log.Println("iron_node_worker: START: ", errs)
		}
		time.Sleep(REPEAT_MIN * time.Minute)
	}
}

func handleServers(i int) []error {
	errs := []error{}
	//it should be wrapped in cycle.
	// get initial list of servers
	result, err := db.IronReadServers()
	if err != nil {
		return []error{fmt.Errorf("node_worker_handleServers: %w", err)}
	}

	servers, err := GetServers(result)
	if err != nil {
		return []error{fmt.Errorf("node_worker_handleServers: %w", err)}
	}

	last_block, err := GetLastNetworkBlock()
	if err == nil { // if error IS nil
		LAST_BLOCK = last_block
	} else {
		errs = append(errs, err)
	}
	last_version, err := GetLastNetworkVersion()
	if err == nil { // if error IS nil
		LAST_VERSION = last_version
	} else {
		errs = append(errs, err)
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
		errs = append(errs,
			errror.NewErrorf(
				errror.ErrorCodeFailure,
				fmt.Sprintf("handle_servers: before send msgs: last bl/v are empty. cycle N%d", i)))
		return errs
	}

	if ers := DefineNotifyBadServers(SERVERS, LAST_BLOCK, LAST_VERSION); errs != nil {
		errs = append(errs, ers...)
	}

	if ers := UpdateDbServers(SERVERS); errs != nil {
		errs = append(errs, ers...)
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

func UpdateDbServers(SERVERS []Server) []error {
	errs := []error{}
	for _, s := range SERVERS {
		err := s.UpdateInDb()
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func DialServers(servers []Server) {
	//this should be done asynchorously
	//for all pinged servers check versions and last blocks, write response to interface

	if len(servers) == 0 {
		log.Println("no servers to dial")
		return
	}

	WG_DialServers.Add(len(servers))

	for i, s := range servers {
		log.Printf("Ponged. Try update server: %s, n: %d", s.Ip, i)
		go UpdateSERVER(i, s)
	}

	log.Println("node_worker: DialServers: Deployed goroutines to update servers. Waiting for responses")

	WG_DialServers.Wait()
}

func UpdateSERVER(i int, s Server) {
	defer WG_DialServers.Done()

	server, err := s.GetUpdate()
	if err != nil {
		log.Println("goroutine N", i, "; error:", err)
	} else {
		//pushes server to SERVERS
		SERVERS = append(SERVERS, server)
		log.Printf("Success updating server: %s, n: %d", s.Ip, i)
	}
}

func DefineNotifyBadServers(servers []Server, last_block int64, last_version string) []error {
	errs := []error{}
	for i, s := range servers {
		if s.Last_block < last_block-10 {
			if err := notifySBadSync(s, last_block); err != nil {
				errs = append(errs, err)
			}
			SERVERS[i].Synced = false
		} else {
			for i := range SERVERS {
				if s.Ip == SERVERS[i].Ip {
					SERVERS[i].Synced = true
				}
			}
		}

		if s.Version != last_version {
			if err := notifySBadVersion(s, last_version); err != nil {
				errs = append(errs, err)
			}
			SERVERS[i].Updated = false
		} else {
			for i := range SERVERS {
				if s.Ip == SERVERS[i].Ip {
					SERVERS[i].Updated = true
				}
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

func notifySBadVersion(s Server, last_version string) error {
	//check if server was not updated before that lap
	// if wasn't updated than don't notify, because we already did notify
	if !s.Updated {
		return nil
	}

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

	err := sendMsg(msg)
	if err != nil {
		return fmt.Errorf("noty_bad_v: %w", err)
	}
	return nil
}

func notifySBadSync(s Server, last_block int64) error {
	//check if server was not synced before that lap
	// if wasn't synced than don't notify, because we already did notify
	if !s.Synced {
		return nil
	}

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

	err := sendMsg(msg)
	if err != nil {
		return fmt.Errorf("noty_bad_sync: %w", err)
	}
	return nil
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

		if len(serverFields) < 8 {
			return nil, errror.NewErrorf(
				errror.ErrorCodeFailure,
				"getservers_not-enough-args",
			)
		}

		owner_id := serverFields[1]
		ip := serverFields[2]
		syncedS := serverFields[5]
		synced := false
		if syncedS == "true" {
			synced = true
		}
		updatedS := serverFields[6]
		updated := false
		if updatedS == "true" {
			updated = true
		}
		lastPongS := serverFields[7]
		lastPong := time.Now().Unix()
		if lastPongS != "" {
			lastPong, _ = strconv.ParseInt(lastPongS, 10, 64)
		}

		servers = append(servers, Server{
			Owner_id:   owner_id,
			Ip:         ip,
			Last_block: 0,
			Version:    "",
			Synced:     synced,
			Updated:    updated,
			LastPong:   lastPong,
		})
	}

	return servers, nil
}
