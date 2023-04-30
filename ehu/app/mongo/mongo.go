package mongo

import (
	"github.com/go-ee/utils/ehu"
	"github.com/go-ee/utils/ehu/app"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	eb "github.com/looplab/eventhorizon/eventbus/local"
	es "github.com/looplab/eventhorizon/eventstore/mongodb_v2"
	repo "github.com/looplab/eventhorizon/repo/mongodb"
)

func NewAppMongo(appInfo *app.Info, serverConfig *app.ServerConfig, secure bool, mongoUrl string) *app.Base {
	// Create the event store.
	eventStore := &ehu.EventStoreDelegate{Factory: func() (ret eventhorizon.EventStore, err error) {
		return es.NewEventStore("localhost", appInfo.ProductName)
	}}

	// Create the event bus that distributes events.
	eventBus := eb.NewEventBus(nil)

	// Create the command bus.
	commandBus := bus.NewCommandHandler()

	repos := make(map[string]eventhorizon.ReadWriteRepo)
	reposFactory := func(name string, factory func() eventhorizon.Entity) (ret eventhorizon.ReadWriteRepo, err error) {
		if item, ok := repos[name]; !ok {
			var repoInst *repo.Repo
			if repoInst, err = repo.NewRepo(mongoUrl, appInfo.ProductName, name); err == nil {
				repoInst.SetEntityFactory(factory)
				ret = repoInst
			}
			repos[name] = ret
		} else {
			ret = item
		}
		return
	}
	return app.NewAppBase(appInfo, serverConfig, secure,
		&ehu.Middleware{
			EventStore: eventStore,
			EventBus:   eventBus,
			CommandBus: commandBus,
			Repos:      reposFactory,
		})
}
