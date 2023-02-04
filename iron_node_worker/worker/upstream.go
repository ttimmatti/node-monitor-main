package node_worker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ttimmatti/nodes-bot/ironfish/errror"
)

const IRON_BLOCK_API = "https://api.ironfish.network/blocks"
const IRON_VERSION_API = "https://api.github.com/repos/iron-fish/ironfish/releases/latest"

var GITHUB_TOKEN string

func GetLastNetworkBlock() (int64, error) {
	resp, err := http.Get(IRON_BLOCK_API)
	if err != nil {
		return 0, errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"get-last-network-block")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	response := make(map[string][]struct {
		Sequence int64
	})
	json.Unmarshal(body, &response)

	blocks := response["data"]
	if len(blocks) == 0 {
		return 0, errror.NewErrorf(errror.ErrorCodeFailure,
			"get-last-network-block_empty-list")
	}

	last_block := blocks[0]

	return last_block.Sequence, nil
}

func GetLastNetworkVersion() (string, error) {
	ctx, _ := context.WithTimeout(context.Background(),
		4000*time.Millisecond)
	r, err := http.NewRequestWithContext(ctx,
		http.MethodGet,
		IRON_VERSION_API, nil)
	r.Header["authorization"] = []string{fmt.Sprintf("Bearer %s", GITHUB_TOKEN)}
	if err != nil {
		return "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"GetLastNetworkVersion_newreq-withctx-construct:")
	}

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"get-last-network-version:")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	response := make(map[string]string)
	json.Unmarshal(body, &response)

	tag_name := response["tag_name"]
	if len(tag_name) < 2 {
		return "", errror.NewErrorf(errror.ErrorCodeFailure,
			"get-last-network-version_tag-name-empty: ", string(body))
	}

	return tag_name, nil
}
