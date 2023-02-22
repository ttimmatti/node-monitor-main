package node_worker

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ttimmatti/nodes-bot/sui/errror"
)

const SUI_RPC = "https://fullnode.devnet.sui.io:443"

// global version :
// curl --location --request POST http://fullnode.testnet.sui.io:9000/ --header 'Content-Type: application/json' --data-raw '{ "jsonrpc":"2.0", "method":"rpc.discover","id":1}'
var METHOD_DISCOVER = map[string]string{"jsonrpc": "2.0", "method": "rpc.discover", "id": "1"}
var METHOD_TOTAL_TX = map[string]string{"jsonrpc": "2.0", "method": "sui_getTotalTransactionNumber", "id": "1"}

// global txs :
// curl --location --request POST http://fullnode.testnet.sui.io:9000/ --header 'Content-Type: application/json' --data-raw '{ "jsonrpc":"2.0", "method":"sui_getTotalTransactionNumber","id":1}'

func GetNetworkVersion() (string, error) {
	bodyBytes, _ := json.Marshal(METHOD_DISCOVER)
	postBody := bytes.NewBuffer(bodyBytes)

	ctx, _ := context.WithTimeout(context.Background(),
		10000*time.Millisecond)
	r, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		SUI_RPC, postBody)
	r.Header["Content-Type"] = []string{"application/json"}
	if err != nil {
		return "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"GetNetworkVersion_newreq-withctx-construct:")
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"get-network-version:")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	response := Rpc_discover{}
	json.Unmarshal(body, &response)

	version := response.Result.Info.Version
	if version == "" {
		return "", errror.NewErrorf(errror.ErrorCodeFailure,
			"get_network_version-empty:", body)
	}

	return version, nil
}

func GetNetworkTxId() (int64, error) {
	bodyBytes, _ := json.Marshal(METHOD_TOTAL_TX)
	postBody := bytes.NewBuffer(bodyBytes)

	ctx, _ := context.WithTimeout(context.Background(),
		10000*time.Millisecond)
	r, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		SUI_RPC, postBody)
	r.Header["Content-Type"] = []string{"application/json"}
	if err != nil {
		return 0, errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"GetNetworkTxId_newreq-withctx-construct:")
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return 0, errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"get-network-tx_id:")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	response := make(map[string]int64)
	json.Unmarshal(body, &response)

	txid := response["result"]
	if txid == 0 {
		return 0, errror.NewErrorf(errror.ErrorCodeFailure,
			"get-network-tx_id-empty: ", string(body))
	}

	return txid, nil
}

type Rpc_discover struct {
	Result struct {
		Info struct {
			Version string
		}
	}
}
