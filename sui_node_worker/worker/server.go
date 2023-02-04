package node_worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	Tx_id    int64
	Version  string
	Synced   bool
	Updated  bool
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
	uri := "http://" + s.Ip + PORT

	ctx, _ := context.WithTimeout(context.Background(),
		5000*time.Millisecond)
	r, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		uri, nil)
	if err != nil {
		return 0, "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"server_getinfo_newreq-withctx-construct:")
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return 0, "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"server_getinfo_get:")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	result := make(map[string]string)
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"server_getinfo_unmarshal:", string(body))
	}

	bl := result["lastBlock"]
	v := result["version"]
	block, err := strconv.ParseInt(bl, 10, 64)
	if len(bl) < 1 || len(v) < 1 || err != nil {
		return 0, "", errror.NewErrorf(errror.ErrorCodeFailure,
			"server_getinfo_params_bad:", string(body))
	}

	version := "v" + v

	return block, version, nil
}

func (s *Server) GetUpdate() (Server, error) {
	block, version, err := s.getInfo()
	if err != nil {
		return *s, fmt.Errorf("server_Update: %w", err)
	}

	s.Tx_id = block
	s.Version = version

	//update server in db
	if err := db.UpdateTxVServer(s.Ip, version, block); err != nil {
		return *s, fmt.Errorf("server_update: %w", err)
	}

	return *s, nil
}

func (s *Server) UpdateInDb() error {
	if err := db.UpdateSyncUpdServer(s.Ip, s.Synced, s.Updated); err != nil {
		return fmt.Errorf("server_update-in-db: %w", err)
	}

	return nil
}
