package http

import "github.com/qinglin89/gobsc/types"

// Client http client
type Client interface {
	PostJSON(data, url string) types.HttpResponse
}
