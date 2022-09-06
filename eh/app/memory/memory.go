package memory

import (
	"github.com/go-ee/utils/eh"
	"github.com/go-ee/utils/eh/app"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	eb "github.com/looplab/eventhorizon/eventbus/local"
	es "github.com/looplab/eventhorizon/eventstore/memory"
	repo "github.com/looplab/eventhorizon/repo/memory"
)

func NewAppMemory(appInfo *app.Info, serverConfig *app.ServerConfig, secure bool) *app.Base {
	// Create the event store.
	eventStore, _ := es.NewEventStore()

	// Create the event bus that distributes events.
	eventBus := eb.NewEventBus(nil)

	// Create the command bus.
	commandBus := bus.NewCommandHandler()

	repos := make(map[string]eventhorizon.ReadWriteRepo)
	readRepos := func(name string, factory func() eventhorizon.Entity) (ret eventhorizon.ReadWriteRepo, err error) {
		if item, ok := repos[name]; !ok {
			repoInst := repo.NewRepo()
			repoInst.SetEntityFactory(factory)
			ret = repoInst
			repos[name] = ret
		} else {
			ret = item
		}
		return
	}
	return app.NewAppBase(appInfo, serverConfig, secure,
		&eh.Middleware{
			EventStore: eventStore,
			EventBus:   eventBus,
			CommandBus: commandBus,
			Repos:      readRepos,
		})
}
