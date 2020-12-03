package eh

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/go-ee/utils/enum"
	"github.com/go-ee/utils/net"
	"github.com/google/uuid"
	"github.com/looplab/eventhorizon"
	"github.com/looplab/eventhorizon/aggregatestore/events"
	"github.com/looplab/eventhorizon/commandhandler/aggregate"
	"github.com/looplab/eventhorizon/commandhandler/bus"
	"github.com/looplab/eventhorizon/eventhandler/projector"
	"github.com/pkg/errors"
)

type AggregateInitializer struct {
	ctx              context.Context
	cancel           context.CancelFunc
	aggregateType    eventhorizon.AggregateType
	aggregateFactory func(id uuid.UUID) eventhorizon.Aggregate
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
	aggregateFactory func(id uuid.UUID) eventhorizon.Aggregate,
	entityFactory func() eventhorizon.Entity,
	commands []enum.Literal, events []enum.Literal,
	projectorListener DelegateEventHandler,
	setupCallbacks []func() error, eventStore eventhorizon.EventStore, eventBus eventhorizon.EventBus, commandBus *bus.CommandHandler,
	readRepos func(name string, factory func() eventhorizon.Entity) eventhorizon.ReadWriteRepo) (ret *AggregateInitializer) {
	ctx, cancel := context.WithCancel(context.Background())
	ret = &AggregateInitializer{
		ctx:               ctx,
		cancel:            cancel,
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
			if err = callback(); err != nil {
				return
			}
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
	proj := projector.NewEventHandler(NewProjector(projectorType, listener), repo)
	proj.SetEntityFactory(o.entityFactory)
	err = o.RegisterForAllEvents(proj)
	ret = repo
	return
}

func (o *AggregateInitializer) RegisterForAllEvents(handler eventhorizon.EventHandler) (err error) {
	eventTypes := make([]eventhorizon.EventType, len(o.events))
	for i, v := range o.events {
		eventTypes[i] = eventhorizon.EventType(v.Name())
	}
	err = o.eventBus.AddHandler(o.ctx, eventhorizon.MatchEvents(eventTypes), handler)
	return
}

func (o *AggregateInitializer) RegisterForEvent(handler eventhorizon.EventHandler, event enum.Literal) (err error) {
	err = o.eventBus.AddHandler(o.ctx, eventhorizon.MatchEvents([]eventhorizon.EventType{eventhorizon.EventType(event.Name())}), handler)
	return
}

type AggregateStoreEvent interface {
	AppendEvent(eventhorizon.EventType, eventhorizon.EventData, time.Time) eventhorizon.Event
}

type DelegateCommandHandler interface {
	Execute(cmd eventhorizon.Command, entity eventhorizon.Entity, store AggregateStoreEvent) error
}

type DelegateEventHandler interface {
	Apply(event eventhorizon.Event, entity eventhorizon.Entity) error
}

type Entity interface {
	EntityID() uuid.UUID
	Deleted() *time.Time
}

type AggregateBase struct {
	*events.AggregateBase
	DelegateCommandHandler
	DelegateEventHandler
	Entity eventhorizon.Entity
}

func (o *AggregateBase) HandleCommand(ctx context.Context, cmd eventhorizon.Command) error {
	return o.Execute(cmd, o.Entity, o.AggregateBase)
}

func (o *AggregateBase) ApplyEvent(ctx context.Context, event eventhorizon.Event) error {
	return o.Apply(event, o.Entity)
}

func NewAggregateBase(aggregateType eventhorizon.AggregateType, id uuid.UUID,
	commandHandler DelegateCommandHandler, eventHandler DelegateEventHandler,
	entity eventhorizon.Entity) *AggregateBase {
	return &AggregateBase{
		AggregateBase:          events.NewAggregateBase(aggregateType, id),
		DelegateCommandHandler: commandHandler,
		DelegateEventHandler:   eventHandler,
		Entity:                 entity,
	}
}

func CommandHandlerNotImplemented(commandType eventhorizon.CommandType) error {
	return errors.New(fmt.Sprintf("handler not implemented for %v", commandType))
}

func EventHandlerNotImplemented(eventType eventhorizon.EventType) error {
	return errors.New(fmt.Sprintf("handler not implemented for %v", eventType))
}

func EntityAlreadyExists(entityId uuid.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("entity already exists with id=%v and aggregateType=%v", entityId, aggregateType))
}

func EntityNotExists(entityId uuid.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("entity not exists with id=%v and aggregateType=%v", entityId, aggregateType))
}

func IdNotDefined(currentId uuid.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("id not defined for aggregateType=%v", aggregateType))
}

func IdsDismatch(entityId uuid.UUID, currentId uuid.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("Dismatch entity id and current id, %v != %v, for aggregateType=%v",
		entityId, currentId, aggregateType))
}

func QueryNotImplemented(queryName string) error {
	return errors.New(fmt.Sprintf("Query not implemented for %v", queryName))
}

func ValidateNewId(entityId uuid.UUID, currentId uuid.UUID, aggregateType eventhorizon.AggregateType) (ret error) {
	if entityId != uuid.Nil {
		ret = EntityAlreadyExists(entityId, aggregateType)
	} else if currentId == uuid.Nil {
		ret = IdNotDefined(currentId, aggregateType)
	}
	return
}

func ValidateIdsMatch(entityId uuid.UUID, currentId uuid.UUID, aggregateType eventhorizon.AggregateType) (ret error) {
	if entityId == uuid.Nil {
		ret = EntityNotExists(currentId, aggregateType)
	} else if entityId != currentId {
		ret = IdsDismatch(entityId, currentId, aggregateType)
	}
	return
}

type HttpQueryHandler struct {
}

func NewHttpQueryHandlerFull() (ret *HttpQueryHandler) {
	ret = &HttpQueryHandler{}
	return
}

func (o *HttpQueryHandler) HandleResult(
	ret interface{}, err error, method string, w http.ResponseWriter, r *http.Request) {

	if err == nil {
		var jsonData []byte
		if jsonData, err = json.Marshal(ret); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonData)
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

func NewHttpCommandHandlerFull(
	context context.Context, commandBus eventhorizon.CommandHandler) (ret *HttpCommandHandler) {

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
		net.ResponseResultErr(err, fmt.Sprintf("failed, command %T, %v", command, command),
			http.StatusExpectationFailed, w)
	} else {
		net.ResponseResultOk(fmt.Sprintf("succefully, command %T, %v, %v", command, command,
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

func (o *ProjectorEventHandler) Project(
	ctx context.Context, event eventhorizon.Event, entity eventhorizon.Entity) (ret eventhorizon.Entity, err error) {

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

func (o *ReadWriteRepoDelegate) Remove(ctx context.Context, id uuid.UUID) (err error) {
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

func (o *ReadWriteRepoDelegate) Find(ctx context.Context, id uuid.UUID) (ret eventhorizon.Entity, err error) {
	var repo eventhorizon.ReadWriteRepo
	if repo, err = o.delegate(); err == nil {
		ret, err = repo.Find(ctx, id)
	}
	return
}

func (o *ReadWriteRepoDelegate) FindAll(ctx context.Context) (ret []eventhorizon.Entity, err error) {
	var repo eventhorizon.ReadWriteRepo
	if repo, err = o.delegate(); err == nil {
		if ret, err = repo.FindAll(ctx); err == nil {
			ret = o.FilterDeleted(ctx, ret)
		}
	}
	return
}

func (o *ReadWriteRepoDelegate) FilterDeleted(ctx context.Context, ret []eventhorizon.Entity) []eventhorizon.Entity {
	n := 0
	for _, x := range ret {
		if e, ok := x.(Entity); ok {
			if e.Deleted() == nil {
				ret[n] = x
				n++
			} else {
				o.repo.Remove(ctx, e.EntityID())
			}
		} else {
			ret[n] = x
			n++
		}
	}
	ret = ret[:n]
	return ret
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

func (o *EventStoreDelegate) Load(ctx context.Context, id uuid.UUID) (ret []eventhorizon.Event, err error) {
	var eventStore eventhorizon.EventStore
	if eventStore, err = o.delegate(); err == nil {
		ret, err = eventStore.Load(ctx, id)
	}
	return
}
