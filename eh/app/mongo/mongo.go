package mongo

import (
	"github.com/go-ee/utils/eh"
	"github.com/go-ee/utils/eh/app"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	eventbus "github.com/looplab/eventhorizon/eventbus/local"
	eventstore "github.com/looplab/eventhorizon/eventstore/mongodb"
	repo "github.com/looplab/eventhorizon/repo/mongodb"
)

func NewAppMongo(productName string, appName string, secure bool, mongoUrl string) *app.AppBase {
	// Create the event store.
	eventStore := &eh.EventStoreDelegate{Factory: func() (ret eventhorizon.EventStore, err error) {
		return eventstore.NewEventStore("localhost", productName)
	}}

	// Create the event bus that distributes events.
	eventBus := eventbus.NewEventBus(nil)

	// Create the command bus.
	commandBus := bus.NewCommandHandler()

	repos := make(map[string]eventhorizon.ReadWriteRepo)
	readRepos := func(name string, factory func() eventhorizon.Entity) (ret eventhorizon.ReadWriteRepo) {
		if item, ok := repos[name]; !ok {
			ret = &eh.ReadWriteRepoDelegate{Factory: func() (ret eventhorizon.ReadWriteRepo, err error) {
				var retRepo *repo.Repo
				if retRepo, err = repo.NewRepo(mongoUrl, productName, name); err == nil {
					retRepo.SetEntityFactory(factory)
					ret = retRepo
				}
				return
			}}
			repos[name] = ret
		} else {
			ret = item
		}
		return
	}
	return app.NewAppBase(productName, appName, secure, eventStore, eventBus, commandBus, readRepos)
}
