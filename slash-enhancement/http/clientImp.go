package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/qinglin89/gobsc/types"
	//	"github.com/valyala/fasthttp"
)

// FClient client imp
type FClient struct {
}

var client *http.Client

func init() {
	client = &http.Client{
		Timeout: 10 * time.Second,
	}
}

// PostJSON post
func (c *FClient) PostJSON(data, url string) types.HttpResponse {
	fmt.Println("DEBUG: url", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	fmt.Println("DEBUG req-err:", req, err)
	req.Header.Set("Content-Type", "application/json")
	resp := &http.Response{}
	//	var err error
	if resp, err = client.Do(req); err != nil {
		return types.HttpResponse{
			Err: err,
		}
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return types.HttpResponse{
		Status: resp.StatusCode,
		Data:   string(body),
		Err:    nil,
	}
}
