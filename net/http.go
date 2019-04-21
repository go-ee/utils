package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/schema"
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

func PostById(item interface{}, id interface{}, url string, client *http.Client) (err error) {

	var req *http.Request
	var itemJSON []byte

	if itemJSON, err = json.Marshal(item); err != nil {
		log.Fatal("Cannot marshal JSON", err)
		return
	}
	requestUrl := fmt.Sprintf("%v/%v", url, id)
	if req, err = http.NewRequest("POST", requestUrl, bytes.NewBuffer(itemJSON)); err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		log.Fatal("Cannot send POST request", err)
		return
	}
	log.Println("response", requestUrl, resp.Status)
	resp.Body.Close()
	return
}
