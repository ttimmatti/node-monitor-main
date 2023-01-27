package node_worker

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ttimmatti/ironfish-node-tg/db"
	"github.com/ttimmatti/ironfish-node-tg/errror"
)

func Start() error {
	//checks the list of servers that pong me back
	// save the list

	//get ironfish status response from every server
	//get block from explorer
	// compare latest blocks and versions

	//also check version through their github releases

	//it should be wrapped in cycle.
	// get initial list of servers
	result, err := db.ReadServers()
	if err != nil {
		return fmt.Errorf("node_worker_Start: %w", err)
	}

	log.Println(result)

	servers, err := GetServers(result)
	if err != nil {
		return fmt.Errorf("node_worker_Start: %w", err)
	}

	log.Println(servers)

	last_block, err := GetLastNetworkBlock()
	if err != nil {
		return fmt.Errorf("node_worker_Start: %w", err)
	}
	last_version, err := GetLastNetworkVersion()
	if err != nil {
		return fmt.Errorf("node_worker_Start: %w", err)
	}

	//check Ping
	servers = FilterPing(servers)

	//this should be done asynchorously
	//for all pinged servers check versions and last blocks, write response to interface
	for i, s := range servers {
		block, version, err := s.GetInfo()
		if err != nil {
			return fmt.Errorf("node_worker_Start: %w", err)
		}

		servers[i].Last_block = block
		servers[i].Version = version
	}

	//compare
	for _, s := range servers {
		if s.Last_block < last_block-10 {
			//sendMsg to owner that server is not synced
		}
		if s.Version != last_version {
			//sendMsg to owner that server is version behind and needs to update with cmd
		}
		//server is working fine
	}

	for {
		log.Println(servers, last_block, last_version)
		time.Sleep(1 * time.Minute)
	}
	return nil
}

func FilterPing(servers []Server) []Server {
	var filteredServers []Server

	for _, server := range servers {
		err := server.Ping()
		if err == nil {
			// append to new arr
			filteredServers = append(filteredServers, server)
		}
	}

	return filteredServers
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
