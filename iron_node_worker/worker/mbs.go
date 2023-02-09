package node_worker

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/ttimmatti/nodes-bot/ironfish/db"
	"github.com/ttimmatti/nodes-bot/ironfish/errror"
)

var MBS_REPEAT = 5

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
		if errs := handleMBS(i); errs != nil {
			log.Println("iron_node_worker: StartMBS: ", errs)
		}
		// every 5 mi
		if i%(36*4) == 0 {
			// every 12 hours
			filterLost()
		}
		time.Sleep(REPEAT_MIN * time.Minute)

		t := time.Now().Unix() & 604800
		if MONDAY0 <= t && t < THURSDAY0 {
			// if server doesnt have time assigned
			// assign time from t to THURSDAY0
			for i1 := range SERVERS_MBS {
				if !SERVERS_MBS[i1].Mbs_completed && SERVERS_MBS[i1].Time_mbs == 0 {
					t_mbs := rand.Int63()%(WEEK-MONDAY0) + MONDAY0
					SERVERS_MBS[i].Time_mbs = rand
				}
			}
		} else if THURSDAY0 <= t && t < SATURDAY0 {
			// if server doesnt have time assigned
			// assign time from t to SATURDAY0
		}
	}
}

func handleMBS(i int) []error {
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
