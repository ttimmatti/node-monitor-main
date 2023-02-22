package sui

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ttimmatti/nodes-bot/tg-db/errror"
)

var METHOD_TOTAL_TX = map[string]string{"jsonrpc": "2.0", "method": "sui_getTotalTransactionNumber", "id": "1"}

const SUI_PORT = ":9000"

func GetTxId(ip string, c chan struct {
	Tx_id int64
	Err   error
}) {
	uri := "http://" + ip + SUI_PORT

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
			errror.ErrorCodePongFalse,
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
			"get-server-tx_id-empty: ", ip, string(body)),
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
