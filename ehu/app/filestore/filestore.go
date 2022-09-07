package filestore

import (
	"github.com/go-ee/utils/ehu"
	"github.com/go-ee/utils/ehu/app"
	es "github.com/go-ee/utils/ehu/filestore"
	repo "github.com/go-ee/utils/ehu/filestore"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	eb "github.com/looplab/eventhorizon/eventbus/local"
	"path/filepath"
)

func NewAppFileStore(appInfo *app.Info, serverConfig *app.ServerConfig, secure bool, storeFolder string) *app.Base {

	// Create the event store.
	eventStore := es.NewEventStore(filepath.Join(storeFolder, "eventstore"))

	// Create the event bus that distributes events.
	eventBus := eb.NewEventBus(nil)

	// Create the command bus.
	commandBus := bus.NewCommandHandler()

	repos := make(map[string]eventhorizon.ReadWriteRepo)
	reposFactory := func(name string, factory func() eventhorizon.Entity) (ret eventhorizon.ReadWriteRepo, err error) {
		if item, ok := repos[name]; !ok {
			var repoInst *repo.Repo
			if repoInst, err = repo.NewRepo(filepath.Join(storeFolder, "repos")); err == nil {
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
