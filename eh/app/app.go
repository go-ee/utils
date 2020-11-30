package app

import (
	"context"
	"fmt"
	"github.com/go-ee/utils/net/muxlist"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/go-ee/utils/eh"
	"github.com/go-ee/utils/net"
	"github.com/gorilla/mux"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	"github.com/rs/cors"
)

type AppBase struct {
	ProductName       string
	Name              string
	EventStore        eventhorizon.EventStore
	EventBus          eventhorizon.EventBus
	CommandBus        *bus.CommandHandler
	ProjectorListener eh.DelegateEventHandler
	SetupCallbacks    []func() error
	ReadRepos         func(name string, factory func() eventhorizon.Entity) eventhorizon.ReadWriteRepo

	Log        *logrus.Entry
	NewContext func(namespace string) context.Context
	Router     *mux.Router
	Jwt        *net.JwtController
	Secure     bool

	serverAddress string
	serverPort    int

	notFoundMessage string
}

func NewAppBase(productName string, appName string, secure bool, serverAddress string, serverPort int,
	eventStore eventhorizon.EventStore, eventBus eventhorizon.EventBus, commandBus *bus.CommandHandler,
	readRepos func(name string, factorySetup func() eventhorizon.Entity) eventhorizon.ReadWriteRepo) (ret *AppBase) {
	ret = &AppBase{
		ProductName: productName,
		Name:        appName,
		Secure:      secure,
		EventStore:  eventStore,
		EventBus:    eventBus,
		CommandBus:  commandBus,
		ReadRepos:   readRepos,

		Log: logrus.WithFields(logrus.Fields{"app": appName}),
		NewContext: func(structure string) context.Context {
			return eventhorizon.NewContextWithNamespace(context.Background(), appName+"/"+structure)
		},
		Router: mux.NewRouter().StrictSlash(true),

		serverAddress: serverAddress,
		serverPort:    serverPort,

		notFoundMessage: fmt.Sprintf("%v: the page is not found", appName),
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

	linkAddress := o.serverAddress
	if linkAddress == "" {
		linkAddress = "127.0.0.1"
	}

	o.Log.Infof("server started, http://%v:%v", linkAddress, o.serverPort)
	err = http.ListenAndServe(fmt.Sprintf("%v:%v", o.serverAddress, o.serverPort), nil)
	return
}

func (o *AppBase) NoFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, o.notFoundMessage, http.StatusNotFound)
}
