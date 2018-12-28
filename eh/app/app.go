package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/eugeis/gee/eh"
	"github.com/eugeis/gee/lg"
	"github.com/eugeis/gee/net"
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

	Log    *lg.DebugLogger
	Ctx    context.Context
	Router *mux.Router
	Jwt    *net.JwtController
	Secure bool

	notFoundMessage string
}

func NewAppBase(productName string, appName string, secure bool, eventStore eventhorizon.EventStore, eventBus eventhorizon.EventBus,
	commandBus *bus.CommandHandler,
	readRepos func(name string, factory func() eventhorizon.Entity) eventhorizon.ReadWriteRepo) (ret *AppBase) {
	ret = &AppBase{
		ProductName: productName,
		Name:        appName,
		Secure:      secure,
		EventStore:  eventStore,
		EventBus:    eventBus,
		CommandBus:  commandBus,
		ReadRepos:   readRepos,

		Log:    lg.NewLogger(appName),
		Ctx:    eventhorizon.NewContextWithNamespace(context.Background(), appName),
		Router: mux.NewRouter().StrictSlash(true),

		notFoundMessage: fmt.Sprintf("%v: the page is not found", appName),
	}
	return
}

func (o *AppBase) StartServer() {
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

	o.Log.Info("Server started %v", "127.0.0.1:8080")
	o.Log.Err("%v", http.ListenAndServe("127.0.0.1:8080", nil))
}

func (o *AppBase) NoFound(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, o.notFoundMessage, http.StatusNotFound)
}
