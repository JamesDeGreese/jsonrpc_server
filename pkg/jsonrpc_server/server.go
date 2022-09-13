package jsonrpc_server

import (
	"context"
	"encoding/json"
	"github.com/go-playground/validator/v10"
	"io/ioutil"
	"net/http"
	"sync"
)

type Server struct {
	Address string
	Router  Router
	Ctx     context.Context
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

			// Проверка валидности формата запроса
			if !validate(req) {
				res.Error = &Error{Code: -32600, Message: "The JSON sent is not a valid Request object."}
				resps = append(resps, res)
				return
			}

			// Ищем обработчик метода
			wrk, e := s.Router.ResolveWorker(req.Method)
			if e != nil {
				res.Error = e
				resps = append(resps, res)
				return
			}

			if req.Params != nil {
				j, _ := json.Marshal(req.Params)
				err := json.Unmarshal(j, &wrk)

				// Валидация параметров, которые передаются в обработчик
				if err != nil || !validate(wrk) {
					res.Error = &Error{Code: -32602, Message: "Invalid method parameter(s)."}
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

	wg.Wait()

	if len(resps) == 1 {
		jsonResp, _ := json.Marshal(resps[0])
		w.Write(jsonResp)
	} else {
		jsonResp, _ := json.Marshal(resps)
		w.Write(jsonResp)
	}

	return
}

func validate(req interface{}) bool {
	v := validator.New()
	err := v.Struct(req)

	return err == nil
}

func unmarshalRequest(b []byte) ([]Request, *Error) {
	if len(b) == 0 {
		return make([]Request, 0), &Error{Code: -32700, Message: "Invalid JSON was received by the server."}
	}
	switch b[0] {
	case '{':
		return unmarshalSingle(b)
	case '[':
		return unmarshalMany(b)
	}
	return make([]Request, 0), &Error{Code: -32700, Message: "Invalid JSON was received by the server."}
}

func unmarshalMany(b []byte) ([]Request, *Error) {
	var reqs []Request
	err := json.Unmarshal(b, &reqs)
	if err != nil {
		return reqs, &Error{Code: -32700, Message: "Invalid JSON was received by the server."}
	}

	return reqs, nil
}

func unmarshalSingle(b []byte) ([]Request, *Error) {
	req := Request{}
	err := json.Unmarshal(b, &req)
	if err != nil {
		return []Request{req}, &Error{Code: -32700, Message: "Invalid JSON was received by the server."}
	}

	return []Request{req}, nil
}
