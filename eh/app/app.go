package app

import (
	"context"
	"fmt"
	"github.com/go-ee/utils/net/muxlist"
	"go.uber.org/zap"
	"net/http"

	"github.com/go-ee/utils/eh"
	"github.com/go-ee/utils/net"
	"github.com/gorilla/mux"
	"github.com/looplab/eventhorizon"
	"github.com/rs/cors"
)

type AppInfo struct {
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

type AppBase struct {
	*eh.Middleware
	*AppInfo
	*ServerConfig

	ProjectorListener eh.DelegateEventHandler
	SetupCallbacks    []func() error
	Log               *logrus.Entry
	NewContext        func(namespace string) context.Context
	Router            *mux.Router

	Jwt    *net.JwtController
	Secure bool

	notFoundMessage string
}

func NewAppBase(appInfo *AppInfo, serverConfig *ServerConfig, secure bool, middleware *eh.Middleware) (ret *AppBase) {
	ret = &AppBase{
		Middleware:   middleware,
		AppInfo:      appInfo,
		ServerConfig: serverConfig,

		Log: logrus.WithFields(logrus.Fields{"app": appInfo.AppName}),
		NewContext: func(structure string) context.Context {
			return eventhorizon.NewContextWithNamespace(context.Background(), appInfo.AppName+"/"+structure)
		},
		Router:          mux.NewRouter().StrictSlash(true),
		Secure:          secure,
		notFoundMessage: fmt.Sprintf("%v: the page is not found", appInfo.AppName),
	}
	return
}

func (o *AppBase) StartServer() (err error) {
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

	o.Log.Info(listener.List())

	o.Log.Infof("server started, http://%v", o.ServerConfig.Link())
	err = http.ListenAndServe(o.Listen(), nil)
	return
}

func (o *AppBase) NoFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, o.notFoundMessage, http.StatusNotFound)
}
