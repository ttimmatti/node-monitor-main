package node_worker

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ttimmatti/nodes-bot/sui/db"
	"github.com/ttimmatti/nodes-bot/sui/errror"
)

const REPEAT_MIN = 5
const TX_ALERT_GAP = 5000

var SERVERS []Server
var WG_DialServers sync.WaitGroup

var LAST_TX int64
var LAST_VERSION string

func Start() {
	//checks the list of servers that pong me back
	// save the list

	//get sui status response from every server
	//get tx_id from explorer
	// compare latest blocks and versions

	//also check version through their github releases

	for i := 0; ; i++ {
		if errs := handleServers(i); errs != nil {
			log.Println("iron_node_worker: START: ", errs)
		}
		// once every few hours inspect all servers and notify those whose lastpong was more than 3 hours ago
		// tell them if they don't use the bot they should consider deleting their server
		// or there is something with the server
		// if lastpong = 0 do nothing
		if i%(36*4) == 0 {
			// every 12 hours
			filterLost()
		}
		time.Sleep(REPEAT_MIN * time.Minute)
	}
}

func handleServers(i int) []error {
	errs := []error{}
	//it should be wrapped in cycle.
	// get initial list of servers
	result, err := db.SuiReadServers()
	if err != nil {
		return []error{fmt.Errorf("node_worker_handleServers: %w", err)}
	}

	if len(result) < 2 {
		return nil
	}

	servers, err := GetServers(result)
	if err != nil {
		return []error{fmt.Errorf("node_worker_handleServers: %w", err)}
	}

	last_v, err := GetNetworkVersion()
	if err == nil { // if error IS nil
		LAST_VERSION = last_v
	} else {
		errs = append(errs, err)
	}
	last_tx_id, err := GetNetworkTxId()
	if err == nil { // if error IS nil
		LAST_TX = last_tx_id
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
	if LAST_TX < 2 || len(LAST_VERSION) < 2 {
		errs = append(errs,
			errror.NewErrorf(
				errror.ErrorCodeFailure,
				fmt.Sprintf("handle_servers: before send msgs: last bl/v are empty. cycle N%d", i)))
		return errs
	}

	// i need to define servers that are syncing and servers that were synced but lost sync
	// servers that are syncing

	if ers := DefineNotifyBadServers(SERVERS, LAST_TX, LAST_VERSION); errs != nil {
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

func DefineNotifyBadServers(servers []Server, last_block int64, last_version string) []error {
	errs := []error{}
	for i, s := range servers {
		if s.Tx_id < last_block-TX_ALERT_GAP {
			if s.Tx_id0 < s.Tx_id {
				// means it's syncing
				SERVERS[i].Status = "Syncing"
			} else {
				// if it's not synced and it STOPPED syncing (means it was syncing before now)
				err := notifySNotSyncing(s)
				if err != nil {
					errs = append(errs, err)
				}
				SERVERS[i].Status = "Not syncing"
			}

			if err := notifySBadSync(s, last_block); err != nil {
				errs = append(errs, err)
			}
			SERVERS[i].Synced = false
		} else {
			for i := range SERVERS {
				if s.Ip == SERVERS[i].Ip {
					SERVERS[i].Synced = true
					SERVERS[i].Status = "Synced"
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

func notifySNotSyncing(s Server) error {
	//check if server was not synced before that lap
	// if wasn't synced than don't notify, because we already did notify
	if s.Status != "Syncing" {
		return nil
	}

	//sendMsg to owner that server is not synced
	chat_id, _ := strconv.ParseInt(s.Owner_id, 10, 64)
	msg := &SendMsg{
		Chat_id: chat_id,
		Text: fmt.Sprintf(
			"#sui\n\nHey! Your server: %s stopped syncing.\n\nYour server last tx: %d", s.Ip, s.Tx_id,
		) + "\n\n-----------------------------------------------------------" + fmt.Sprintf(
			"\n\nЙоу! Синхронизация на твоем сервере: %s остановилась.\n\nПоследняя транзакция на этом сервере: %d", s.Ip, s.Tx_id,
		),
		Parse_mode: "",
	}

	err := sendMsg(msg)
	if err != nil {
		return fmt.Errorf("noty_bad_sync: %w", err)
	}
	return nil
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
			"#sui\n\nHey! Your server: %s has old Sui version, there is an update.\n\nYour server version: %s, Last version: %s\n\nYou can update your node via our <a href=\"https://fackblock.com/fiuKCgPWS25#EMof\">Guide</a>", s.Ip, s.Version, last_version,
		) + "\n\n-----------------------------------------------------------" + fmt.Sprintf(
			"\n\nЙоу! Версия Sui на твоем сервере: %s устарела.\n\nВерсия на этом сервере: %s, Последняя версия: %s\n\nМожешь обновиться командами из раздела гайда <a href=\"https://fackblock.com/fiuKCgPWS25#EMof\">Обновление</a>", s.Ip, s.Version, last_version,
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
			"#sui\n\nHey! Your server: %s has lost sync.\n\nYour server last tx: %d, Network last tx: %d", s.Ip, s.Tx_id, last_block,
		) + "\n\n-----------------------------------------------------------" + fmt.Sprintf(
			"\n\nЙоу! Твой сервер: %s отстал от сети.\n\nПоследняя транзакция на этом сервере: %d, Последняя транза в сети: %d", s.Ip, s.Tx_id, last_block,
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

		if len(serverFields) < 9 {
			return nil, errror.NewErrorf(
				errror.ErrorCodeFailure,
				"getservers_not-enough-args", serverFields,
			)
		}

		owner_id := serverFields[1]
		ip := serverFields[2]
		tx_idS := serverFields[3]
		var tx_id int64
		if tx_idS != "0" && tx_idS != "" {
			tx_id, _ = strconv.ParseInt(tx_idS, 10, 64)
		}
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
		status := serverFields[8]
		if status == "" {
			status = "Pending"
		}

		servers = append(servers, Server{
			Owner_id: owner_id,
			Ip:       ip,
			Tx_id0:   tx_id,
			Tx_id:    tx_id,
			Version:  "",
			Synced:   synced,
			Updated:  updated,
			Status:   status,
			LastPong: lastPong,
		})
	}

	return servers, nil
}
