package app

import (
	"context"
	"fmt"
	"github.com/go-ee/utils/lg"
	"github.com/go-ee/utils/net/muxlist"
	"github.com/looplab/eventhorizon/namespace"
	"net/http"

	"github.com/go-ee/utils/ehu"
	"github.com/go-ee/utils/net"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Info struct {
	AppName       string
	ProductName   string
	WorkingFolder string
}

type ServerConfig struct {
	ServerAddress string
	ServerPort    int
}

func (o *ServerConfig) Link() (ret string) {
	if o.ServerAddress == "" {
		ret = fmt.Sprintf("127.0.0.1:%v", o.ServerPort)
	} else {
		ret = o.Listen()
	}
	return
}

func (o *ServerConfig) Listen() (ret string) {
	return fmt.Sprintf("%v:%v", o.ServerAddress, o.ServerPort)
}

type Base struct {
	*ehu.Middleware
	*Info
	*ServerConfig

	ProjectorListener ehu.DelegateEventHandler
	SetupCallbacks    []func() error
	NewContext        func(namespace string) context.Context
	Router            *mux.Router

	Jwt    *net.JwtController
	Secure bool

	notFoundMessage string
}

func NewAppBase(appInfo *Info, serverConfig *ServerConfig, secure bool, middleware *ehu.Middleware) (ret *Base) {
	ret = &Base{
		Middleware:   middleware,
		Info:         appInfo,
		ServerConfig: serverConfig,

		NewContext: func(structure string) context.Context {
			return namespace.NewContext(context.Background(), appInfo.AppName+"/"+structure)
		},
		Router:          mux.NewRouter().StrictSlash(true),
		Secure:          secure,
		notFoundMessage: fmt.Sprintf("%v: the page is not found", appInfo.AppName),
	}
	return
}

func (o *Base) StartServer() (err error) {
	o.Router.NotFoundHandler = http.HandlerFunc(o.NoFound)
	if o.Secure {
		o.Router.Path("/logout").Name("Logout").Handler(o.Jwt.LogoutHandler())
		http.Handle("/login", cors.Default().Handler(o.Jwt.LoginHandler()))
		handler := cors.Default().Handler(o.Jwt.ValidateTokenHandler(o.Router))
		http.Handle("/", handler)
	} else {
		handler := cors.Default().Handler(o.Router)
		http.Handle("/", handler)
	}

	listener := muxlist.NewGorillaMuxLister(o.Router)

	lg.LOG.Info(listener.List())

	lg.LOG.Infof("server started, http://%v", o.ServerConfig.Link())
	err = http.ListenAndServe(o.Listen(), nil)
	return
}

func (o *Base) NoFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, o.notFoundMessage, http.StatusNotFound)
}
