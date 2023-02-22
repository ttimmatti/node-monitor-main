package node_worker

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/ttimmatti/nodes-bot/ironfish/errror"
)

var ADMIN_ID int64

const MBS_REPEAT = 2
const MINUTES_BETWEEN_TXS = 2

var SERVERS_MBS []Server_mbs

const MONDAY0 = 345600
const THURSDAY0 = 1
const SATURDAY0 = 259201
const WEEK = 604800

const (
	PHASE0 = iota // at start
	PHASE1        // monday - thursday
	PHASE2        // thursday - saturday
)

func StartMBS() {
	for i := 0; ; i++ {
		// do work if needed. checks evry 2 minutes

		if errs := handleMBS(i); errs != nil {
			log.Println("iron_node_worker: StartMBS: ", errs)
		}

		time.Sleep(MBS_REPEAT * time.Minute)
	}
}

func handleMBS(i int) []error {
	log.Println("MBS")
	// check if theres work to be done

	errs := []error{}
	//it should be wrapped in cycle.
	// get initial list of servers

	// result, err := db.IronReadServers()
	// if err != nil {
	// 	return []error{fmt.Errorf("node_worker_handleServers: %w", err)}
	// }

	// servers, err := GetServers(result)
	// if err != nil {
	// 	return []error{fmt.Errorf("node_worker_handleServers: %w", err)}
	// }

	// srvrs_appended := GetMbsServers(servers)
	// if srvrs_appended != 0 {
	// 	log.Printf("MBS: appended %d servers", srvrs_appended)
	// }
	// AssignTime(i)

	if i == 0 {
		SERVERS_MBS = AppendMyTestServer()
	}

	//check Ping
	//servers = FilterPing(servers)

	//clear servers before append
	SERVERS = []Server{}

	// execs mbs for SERVERS_MBS where time has come to exec
	log.Println("SERVERS:", SERVERS_MBS)
	var c_res_mbs []chan Mbs_result
	var c_ss_mbs []chan Server_mbs
	go ExecMbs(c_res_mbs, c_ss_mbs)

	// is result assign 0 and true to mbs stats to server
	HandleMbsResults(c_res_mbs, c_ss_mbs)

	if ers := DefineNotifyBadServers(SERVERS, LAST_BLOCK, LAST_VERSION); ers != nil {
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

func HandleMbsResults(c_mbs []chan Mbs_result, c_ss_mbs []chan Server_mbs) {
	for i, c := range c_mbs {
		go handleResult(i, c, c_ss_mbs[i])
	}
}

func handleResult(i int, c chan Mbs_result, c_s_mbs chan Server_mbs) {
	log.Println("handleResult")
	mbs_result := <-c
	log.Printf("%d received mbs_result", i)
	s := <-c_s_mbs
	log.Printf("%d received s %s", i, s.Ip)
	index := s.IndexIn(SERVERS_MBS)

	id, _ := strconv.ParseInt(s.Owner_id, 10, 64)
	if mbs_result.Error != nil {
		myerr := &errror.Error{}
		if errors.As(mbs_result.Error, &myerr) {
			if myerr.Code() == errror.ErrorCodeNoFunds {
				_, err := s.ReqFaucet()
				if err != nil {
					msg := &SendMsg{
						Text: fmt.Sprintf("#ironfish\n\nServer: %s\nMBS error: %s\nStatus: %s",
							strings.Join(strings.Split(s.Ip, "."), "\\."),
							"Not enough funds; Tried using the faucet but faucet returned error",
							mbs_result.Mbs_result),
						Chat_id:    id,
						Parse_mode: "",
					}
					err := sendMsg(msg)
					if err != nil {
						msg2 := &SendMsg{
							Text:       err.Error(),
							Chat_id:    ADMIN_ID,
							Parse_mode: "",
						}
						if err := sendMsg(msg2); err != nil {
							log.Printf("handleResult: ERROR SENDING MSG TO ADMIN: %s. Text: %s", err, msg.Text)
						}
					}
				}
				SERVERS_MBS[index].Time_mbs = 0
			}
			return
		}

		errS := fmt.Sprintf("handleResult: Server: %s; Error: %s; Status: %s;", s.Ip, mbs_result.Error, mbs_result.Mbs_result)
		log.Println(errS)
		msg2 := &SendMsg{
			Text:       errS,
			Chat_id:    ADMIN_ID,
			Parse_mode: "",
		}
		if err := sendMsg(msg2); err != nil {
			log.Printf("handleResult: ERROR SENDING MSG TO ADMIN: %s. Text: %s", err, msg2.Text)
		}

		// if it's not funds set only mbs true. time is as was, so that server does not re enter AssignTime
		// if funds it'll re enter assign time and Exec again
		SERVERS_MBS[index].Mbs_completed = false
		return
	}

	msg := &SendMsg{
		Text: fmt.Sprintf("#ironfish\n\nServer: %s\nMBS completed\nStatus: %s",
			strings.Join(strings.Split(s.Ip, "."), "\\."),
			mbs_result.Mbs_result),
		Chat_id:    id,
		Parse_mode: "",
	}
	if err := sendMsg(msg); err != nil {
		log.Printf("handleResult: ERROR SENDING MSG: %s. Text: %s", err, msg.Text)
	}

	SERVERS_MBS[index].Mbs_completed = true
	SERVERS_MBS[index].Time_mbs = 0
}

func ExecMbs(c_mbs []chan Mbs_result, c_ss_mbs []chan Server_mbs) {
	// get time now
	log.Println("StartMBS: ExecMbs")
	t := time.Now().Unix()
	// if s.Time_mbs has passed and mbs not completed exec mbs
	i := 0
	for _, s_mbs := range SERVERS_MBS {
		if s_mbs.Time_mbs > t || s_mbs.Mbs_entered || s_mbs.Mbs_completed || s_mbs.Time_mbs == 0 {
			continue
		}

		c_mbs = append(c_mbs, make(chan Mbs_result))
		c_ss_mbs = append(c_ss_mbs, make(chan Server_mbs))
		log.Printf("ExecMbs: deploying exec %d for %s", i, s_mbs.Ip)

		index := s_mbs.IndexIn(SERVERS_MBS)
		SERVERS_MBS[index].Mbs_entered = true

		go ExecMbsForServer(s_mbs, c_mbs[i], c_ss_mbs[i])
		i++
	}
}

func ExecMbsForServer(s_mbs Server_mbs, c chan Mbs_result, c_s_mbs chan Server_mbs) {
	defer func() {
		c_s_mbs <- s_mbs
	}()
	id, _ := strconv.ParseInt(s_mbs.Owner_id, 10, 64)
	// mint
	log.Println("ExecMbsForServer " + s_mbs.Ip + " started\nMinting")
	respM, err := s_mbs.ReqMint()
	if err != nil {
		log.Println(err)
		c <- Mbs_result{
			Mbs_result:    "Error while minting",
			Mbs_completed: false,
			Error:         err,
		}
		return
	} else {
		log.Println("success")
		// return success to user
		msg := &SendMsg{
			Text: fmt.Sprintf("#ironfish\n\nServer: %s\n\nMint tx completed\n\nTx hash: %s\nAsset_id: %s",
				s_mbs.Ip, respM.Hash, respM.Asset_id),
			Chat_id:    id,
			Parse_mode: "",
		}
		if err := sendMsg(msg); err != nil {
			msg2 := &SendMsg{
				Text:       err.Error(),
				Chat_id:    ADMIN_ID,
				Parse_mode: "",
			}
			if err := sendMsg(msg2); err != nil {
				log.Printf("handleResult: ERROR SENDING MSG TO ADMIN: %s. Text: %s", err, msg.Text)
			}
		}
	}
	time.Sleep(MINUTES_BETWEEN_TXS * time.Minute)

	log.Println("Burning")

	// burn
	respB, err := s_mbs.ReqBurn()
	if err != nil {
		log.Println(err)
		c <- Mbs_result{
			Mbs_result:    "Error while burning; Mint completed",
			Mbs_completed: false,
			Error:         err,
		}
		return
	} else {
		// return success to user
		log.Println("success")
		msg := &SendMsg{
			Text: fmt.Sprintf("#ironfish\n\nServer: %s\n\nBurn tx completed\n\nTx hash: %s",
				s_mbs.Ip, respM.Hash),
			Chat_id:    id,
			Parse_mode: "",
		}
		if err := sendMsg(msg); err != nil {
			msg2 := &SendMsg{
				Text:       err.Error(),
				Chat_id:    ADMIN_ID,
				Parse_mode: "",
			}
			if err := sendMsg(msg2); err != nil {
				log.Printf("handleResult: ERROR SENDING MSG TO ADMIN: %s. Text: %s", err, msg.Text)
			}
		}
	}
	time.Sleep(MINUTES_BETWEEN_TXS * time.Minute)

	// send
	respS, err := s_mbs.ReqSend()
	if err != nil {
		c <- Mbs_result{
			Mbs_result:    "Error while sending; Mint and Burn completed",
			Mbs_completed: false,
			Error:         err,
		}
		return
	} else {
		// return success to user
		msg := &SendMsg{
			Text: fmt.Sprintf("#ironfish\n\nServer: %s\n\nSend tx completed\n\nTx hash: %s",
				s_mbs.Ip, respM.Hash),
			Chat_id:    id,
			Parse_mode: "",
		}
		if err := sendMsg(msg); err != nil {
			msg2 := &SendMsg{
				Text:       err.Error(),
				Chat_id:    ADMIN_ID,
				Parse_mode: "",
			}
			if err := sendMsg(msg2); err != nil {
				log.Printf("handleResult: ERROR SENDING MSG TO ADMIN: %s. Text: %s", err, msg.Text)
			}
		}
	}
	time.Sleep(MINUTES_BETWEEN_TXS * time.Minute)

	c <- Mbs_result{
		Mbs_result: fmt.Sprintf("%s;;%s;;%s",
			respM.Hash, respB.Hash, respS.Hash),
		Mbs_completed: true,
		Error:         nil,
	}
}

type Mbs_result struct {
	Mbs_result    string
	Mbs_completed bool
	Error         error
}

// returns number of servers appended to SERVERS_MBS
func GetMbsServers(servers []Server) int {
	counter := 0
	for _, s := range servers {
		server_is_in_mbs := false
		for _, s_mbs := range SERVERS_MBS {
			if s.Ip == s_mbs.Ip {
				server_is_in_mbs = true
				break
			}
		}

		if !server_is_in_mbs {
			if !s.Synced {
				continue
			}

			SERVERS_MBS = append(SERVERS_MBS,
				Server_mbs{
					Server:        s,
					Time_mbs:      0,
					Mbs_completed: false,
				})
			counter++
		}
	}
	return counter
}

func AssignTime(i int) {
	if i%(30*2) != 0 {
		// do next lines only each 2 hours

		// time of the week relative to the unix's 0
		t := time.Now().Unix() % 604800
		if MONDAY0 <= t && t < THURSDAY0 {
			// if server doesnt have time assigned
			// assign time from t to THURSDAY0
			for i1 := range SERVERS_MBS {
				if !SERVERS_MBS[i1].Mbs_completed && SERVERS_MBS[i1].Time_mbs == 0 {
					//time of the week unix0 to do the work
					t_mbs := rand.Int63()%(WEEK-MONDAY0-t) + t
					// weeks since unix0. for some reason i have to subtract one. idk why. otherwise it's next week
					weeks_unix0 := math.Ceil(float64(time.Now().Unix())/604800) - 1
					// week_time + seconds to this week
					SERVERS_MBS[i1].Time_mbs = t_mbs + int64(weeks_unix0)*604800
				}
			}
		} else if THURSDAY0 <= t && t < SATURDAY0 {
			// if server doesnt have time assigned
			// assign time from t to SATURDAY0
			for i1 := range SERVERS_MBS {
				if !SERVERS_MBS[i1].Mbs_completed && SERVERS_MBS[i1].Time_mbs == 0 {
					t_mbs := rand.Int63()%(SATURDAY0-THURSDAY0-t) + t
					weeks_unix0 := math.Ceil(float64(time.Now().Unix())/604800) - 1
					SERVERS_MBS[i1].Time_mbs = t_mbs + int64(weeks_unix0)*604800
				}
			}
		}
	}
}

func AppendMyTestServer() []Server_mbs {
	sever_mbs := []Server_mbs{}

	sever_mbs = append(sever_mbs, Server_mbs{
		Server: Server{
			Owner_id: fmt.Sprintf("%d", ADMIN_ID),
			Ip:       "135.181.149.202",
			Synced:   true,
		},
		Mbs_completed: false,
		Time_mbs:      1676119774,
	})

	sever_mbs = append(sever_mbs, Server_mbs{
		Server: Server{
			Owner_id: fmt.Sprintf("%d", ADMIN_ID),
			Ip:       "173.212.216.234",
			Synced:   true,
		},
		Mbs_completed: false,
		Time_mbs:      1676119774,
	})

	sever_mbs = append(sever_mbs, Server_mbs{
		Server: Server{
			Owner_id: fmt.Sprintf("%d", ADMIN_ID),
			Ip:       "159.69.68.42",
			Synced:   true,
		},
		Mbs_completed: false,
		Time_mbs:      1676119774,
	})

	return sever_mbs
}
