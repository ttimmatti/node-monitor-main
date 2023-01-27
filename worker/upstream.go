package node_worker

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ttimmatti/ironfish-node-tg/errror"
)

const IRON_BLOCK_API = "https://api.ironfish.network/blocks"
const IRON_VERSION_API = "https://api.github.com/repos/iron-fish/ironfish/releases/latest"

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
	resp, err := http.Get(IRON_VERSION_API)
	if err != nil {
		return "", errror.WrapErrorF(err,
			errror.ErrorCodeFailure,
			"get-last-network-version")
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	response := make(map[string]string)
	json.Unmarshal(body, &response)

	tag_name := response["tag_name"]
	if len(tag_name) < 2 {
		return "", errror.NewErrorf(errror.ErrorCodeFailure,
			"get-last-network-version_tag-name-empty")
	}

	return tag_name, nil
}
