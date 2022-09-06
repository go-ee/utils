package net

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-ee/utils/lg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/schema"
)

const QueryType = "qType"
const QueryTypeCount = "count"
const QueryTypeExist = "exist"
const QueryTypeFind = "find"

const Command = "command"

func ResponseJson(response interface{}, w http.ResponseWriter) error {
	return ResponseJsonCode(response, http.StatusOK, w)
}

func ResponseJsonCode(response interface{}, code int, w http.ResponseWriter) (err error) {

	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(jsonData)
	return
}

type Result struct {
	Ok   bool        `json:"ok,omitempty"`
	Item interface{} `json:"item,omitempty"`
	Msg  string      `json:"msg,omitempty"`
	Err  string      `json:"err,omitempty"`
}

func ResponseResultErr(err error, msg string, item interface{}, code int, w http.ResponseWriter) error {
	return ResponseJsonCode(Result{Ok: false, Msg: msg, Err: err.Error()}, code, w)
}

func ResponseResultOk(msg string, item interface{}, w http.ResponseWriter) error {
	return ResponseJson(Result{Ok: true, Item: item, Msg: msg}, w)
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
		log.Fatal("cannot marshal JSON", err)
		return
	}
	requestUrl := fmt.Sprintf("%v/%v", url, id)
	if req, err = http.NewRequest(http.MethodPost, requestUrl, bytes.NewBuffer(itemJSON)); err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		log.Fatal("cannot send POST request", err)
		return
	}
	log.Println("response", requestUrl, resp.Status)
	resp.Body.Close()
	return
}

func DeleteById(id interface{}, url string, client *http.Client) (err error) {
	var req *http.Request

	requestUrl := fmt.Sprintf("%v/%v", url, id)
	if req, err = http.NewRequest(http.MethodDelete, requestUrl, nil); err != nil {
		return
	}

	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		log.Fatal("cannot send DELETE request", err)
		return
	}
	log.Println("response", requestUrl, resp.Status)
	resp.Body.Close()
	return
}

func GetItems(items interface{}, url string, client *http.Client) (err error) {

	var req *http.Request

	if req, err = http.NewRequest(http.MethodGet, url, nil); err != nil {
		return
	}

	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		log.Fatal("cannot send GET request", err)
		return
	}

	defer resp.Body.Close()
	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(bodyBytes, &items)

	return
}

func FormatRequestFrom(r *http.Request) string {
	var request []string
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	request = append(request, fmt.Sprintf("RemoteAddr: %v", r.RemoteAddr))
	return strings.Join(request, "\n")
}

// formatRequest generates ascii representation of a request
func FormatRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return strings.Join(request, "\n")
}

func GetQueryOrFormValue(paramName string, r *http.Request) (ret string) {
	if ret = r.URL.Query().Get(paramName); ret == "" {
		if r.Method == "POST" {
			ret = r.FormValue(paramName)
		}
	}
	return
}

func CorsAllowAll(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func CorsWrap(allowPattern string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", allowPattern)
		h.ServeHTTP(w, r)
	})
}

func LogBody(w http.ResponseWriter, r *http.Request) bool {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		lg.LOG.Infof("error reading body: %v", err)
		http.Error(w, "can't read body", http.StatusBadRequest)
		return true
	} else {
		lg.LOG.Infof("body: %s", body)
	}
	return false
}
