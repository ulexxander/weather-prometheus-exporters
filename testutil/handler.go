package testutil

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type HTTPHandler struct {
	Requests  chan *http.Request
	Responses chan interface{}
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{
		Requests:  make(chan *http.Request),
		Responses: make(chan interface{}),
	}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Requests <- r

	res := <-h.Responses

	var resJSON []byte
	switch v := res.(type) {
	case []byte:
		resJSON = v
	default:
		var err error
		resJSON, err = json.Marshal(v)
		if err != nil {
			panic(fmt.Errorf("marshaling JSON response: %w", err))
		}
	}

	if _, err := w.Write(resJSON); err != nil {
		panic(fmt.Errorf("writing response: %w", err))
	}
}
