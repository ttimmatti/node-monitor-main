package node_worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ttimmatti/nodes-bot/sui/db"
	"github.com/ttimmatti/nodes-bot/sui/errror"
)

const PORT = ":9000"

// total txs on node :
// curl --location --request POST http://0.0.0.0:9000/ --header 'Content-Type: application/json' --data-raw '{ "jsonrpc":"2.0", "method":"sui_getTotalTransactionNumber","id":1}'

// version on node
// curl http://127.0.0.1:9184/metrics | grep uptime

type Server struct {
	Owner_id string
	Ip       string
	Tx_id0   int64
	Tx_id    int64
	Version  string
	Synced   bool
	Updated  bool
	Status   string
	LastPong int64
}

func (s *Server) Ping() error {
	uri := "http://" + s.Ip + PORT

	ctx, _ := context.WithTimeout(context.Background(),
		5000*time.Millisecond)
	r, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		uri, nil)
	r.Header["Content-Type"] = []string{"application/json"}
	if err != nil {
		return errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"server_ping_newreq-withctx-construct:")
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"server_ping_get:")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := string(body)

	if !strings.Contains(result, "HTTP") {
		return errror.NewErrorf(errror.ErrorCodeFailure,
			"server_ping_result-not-ok")
	}

	return nil
}

// last_block int64, version string, err
func (s *Server) getInfo() (int64, string, error) {
	chV := make(chan struct {
		Version string
		Err     error
	})

	chTx := make(chan struct {
		Tx_id int64
		Err   error
	})

	go s.GetVersion(chV)
	go s.GetTxId(chTx)

	resultV := <-chV
	resultTx := <-chTx

	if err := resultV.Err; err != nil {
		return 0, "", err
	}
	if err := resultTx.Err; err != nil {
		return 0, "", err
	}

	return resultTx.Tx_id, resultV.Version, nil
}

func (s *Server) GetVersion(c chan struct {
	Version string
	Err     error
}) {
	uri := "http://" + s.Ip + PORT

	bodyBytes, _ := json.Marshal(METHOD_DISCOVER)
	postBody := bytes.NewBuffer(bodyBytes)

	ctx, _ := context.WithTimeout(context.Background(),
		15000*time.Millisecond)
	r, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		uri, postBody)
	r.Header["Content-Type"] = []string{"application/json"}
	if err != nil {
		c <- struct {
			Version string
			Err     error
		}{"", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"server_getversion_newreq-withctx-construct:")}
		return
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		c <- struct {
			Version string
			Err     error
		}{"", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"server_getversion_get:")}
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	response := Rpc_discover{}
	json.Unmarshal(body, &response)

	version := response.Result.Info.Version
	if version == "" {
		c <- struct {
			Version string
			Err     error
		}{"", errror.NewErrorf(errror.ErrorCodeFailure,
			"get_server_version-empty:", s.Ip, string(body))}
		return
	}

	c <- struct {
		Version string
		Err     error
	}{
		version,
		nil,
	}
}

func (s *Server) GetTxId(c chan struct {
	Tx_id int64
	Err   error
}) {
	uri := "http://" + s.Ip + PORT

	bodyBytes, _ := json.Marshal(METHOD_TOTAL_TX)
	postBody := bytes.NewBuffer(bodyBytes)

	ctx, _ := context.WithTimeout(context.Background(),
		15000*time.Millisecond)
	r, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		uri, postBody)
	r.Header["Content-Type"] = []string{"application/json"}
	if err != nil {
		c <- struct {
			Tx_id int64
			Err   error
		}{0, errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"GetServerTxId_newreq-withctx-construct:"),
		}
		return
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		c <- struct {
			Tx_id int64
			Err   error
		}{0, errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"get-server-tx_id:"),
		}
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	response := make(map[string]int64)
	json.Unmarshal(body, &response)

	txid := response["result"]
	if txid == 0 {
		c <- struct {
			Tx_id int64
			Err   error
		}{0, errror.NewErrorf(errror.ErrorCodeFailure,
			"get-server-tx_id-empty: ", s.Ip, string(body)),
		}
		return
	}

	c <- struct {
		Tx_id int64
		Err   error
	}{
		txid,
		nil,
	}
}

func (s *Server) GetUpdate() (Server, error) {
	tx_id, version, err := s.getInfo()
	if err != nil {
		return *s, fmt.Errorf("server_GetUpdate: %w", err)
	}

	s.Tx_id = tx_id
	s.Version = version

	//update server in db
	if err := db.UpdateTxVServer(s.Ip, version, tx_id); err != nil {
		return *s, fmt.Errorf("server_update: %w", err)
	}

	return *s, nil
}

func (s *Server) UpdateInDb() error {
	if err := db.UpdateSyncUpdServer(s.Ip, s.Status, s.Synced, s.Updated); err != nil {
		return fmt.Errorf("server_update-in-db: %w", err)
	}

	return nil
}
