package node_worker

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/ttimmatti/nodes-bot/ironfish/errror"
)

type Server_mbs struct {
	Server
	Time_mbs      int64
	Mbs_completed bool
	Mbs_entered   bool
}

func (s_mbs *Server_mbs) ReqMint() (Mbs_server_response, error) {
	mbs_response := Mbs_server_response{}

	uri := "http://" + s_mbs.Ip + PORT + "/mint"

	resp, err := http.Get(uri)
	if err != nil {
		return mbs_response, errror.WrapErrorF(
			err,
			errror.ErrorCodePongFalse,
			"ReqMint: No response",
		)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	body := string(b)

	if strings.Contains(body, "Insufficient funds") {
		return mbs_response, errror.NewErrorf(errror.ErrorCodeNoFunds,
			"ReqMint: not enough funds", body)
	}

	if err := json.Unmarshal(b, &mbs_response); err != nil {
		return mbs_response, errror.WrapErrorF(err, errror.ErrorCodeUnknown,
			"ReqMint: unmarshal error:", body)
	}

	if mbs_response.Result == "ok" {
		return mbs_response, nil
	}

	return mbs_response, errror.NewErrorf(errror.ErrorCodeUnknown,
		"ReqMint: result not ok:", body)
}

func (s_mbs *Server_mbs) ReqBurn() (Mbs_server_response, error) {
	mbs_response := Mbs_server_response{}

	uri := "http://" + s_mbs.Ip + PORT + "/burn"

	resp, err := http.Get(uri)
	if err != nil {
		return mbs_response, errror.WrapErrorF(
			err,
			errror.ErrorCodePongFalse,
			"ReqMint: No response",
		)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	body := string(b)

	if strings.Contains(body, "Insufficient funds") {
		return mbs_response, errror.NewErrorf(errror.ErrorCodeNoFunds,
			"ReqMint: not enough funds", body)
	}

	if err := json.Unmarshal(b, &mbs_response); err != nil {
		return mbs_response, errror.WrapErrorF(err, errror.ErrorCodeUnknown,
			"ReqMint: unmarshal error:", body)
	}

	if mbs_response.Result == "ok" {
		return mbs_response, nil
	}

	return mbs_response, errror.NewErrorf(errror.ErrorCodeUnknown,
		"ReqMint: result not ok:", body)
}

func (s_mbs *Server_mbs) ReqSend() (Mbs_server_response, error) {
	mbs_response := Mbs_server_response{}

	uri := "http://" + s_mbs.Ip + PORT + "/send"

	resp, err := http.Get(uri)
	if err != nil {
		return mbs_response, errror.WrapErrorF(
			err,
			errror.ErrorCodePongFalse,
			"ReqMint: No response",
		)
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	body := string(b)

	if strings.Contains(body, "Insufficient funds") {
		return mbs_response, errror.NewErrorf(errror.ErrorCodeNoFunds,
			"ReqMint: not enough funds", body)
	}

	if err := json.Unmarshal(b, &mbs_response); err != nil {
		return mbs_response, errror.WrapErrorF(err, errror.ErrorCodeUnknown,
			"ReqMint: unmarshal error:", body)
	}

	if mbs_response.Result == "ok" {
		return mbs_response, nil
	}

	return mbs_response, errror.NewErrorf(errror.ErrorCodeUnknown,
		"ReqMint: result not ok:", body)
}

func (s_mbs *Server_mbs) ReqFaucet() (bool, error) {
	mbs_response := Mbs_server_response{}

	uri := "http://" + s_mbs.Ip + PORT + "/faucet"

	resp, err := http.Get(uri)
	if err != nil {
		return false, nil
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)

	body := string(b)

	if mbs_response.Result == "ok" {
		return true, nil
	}

	return false, errror.NewErrorf(errror.ErrorCodeUnknown,
		"ReqFaucet: result not ok:", body)
}

type Mbs_server_response struct {
	Result   string
	Asset_id string
	Hash     string
}

func (s *Server_mbs) IndexIn(ss []Server_mbs) int {
	for i, s := range ss {
		if s.Ip == s.Ip {
			return i
		}
	}
	return -1
}
