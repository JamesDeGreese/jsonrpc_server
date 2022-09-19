package jsonrpc_server

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

type r struct {
}

type w struct {
	Foo     string `json:"foo" validation:"required"`
	Timeout int    `json:"timeout,omitempty" validation:"numeric"`
}

func (w *w) Handle(ctx context.Context) (interface{}, *Error) {
	if w.Timeout != 0 {
		time.Sleep(time.Duration(w.Timeout) * time.Second)
	}
	return w.Foo, nil
}

func (r *r) ResolveWorker(method string) (Worker, *Error) {
	if method != "test" {
		return nil, &Error{Code: -32601, Message: "The method does not exist"}
	}
	return &w{}, nil
}

type Client struct {
	Client *http.Client
	URL    string
}

func (c *Client) MakeRequest(request Request) (Response, error) {
	var res Response
	req, _ := json.Marshal(request)
	resp, err := c.Client.Post(c.URL, "application/json", bytes.NewReader(req))
	if err != nil {
		return res, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &res)

	return res, err
}

func TestServer(t *testing.T) {
	type test struct {
		request  Request
		response Response
	}

	r := r{}

	s := Server{
		Address: "localhost:8888",
		Router:  &r,
		Ctx:     context.Background(),
		Timeout: 3,
	}

	tests := []test{
		{
			request: Request{
				JsonRPC: "2.0",
				Method:  "test",
				Params: struct {
					Foo string
				}{Foo: "bar"},
				ID: "d11aa18a-c689-4576-b3d8-a55b63689389",
			},
			response: Response{
				JsonRPC: "2.0",
				Result:  "bar",
				Error:   nil,
				ID:      "d11aa18a-c689-4576-b3d8-a55b63689389",
			},
		},
		{
			request: Request{
				JsonRPC: "2.0",
				Method:  "test",
				Params: struct {
					Foo int
				}{Foo: 100},
				ID: "0cd909b9-d3ae-4740-88fe-727f739a3bf8",
			},
			response: Response{
				JsonRPC: "2.0",
				Result:  nil,
				Error: &Error{
					Code:    -32602,
					Message: "Invalid method parameter(s).",
				},
				ID: "0cd909b9-d3ae-4740-88fe-727f739a3bf8",
			},
		},
		{
			request: Request{
				JsonRPC: "2.0",
				Method:  "foo",
				Params:  nil,
				ID:      "808576f5-3d30-4b3a-bc48-556df9cf9ada",
			},
			response: Response{
				JsonRPC: "2.0",
				Result:  nil,
				Error: &Error{
					Code:    -32601,
					Message: "The method does not exist",
				},
				ID: "808576f5-3d30-4b3a-bc48-556df9cf9ada",
			},
		},
		{
			request: Request{
				JsonRPC: "2.0",
				Method:  "test",
				Params: struct {
					Foo     string
					Timeout int
				}{Foo: "bar", Timeout: 20},
				ID: "2c7f50cf-963d-42a9-9a92-30657e77c0c6",
			},
			response: Response{
				JsonRPC: "2.0",
				Result:  nil,
				Error: &Error{
					Code:    -32600,
					Message: "Request timeout",
				},
				ID: nil,
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(s.ProcessRequest))
	defer srv.Close()
	client := Client{
		Client: srv.Client(),
		URL:    srv.URL,
	}

	for _, tc := range tests {
		res, err := client.MakeRequest(tc.request)
		if err != nil {
			t.Errorf("Failed to make request %+v", tc.request)
		}
		if !reflect.DeepEqual(res, tc.response) {
			t.Errorf(
				"Response on request %+v returns %+v expects %+v",
				tc.request,
				res,
				tc.response,
			)
		}
	}
}
