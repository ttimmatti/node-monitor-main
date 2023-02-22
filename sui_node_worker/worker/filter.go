package node_worker

import (
	"log"
	"sync"
)

var FILTERED_SERVERS []Server
var NOPONG_SERVERS []Server
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
		s.Status = "Node is down"
		s.Synced = false
		NOPONG_SERVERS = append(NOPONG_SERVERS, s)
		log.Printf("go: tryPing: ping to %s failed", s.Ip)
	}
}
