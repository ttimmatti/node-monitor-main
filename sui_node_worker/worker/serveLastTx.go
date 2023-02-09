package node_worker

import (
	"fmt"
	"log"
	"net/http"
)

func StartListening() {
	http.HandleFunc("/tx_id", GetTxId)

	log.Println("started listening \"localhost:6591\"")
	if err := http.ListenAndServe("localhost:6591", nil); err != nil {
		log.Fatalln(err)
	}
}

func GetTxId(w http.ResponseWriter, r *http.Request) {
	last_tx := LAST_TX

	w.Write([]byte(fmt.Sprintf("%d", last_tx)))
}
