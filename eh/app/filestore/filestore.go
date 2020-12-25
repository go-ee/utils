package filestore

import (
	"github.com/go-ee/utils/eh"
	"github.com/go-ee/utils/eh/app"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	eb "github.com/looplab/eventhorizon/eventbus/local"
	es "github.com/go-ee/utils/eh/filestore"
	repo "github.com/looplab/eventhorizon/repo/memory"
)

func NewAppFileStore(appInfo *app.AppInfo, serverConfig *app.ServerConfig, secure bool, storeFolder string) *app.AppBase {

	// Create the event store.
	eventStore := es.NewEventStore(storeFolder)

	// Create the event bus that distributes events.
	eventBus := eb.NewEventBus(nil)

	// Create the command bus.
	commandBus := bus.NewCommandHandler()

	repos := make(map[string]eventhorizon.ReadWriteRepo)
	reposFactory := func(name string, factory func() eventhorizon.Entity) (ret eventhorizon.ReadWriteRepo) {
		if item, ok := repos[name]; !ok {
			ret = &eh.ReadWriteRepoDelegate{Factory: func() (ret eventhorizon.ReadWriteRepo, err error) {
				retRepo := repo.NewRepo()
				retRepo.SetEntityFactory(factory)
				ret = retRepo
				return
			}}
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
			Repos:      reposFactory,
		})
}
