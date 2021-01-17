package eh

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/looplab/eventhorizon/eventhandler/saga"
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

type Middleware struct {
	EventStore eventhorizon.EventStore
	EventBus   eventhorizon.EventBus
	CommandBus *bus.CommandHandler
	Repos      func(string, func() (ret eventhorizon.Entity)) (ret eventhorizon.ReadWriteRepo, err error)
}

type AggregateEngine struct {
	*Middleware

	ctx    context.Context
	cancel context.CancelFunc

	AggregateType    eventhorizon.AggregateType
	AggregateFactory func(id uuid.UUID) eventhorizon.Aggregate
	EntityFactory    func() eventhorizon.Entity

	Commands []eventhorizon.CommandType
	Events   []eventhorizon.EventType
}

func NewAggregateEngine(middelware *Middleware,
	aggregateType eventhorizon.AggregateType,
	aggregateFactory func(id uuid.UUID) eventhorizon.Aggregate,
	entityFactory func() eventhorizon.Entity,
	commands []enum.Literal, events []enum.Literal) (ret *AggregateEngine) {
	ctx, cancel := context.WithCancel(context.Background())

	ret = &AggregateEngine{
		Middleware:       middelware,
		ctx:              ctx,
		cancel:           cancel,
		AggregateType:    aggregateType,
		AggregateFactory: aggregateFactory,
		EntityFactory:    entityFactory,
		Commands:         ConvertToCommandTypes(commands),
		Events:           ConvertToEventTypes(events),
	}
	return
}

func (o *AggregateEngine) Setup() (err error) {
	//register aggregate factory
	eventhorizon.RegisterAggregate(o.AggregateFactory)

	if err = o.registerCommands(); err != nil {
		return
	}
	return
}

func (o *AggregateEngine) registerCommands() (err error) {

	var aggregateStore eventhorizon.AggregateStore
	if aggregateStore, err = events.NewAggregateStore(o.EventStore, o.EventBus); err != nil {
		return
	}

	var commandHandler *aggregate.CommandHandler
	if commandHandler, err = aggregate.NewCommandHandler(o.AggregateType, aggregateStore); err != nil {
		return
	}

	for _, commandType := range o.Commands {
		if err = o.CommandBus.SetHandler(commandHandler, commandType); err != nil {
			return
		}
	}
	return
}

func (o *AggregateEngine) RegisterProjector(projectorType string, listener DelegateEventHandler) (ret *ProjectorEventHandler, err error) {
	var repo eventhorizon.ReadWriteRepo
	if repo, err = o.Repos(projectorType, o.EntityFactory); err != nil {
		return
	}

	ret = NewProjector(projectorType, listener, repo)
	proj := projector.NewEventHandler(ret, repo)
	proj.SetEntityFactory(o.EntityFactory)
	err = o.RegisterForEvents(proj, listener.EventTypes())
	return
}

func (o *AggregateEngine) RegisterForEventsAll(handler eventhorizon.EventHandler) (err error) {
	err = o.RegisterForEvents(handler, o.Events)
	return
}

func (o *AggregateEngine) RegisterForEvents(handler eventhorizon.EventHandler, events []eventhorizon.EventType) (err error) {
	err = o.EventBus.AddHandler(o.ctx, eventhorizon.MatchEvents(events), handler)
	return
}

func (o *AggregateEngine) RegisterSagaForEvents(aSaga saga.Saga, events []eventhorizon.EventType) (err error) {
	responseSaga := saga.NewEventHandler(aSaga, o.CommandBus)
	err = o.RegisterForEvents(responseSaga, events)
	return
}

func (o *AggregateEngine) RegisterForEvent(handler eventhorizon.EventHandler, event enum.Literal) (err error) {
	err = o.EventBus.AddHandler(o.ctx, eventhorizon.MatchEvents([]eventhorizon.EventType{eventhorizon.EventType(event.Name())}), handler)
	return
}

func ConvertToCommandTypes(commands []enum.Literal) []eventhorizon.CommandType {
	i := 0
	var commandTypes = make([]eventhorizon.CommandType, len(commands))
	for _, command := range commands {
		commandTypes[i] = eventhorizon.CommandType(command.Name())
		i += 1
	}
	return commandTypes
}

func ConvertToEventTypes(events []enum.Literal) []eventhorizon.EventType {
	i := 0
	var eventTypes = make([]eventhorizon.EventType, len(events))
	for _, command := range events {
		eventTypes[i] = eventhorizon.EventType(command.Name())
		i += 1
	}
	return eventTypes
}

type AggregateStoreEvent interface {
	AppendEvent(eventhorizon.EventType, eventhorizon.EventData, time.Time, ...eventhorizon.EventOption) eventhorizon.Event
}

type DelegateCommandHandler interface {
	CommandTypes() []eventhorizon.CommandType
	Execute(cmd eventhorizon.Command, entity eventhorizon.Entity, store AggregateStoreEvent) error
}

type DelegateEventHandler interface {
	EventTypes() []eventhorizon.EventType
	Apply(event eventhorizon.Event, entity eventhorizon.Entity) error
}

type Entity interface {
	EntityID() uuid.UUID
	Deleted() *time.Time
}

func CommandHandlerNotImplemented(commandType eventhorizon.CommandType) error {
	return errors.New(fmt.Sprintf("handler not implemented for %v", commandType))
}

func EventHandlerNotImplemented(eventType eventhorizon.EventType) error {
	return errors.New(fmt.Sprintf("handler not implemented for %v", eventType))
}

func EntityAlreadyExists(entityId uuid.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("entity already exists with id=%v and AggregateType=%v", entityId, aggregateType))
}

func EntityNotExists(entityId uuid.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("entity not exists with id=%v and AggregateType=%v", entityId, aggregateType))
}

func EntityChildNotExists(entityId uuid.UUID, aggregateType eventhorizon.AggregateType, childId uuid.UUID, childType string) error {
	return errors.New(fmt.Sprintf("%v(%v) not exists, %v(%v)",
		childType, childId, aggregateType, entityId))
}

func EntityChildIdNotDefined(entityId uuid.UUID, aggregateType eventhorizon.AggregateType, childType string) error {
	return errors.New(fmt.Sprintf("id for '%v' not defined, %v(%v)",
		childType, aggregateType, entityId))
}

func IdNotDefined(currentId uuid.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("id not defined for AggregateType=%v", aggregateType))
}

func IdsDismatch(entityId uuid.UUID, currentId uuid.UUID, aggregateType eventhorizon.AggregateType) error {
	return errors.New(fmt.Sprintf("Dismatch entity id and current id, %v != %v, for AggregateType=%v",
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
		net.ResponseResultErr(err, fmt.Sprintf("can't decode body to command %T", command),
			command, http.StatusBadRequest, w)
		return
	}

	path := html.EscapeString(r.URL.Path)
	if err = o.CommandBus.HandleCommand(o.Context, command); err != nil {
		net.ResponseResultErr(err,
			fmt.Sprintf("failed, command %T, %v", command, path), command, http.StatusExpectationFailed, w)
	} else {
		net.ResponseResultOk(
			fmt.Sprintf("succefully, command %T, %v", command, path), command, w)
	}
}

type ProjectorEventHandler struct {
	DelegateEventHandler
	Repo          eventhorizon.ReadRepo
	projectorType projector.Type
}

func NewProjector(projectorType string, eventHandler DelegateEventHandler, repo eventhorizon.ReadRepo) (ret *ProjectorEventHandler) {
	ret = &ProjectorEventHandler{
		projectorType:        projector.Type(projectorType),
		Repo:                 repo,
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
