package muxlist

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/gorilla/mux"
)

type GorillaMuxLister struct {
	router *mux.Router
}

func NewGorillaMuxLister(r *mux.Router) *GorillaMuxLister {
	return &GorillaMuxLister{router: r}
}

func (m *GorillaMuxLister) Extract() ResultSet {

	var result ResultSet

	m.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {

		r := make(Result, 10)
		uri, err := route.GetPathTemplate()
		if err != nil {
			return err
		}

		host, err := route.GetHostTemplate()

		if err == nil {
			r[REQUEST_URI] = host + uri
		} else {
			r[REQUEST_URI] = uri
		}

		r[ROUTE_NAME] = route.GetName()
		r[HTTP_METHODS] = getMethodsAsString(route)

		r[HANDLER_NAME] = GetHumanReadableNameForHandler(route.GetHandler())

		result = append(result, r)
		return nil
	})

	return result
}

//Returns a human readable name for a http Handler
func GetHumanReadableNameForHandler(h http.Handler) string {
	reflectValue := reflect.ValueOf(h)

	if !reflectValue.IsValid() {
		return "SUBROUTER"
	}
	return runtime.FuncForPC(reflectValue.Pointer()).Name()
}

func getMethodsAsString(route *mux.Route) string {
	var routes []string
	routes, _ = route.GetMethods()

	return strings.Join(routes, ",")
}

func (m *GorillaMuxLister) List() (ret string) {
	ret += "\n"
	for _, v := range m.Extract() {
		ret += fmt.Sprintf(`%s`+" \t "+`%s`+" \t "+`%s`+" \t "+`%s`+"\n",
			v[REQUEST_URI],
			v[HTTP_METHODS],
			v[ROUTE_NAME],
			v[HANDLER_NAME])
	}
	return
}

//Constants that map to a key in a Result.
//This is provided to be used as "helpers".
const (
	HTTP_METHODS = iota
	REQUEST_URI
	ROUTE_NAME
	HANDLER_NAME
)

//Represents information for a single route
type Result map[int]string

//ResultSet contains all available information for a multiplexer
type ResultSet []Result
