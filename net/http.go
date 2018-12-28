package net

import (
	"net/http"
	"encoding/json"
	"github.com/gorilla/schema"
	"io"
)

const GET = "GET"
const POST = "POST"
const PUT = "PUT"
const DELETE = "DELETE"

const QueryType = "qType"
const QueryTypeCount = "count"
const QueryTypeExist = "exist"
const QueryTypeFind = "find"

const Command = "command"

func ResponseJson(response interface{}, w http.ResponseWriter) {
	ResponseJsonCode(response, http.StatusOK, w)
}

func ResponseJsonCode(response interface{}, code int, w http.ResponseWriter) {

	json, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

type Result struct {
	Ok  bool
	Err error
	Msg string
}

func ResponseResultErr(err error, msg string, code int, w http.ResponseWriter) {
	ResponseJsonCode(Result{Ok: false, Msg: msg, Err: err}, code, w)
}

func ResponseResultOk(msg string, w http.ResponseWriter) {
	ResponseJson(Result{Ok: true, Msg: msg}, w)
}

func Decode(item interface{}, r *http.Request) (err error) {
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(item)
	defer r.Body.Close()
	if err == io.EOF {
		err = nil
	}

	//decode url params to command
	if err == nil {
		if err = r.ParseForm(); err == nil {
			newDecoder := schema.NewDecoder()
			newDecoder.IgnoreUnknownKeys(true)
			err = newDecoder.Decode(item, r.Form)
		}
	}

	if err == io.EOF {
		err = nil
	}
	return
}
