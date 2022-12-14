package jsonrpc_server

import (
	"context"
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	Address string
	Router  Router
	Ctx     context.Context
	Timeout int
}

type Request struct {
	JsonRPC string      `json:"jsonrpc" validate:"required,eq=2.0"`
	Method  string      `json:"method" validate:"required"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id,omitempty"`
}

type Response struct {
	JsonRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

const (
	ErrorParseRequestCode   = -32700
	ErrorInvalidRequestCode = -32600
	ErrorMethodNotFoundCode = -32601
	ErrorInvalidParamsCode  = -32602
	ErrorInternalErrorCode  = -32604

	ErrorParseRequestMessage   = "Invalid JSON was received by the server."
	ErrorInvalidRequestMessage = "The JSON sent is not a valid Request object."
	ErrorMethodNotFoundMessage = "The method does not exist / is not available."
	ErrorInvalidParamsMessage  = "Invalid method parameter(s)."
	ErrorInternalErrorMessage  = "Internal JSON-RPC error."
	ErrorTimeoutMessage        = "Request timeout was reached"
)

type Router interface {
	ResolveWorker(method string) (Worker, *Error)
}

type Worker interface {
	Handle(context.Context) (interface{}, *Error)
}

func (s *Server) Run() {
	http.HandleFunc("/", s.ProcessRequest)

	err := http.ListenAndServe(s.Address, nil)

	if err != nil {
		panic(err)
	}
}

func (s *Server) ProcessRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var wg sync.WaitGroup
	var resps []Response

	b, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	reqs, e := unmarshalRequest(b)

	if e != nil {
		res := Response{JsonRPC: "2.0"}
		res.Error = e
		jsonResp, _ := json.Marshal(res)
		w.Write(jsonResp)
		return
	}

	for _, req := range reqs {
		wg.Add(1)
		go func(req Request) {
			defer wg.Done()
			res := Response{JsonRPC: "2.0"}
			res.ID = req.ID

			defer func() {
				if r := recover(); r != nil {
					var e = Error{Code: ErrorInternalErrorCode, Message: ErrorInternalErrorMessage}
					res.Error = &e
					resps = append(resps, res)
				}
			}()

			// ???????????????? ???????????????????? ?????????????? ??????????????
			if !validate(req) {
				var e = Error{Code: ErrorInvalidRequestCode, Message: ErrorInvalidRequestMessage}
				res.Error = &e
				resps = append(resps, res)
				return
			}

			// ???????? ???????????????????? ????????????
			wrk, e := s.Router.ResolveWorker(req.Method)
			if e != nil {
				res.Error = e
				resps = append(resps, res)
				return
			}

			if req.Params != nil {
				j, _ := json.Marshal(req.Params)
				err := json.Unmarshal(j, &wrk)

				// ?????????????????? ????????????????????, ?????????????? ???????????????????? ?? ????????????????????
				if err != nil || !validate(wrk) {
					var e = Error{Code: ErrorInvalidParamsCode, Message: ErrorInvalidParamsMessage}
					res.Error = &e
					resps = append(resps, res)
					return
				}
			}

			data, e := wrk.Handle(s.Ctx)
			if e != nil {
				res.Error = e
				resps = append(resps, res)
				return
			}

			res.Result = data
			resps = append(resps, res)
			return
		}(req)
	}

	timeout := isReachTimeout(&wg, time.Duration(s.Timeout)*time.Second)

	if timeout {
		var e = Error{Code: ErrorInternalErrorCode, Message: ErrorTimeoutMessage}
		res := Response{
			JsonRPC: "2.0",
			Result:  nil,
			Error:   &e,
			ID:      nil,
		}
		jsonResp, _ := json.Marshal(res)
		w.Write(jsonResp)
	} else {
		if len(resps) == 1 {
			jsonResp, _ := json.Marshal(resps[0])
			w.Write(jsonResp)
		} else {
			jsonResp, _ := json.Marshal(resps)
			w.Write(jsonResp)
		}
	}

	return
}

func validate(req interface{}) bool {
	v := validator.New()
	err := v.Struct(req)

	return err == nil
}

func unmarshalRequest(b []byte) ([]Request, *Error) {
	var e = &Error{Code: ErrorParseRequestCode, Message: ErrorParseRequestMessage}
	if len(b) == 0 {
		return make([]Request, 0), e
	}
	switch b[0] {
	case '{':
		return unmarshalSingle(b)
	case '[':
		return unmarshalMany(b)
	}
	return make([]Request, 0), e
}

func unmarshalMany(b []byte) ([]Request, *Error) {
	var e = &Error{Code: ErrorParseRequestCode, Message: ErrorParseRequestMessage}
	var reqs []Request
	err := json.Unmarshal(b, &reqs)
	if err != nil {
		return reqs, e
	}

	return reqs, nil
}

func unmarshalSingle(b []byte) ([]Request, *Error) {
	var e = &Error{Code: ErrorParseRequestCode, Message: ErrorParseRequestMessage}
	req := Request{}
	err := json.Unmarshal(b, &req)
	if err != nil {
		return []Request{req}, e
	}

	return []Request{req}, nil
}

func isReachTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	done := make(chan struct{})

	go func() {
		defer close(done)
		wg.Wait()
	}()

	select {
	case <-done:
		return false

	case <-time.After(timeout):
		return true
	}
}
