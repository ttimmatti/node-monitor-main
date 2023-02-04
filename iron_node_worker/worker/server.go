package node_worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ttimmatti/nodes-bot/ironfish/db"
	"github.com/ttimmatti/nodes-bot/ironfish/errror"
)

const PORT = ":6596"

type Server struct {
	Owner_id   string
	Ip         string
	Last_block int64
	Version    string
	Synced     bool
	Updated    bool
	LastPong   int64
}

func (s *Server) Ping() error {
	uri := "http://" + s.Ip + PORT + "/ping"

	ctx, _ := context.WithTimeout(context.Background(),
		2000*time.Millisecond)
	r, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		uri, nil)
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

	result := make(map[string]string)
	if err := json.Unmarshal(body, &result); err != nil {
		return errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"server_ping_unmarshal:")
	}

	if result["result"] != "ok" {
		return errror.NewErrorf(errror.ErrorCodeFailure,
			"server_ping_result-not-ok")
	}

	return nil
}

// last_block int64, version string, err
func (s *Server) getInfo() (int64, string, error) {
	uri := "http://" + s.Ip + PORT + "/get?ret=all"

	ctx, _ := context.WithTimeout(context.Background(),
		30000*time.Millisecond)
	r, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
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

	s.Last_block = block
	s.Version = version

	//update server in db
	if err := db.UpdateBlVServer(s.Ip, version, block); err != nil {
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
