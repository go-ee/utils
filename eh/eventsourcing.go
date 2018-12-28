package eh

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/eugeis/gee/enum"
	"github.com/eugeis/gee/net"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/aggregatestore/events"
	"github.com/looplab/eventhorizon/commandhandler/aggregate"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	"github.com/looplab/eventhorizon/eventhandler/projector"
	"github.com/pkg/errors"
)

type AggregateInitializer struct {
	aggregateType    eventhorizon.AggregateType
	aggregateFactory func(id eventhorizon.UUID) eventhorizon.Aggregate
	entityFactory    func() eventhorizon.Entity
	commands         []enum.Literal
	events           []enum.Literal

	eventStore              eventhorizon.EventStore
	eventBus                eventhorizon.EventBus
	commandBus              *bus.CommandHandler
	aggregateStore          *events.AggregateStore
	commandHandler          *aggregate.CommandHandler
	projectorListener       DelegateEventHandler
	setupCallbacks          []func() error
	readRepos               func(name string, factory func() eventhorizon.Entity) eventhorizon.ReadWriteRepo
	DefaultProjectorEnabled bool
	ProjectorRepo           eventhorizon.ReadRepo
}

func NewAggregateInitializer(aggregateType eventhorizon.AggregateType,
	aggregateFactory func(id eventhorizon.UUID) eventhorizon.Aggregate,
	entityFactory func() eventhorizon.Entity,
	commands []enum.Literal, events []enum.Literal,
	projectorListener DelegateEventHandler,
	setupCallbacks []func() error, eventStore eventhorizon.EventStore, eventBus eventhorizon.EventBus, commandBus *bus.CommandHandler,
	readRepos func(name string, factory func() eventhorizon.Entity) eventhorizon.ReadWriteRepo) (ret *AggregateInitializer) {
	ret = &AggregateInitializer{
		aggregateType:     aggregateType,
		aggregateFactory:  aggregateFactory,
		entityFactory:     entityFactory,
		commands:          commands,
		events:            events,
		projectorListener: projectorListener,
		setupCallbacks:    setupCallbacks,

		eventStore:              eventStore,
		eventBus:                eventBus,
		commandBus:              commandBus,
		readRepos:               readRepos,
		DefaultProjectorEnabled: true,
	}
	return
}

func (o *AggregateInitializer) Setup() (err error) {
	//register aggregate factory
	eventhorizon.RegisterAggregate(o.aggregateFactory)

	if o.aggregateStore, err = events.NewAggregateStore(o.eventStore, o.eventBus); err != nil {
		return
	}

	if o.commandHandler, err = aggregate.NewCommandHandler(o.aggregateType, o.aggregateStore); err != nil {
		return
	}

	if err = o.registerCommands(); err != nil {
		return
	}

	if o.DefaultProjectorEnabled {
		if err = o.registerProjector(); err != nil {
			return
		}
	}

	if o.setupCallbacks != nil {
		for _, callback := range o.setupCallbacks {
			callback()
		}
	}
	return
}

func (o *AggregateInitializer) registerCommands() (err error) {
	for _, item := range o.commands {
		if err = o.commandBus.SetHandler(o.commandHandler, eventhorizon.CommandType(item.Name())); err != nil {
			return
		}
	}
	return
}

func (o *AggregateInitializer) registerProjector() (err error) {
	o.ProjectorRepo, err = o.RegisterProjector(o.projectorListener)
	return
}

func (o *AggregateInitializer) RegisterProjector(listener DelegateEventHandler) (ret eventhorizon.ReadRepo, err error) {
	projectorType := string(o.aggregateType)
	repo := o.readRepos(projectorType, o.entityFactory)
	projector := projector.NewEventHandler(NewProjector(projectorType, listener), repo)
	projector.SetEntityFactory(o.entityFactory)
	o.RegisterForAllEvents(projector)
	ret = repo
	return
}

func (o *AggregateInitializer) RegisterForAllEvents(handler eventhorizon.EventHandler) {
	for _, item := range o.events {
		o.eventBus.AddHandler(eventhorizon.MatchEvent(eventhorizon.EventType(item.Name())), handler)
	}
}

func (o *AggregateInitializer) RegisterForEvent(handler eventhorizon.EventHandler, event enum.Literal) {
	o.eventBus.AddHandler(eventhorizon.MatchEvent(eventhorizon.EventType(event.Name())), handler)
}

type AggregateStoreEvent interface {
	StoreEvent(eventhorizon.EventType, eventhorizon.EventData, time.Time) eventhorizon.Event
}

type DelegateCommandHandler interface {
	Execute(cmd eventhorizon.Command, entity eventhorizon.Entity, store AggregateStoreEvent) error
}

type DelegateEventHandler interface {
	Apply(event eventhorizon.Event, entity eventhorizon.Entity) error
}

type AggregateBase struct {
	*events.AggregateBase
	DelegateCommandHandler
	DelegateEventHandler
	Model eventhorizon.Entity
}

func (o *AggregateBase) HandleCommand(ctx context.Context, cmd eventhorizon.Command) error {
	return o.Execute(cmd, o.Model, o.AggregateBase)
}

func (o *AggregateBase) ApplyEvent(ctx context.Context, event eventhorizon.Event) error {
	return o.Apply(event, o.Model)
}

func NewAggregateBase(aggregateType eventhorizon.AggregateType, id eventhorizon.UUID,
	commandHandler DelegateCommandHandler, eventHandler DelegateEventHandler,
	entity eventhorizon.Entity) *AggregateBase {
	return &AggregateBase{
		AggregateBase:          events.NewAggregateBase(aggregateType, id),
		DelegateCommandHandler: commandHandler,
		DelegateEventHandler:   eventHandler,
		Model:                  entity,
	}
}

func CommandHandlerNotImplemented(commandType eventhorizon.CommandType) error {
	return errors.New(fmt.Sprintf("Handler not implemented for %v", commandType))
}

func EventHandlerNotImplemented(eventType eventhorizon.EventType) error {
	return errors.New(fmt.Sprintf("Handler not implemented for %v", eventType))
}

func EntityAlreadyExists(entityId eventhorizon.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("Entity already exists with id=%v and aggregateType=%v", entityId, aggregateType))
}

func EntityNotExists(entityId eventhorizon.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("Entity not exists with id=%v and aggregateType=%v", entityId, aggregateType))
}

func IdNotDefined(currentId eventhorizon.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("Id not defined for aggregateType=%v", aggregateType))
}

func IdsDismatch(entityId eventhorizon.UUID, currentId eventhorizon.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("Dismatch entity id and current id, %v != %v, for aggregateType=%v",
		entityId, currentId, aggregateType))
}

func QueryNotImplemented(queryName string) error {
	return errors.New(fmt.Sprintf("Query not implemented for %v", queryName))
}

func ValidateNewId(entityId eventhorizon.UUID, currentId eventhorizon.UUID, aggregateType eventhorizon.AggregateType) (ret error) {
	if len(entityId) > 0 {
		ret = EntityAlreadyExists(entityId, aggregateType)
	} else if len(currentId) == 0 {
		ret = IdNotDefined(currentId, aggregateType)
	}
	return
}

func ValidateIdsMatch(entityId eventhorizon.UUID, currentId eventhorizon.UUID, aggregateType eventhorizon.AggregateType) (ret error) {
	if len(entityId) == 0 {
		ret = EntityNotExists(currentId, aggregateType)
	} else if entityId != currentId {
		ret = IdsDismatch(entityId, currentId, aggregateType)
	}
	return
}

type HttpQueryHandler struct {
}

func NewHttpQueryHandler() (ret *HttpQueryHandler) {
	ret = &HttpQueryHandler{}
	return
}

func (o *HttpQueryHandler) HandleResult(ret interface{}, err error, method string, w http.ResponseWriter, r *http.Request) {
	if err == nil {
		var js []byte
		if js, err = json.Marshal(ret); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write(js)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

type HttpCommandHandler struct {
	Context    context.Context
	CommandBus eventhorizon.CommandHandler
}

func NewHttpCommandHandler(context context.Context, commandBus eventhorizon.CommandHandler) (ret *HttpCommandHandler) {
	ret = &HttpCommandHandler{
		Context:    context,
		CommandBus: commandBus,
	}
	return
}

func (o *HttpCommandHandler) HandleCommand(command eventhorizon.Command, w http.ResponseWriter, r *http.Request) {
	//decode body to command
	err := net.Decode(command, r)

	if err != nil && err != io.EOF {
		net.ResponseResultErr(err, fmt.Sprintf("Can't decode body to command %T", command),
			http.StatusBadRequest, w)
		return
	}

	if err = o.CommandBus.HandleCommand(o.Context, command); err != nil {
		net.ResponseResultErr(err, fmt.Sprintf("Can't execute command %T %v", command, command),
			http.StatusExpectationFailed, w)
	} else {
		net.ResponseResultOk(fmt.Sprintf("Succefully executed command %T %v from %v", command, command,
			html.EscapeString(r.URL.Path)), w)
	}
}

type ProjectorEventHandler struct {
	DelegateEventHandler
	projectorType projector.Type
}

func NewProjector(projectorType string, eventHandler DelegateEventHandler) (ret *ProjectorEventHandler) {
	ret = &ProjectorEventHandler{
		projectorType:        projector.Type(projectorType),
		DelegateEventHandler: eventHandler,
	}
	return
}

func (o *ProjectorEventHandler) ProjectorType() projector.Type {
	return o.projectorType
}

func (o *ProjectorEventHandler) Project(ctx context.Context, event eventhorizon.Event, entity eventhorizon.Entity) (ret eventhorizon.Entity, err error) {
	ret = entity
	err = o.Apply(event, entity)
	return
}

type ReadWriteRepoDelegate struct {
	Factory func() (ret eventhorizon.ReadWriteRepo, err error)
	repo    eventhorizon.ReadWriteRepo
}

func (o *ReadWriteRepoDelegate) delegate() (ret eventhorizon.ReadWriteRepo, err error) {
	if o.repo == nil {
		o.repo, err = o.Factory()
	}
	ret = o.repo
	return
}

func (o *ReadWriteRepoDelegate) Save(ctx context.Context, entity eventhorizon.Entity) (err error) {
	var repo eventhorizon.ReadWriteRepo
	if repo, err = o.delegate(); err == nil {
		err = repo.Save(ctx, entity)
	}
	return
}

func (o *ReadWriteRepoDelegate) Remove(ctx context.Context, id eventhorizon.UUID) (err error) {
	var repo eventhorizon.ReadWriteRepo
	if repo, err = o.delegate(); err == nil {
		err = repo.Remove(ctx, id)
	}
	return
}

func (o *ReadWriteRepoDelegate) Parent() (ret eventhorizon.ReadRepo) {
	if repo, err := o.delegate(); err == nil {
		ret = repo.Parent()
	}
	return
}

func (o *ReadWriteRepoDelegate) Find(ctx context.Context, id eventhorizon.UUID) (ret eventhorizon.Entity, err error) {
	var repo eventhorizon.ReadWriteRepo
	if repo, err = o.delegate(); err == nil {
		ret, err = repo.Find(ctx, id)
	}
	return
}

func (o *ReadWriteRepoDelegate) FindAll(ctx context.Context) (ret []eventhorizon.Entity, err error) {
	var repo eventhorizon.ReadWriteRepo
	if repo, err = o.delegate(); err == nil {
		ret, err = repo.FindAll(ctx)
	}
	return
}

type EventStoreDelegate struct {
	Factory    func() (ret eventhorizon.EventStore, err error)
	eventStore eventhorizon.EventStore
}

func (o *EventStoreDelegate) delegate() (ret eventhorizon.EventStore, err error) {
	if o.eventStore == nil {
		o.eventStore, err = o.Factory()
	}
	ret = o.eventStore
	return
}

func (o *EventStoreDelegate) Save(ctx context.Context, events []eventhorizon.Event, originalVersion int) (err error) {
	var eventStore eventhorizon.EventStore
	if eventStore, err = o.delegate(); err == nil {
		err = eventStore.Save(ctx, events, originalVersion)
	}
	return
}

func (o *EventStoreDelegate) Load(ctx context.Context, aggregateType eventhorizon.AggregateType, id eventhorizon.UUID) (ret []eventhorizon.Event, err error) {
	var eventStore eventhorizon.EventStore
	if eventStore, err = o.delegate(); err == nil {
		ret, err = eventStore.Load(ctx, aggregateType, id)
	}
	return
}
