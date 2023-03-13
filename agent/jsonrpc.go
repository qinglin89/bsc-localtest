package agent

import (
	"errors"
	"fmt"
	"strconv"

	"encoding/json"

	"github.com/qinglin89/gobsc/http"
	"github.com/qinglin89/gobsc/types"
)

// Bsc jsonrpc methods
type Bsc struct {
	URL string
}

// DefaultBsc
var DefaultBsc = Bsc{
	URL: "http://127.0.0.1:8545",
}

var httpClient http.Client
var id int

func init() {
	httpClient = &http.FClient{}
}

func makeData(method, params string) string {
	id = (id + 1) % 10000
	return fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_%s","params":[%s],"id":`+fmt.Sprintf("%d}", id), method, params)
}

func makeDataMiner(method, params string) string {
	id = (id + 1) % 10000
	return fmt.Sprintf(`{"jsonrpc":"2.0","method":"miner_%s","params":[%s],"id":`+fmt.Sprintf("%d}", id), method, params)
}

func validRes(res types.HttpResponse) error {
	if res.Err != nil {
		return res.Err
	}
	if res.Status != 200 {
		return types.StatusCodeError
	}
	return nil
}
func validResTest(res types.HttpResponse, resDe *types.RpcResponse) error {
	if res.Err != nil {
		return res.Err
	}
	if res.Status != 200 {
		return types.StatusCodeError
	}
	json.Unmarshal([]byte(res.Data), &resDe)
	if resDe.Error != nil {
		if len(resDe.Error.Data) > 0 {
			return errors.New(resDe.Error.Data)
		}
		return errors.New(resDe.Error.Message)
	}

	return nil
}

func (b *Bsc) StartMiner() (bool, error) {
	res := httpClient.PostJSON(makeDataMiner("start", ""), b.URL)
	if err := validRes(res); err != nil {
		return false, err
	}
	return true, nil

}

func (b *Bsc) StopMiner() (bool, error) {
	res := httpClient.PostJSON(makeDataMiner("stop", ""), b.URL)
	if err := validRes(res); err != nil {
		return false, err
	}
	return true, nil
}

// BlockNumber bsc_blockNumber
func (b *Bsc) GetBlockNumber() (int64, error) {
	res := httpClient.PostJSON(makeData("blockNumber", ""), b.URL)
	if err := validRes(res); err != nil {
		return 0, err
	}
	resDe := types.ResBlockNumber{}
	json.Unmarshal([]byte(res.Data), &resDe)
	n, _ := strconv.ParseInt(resDe.Result[2:], 16, 64)
	return n, nil
}

// GetBlockByNumber bsc_getBlockByNumber
func (b *Bsc) GetBlockByNumber(n int64) (*types.Block, error) {
	res := httpClient.PostJSON(makeData("getBlockByNumber", fmt.Sprintf(`"0x%x", false`, n)), b.URL)
	if err := validRes(res); err != nil {
		return nil, err
	}
	resDe := types.ResBlock{}
	json.Unmarshal([]byte(res.Data), &resDe)
	if resDe.Error != nil {
		if len(resDe.Error.Data) > 0 {
			return nil, errors.New(resDe.Error.Data)
		}
		return nil, errors.New(resDe.Error.Message)
	}
	return resDe.Result, nil
}
func (b *Bsc) GetBlockByNumberDetails(n int64) (*types.BlockDetails, error) {
	res := httpClient.PostJSON(makeData("getBlockByNumber", fmt.Sprintf(`"0x%x", true`, n)), b.URL)
	if err := validRes(res); err != nil {
		return nil, err
	}
	resDe := types.ResBlockDetails{}
	json.Unmarshal([]byte(res.Data), &resDe)
	if resDe.Error != nil {
		if len(resDe.Error.Data) > 0 {
			return nil, errors.New(resDe.Error.Data)
		}
		return nil, errors.New(resDe.Error.Message)
	}
	if resDe.Result == nil {
		return nil, errors.New("trying future block")
	}
	return resDe.Result, nil
}

// GetBlockByHash bsc_getBlockByHash
func (b *Bsc) GetBlockByHash(h string) (*types.Block, error) {
	res := httpClient.PostJSON(makeData("getBlockByHash", fmt.Sprintf(`"%s", false`, h)), b.URL)
	if err := validRes(res); err != nil {
		return nil, err
	}
	resDe := types.ResBlock{}
	json.Unmarshal([]byte(res.Data), &resDe)
	return resDe.Result, nil
}
func (b *Bsc) GetBlockByHashDetails(h string) (*types.BlockDetails, error) {
	res := httpClient.PostJSON(makeData("getBlockByHash", fmt.Sprintf(`"%s", true`, h)), b.URL)
	if err := validRes(res); err != nil {
		return nil, err
	}
	resDe := types.ResBlockDetails{}
	json.Unmarshal([]byte(res.Data), &resDe)
	return resDe.Result, nil
}
