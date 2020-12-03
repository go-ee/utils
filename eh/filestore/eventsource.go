package filestore

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	os "os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	eh "github.com/looplab/eventhorizon"
)

// ErrCouldNotClearDB is when the database could not be cleared.
var ErrCouldNotClearDB = errors.New("could not clear database")

// ErrCouldNotMarshalEvent is when an event could not be marshaled into BSON.
var ErrCouldNotMarshalEvent = errors.New("could not marshal event")

// ErrCouldNotUnmarshalEvent is when an event could not be unmarshaled into a concrete type.
var ErrCouldNotUnmarshalEvent = errors.New("could not unmarshal event")

// ErrCouldNotLoadAggregate is when an aggregate could not be loaded.
var ErrCouldNotLoadAggregate = errors.New("could not load aggregate")

// ErrCouldNotSaveAggregate is when an aggregate could not be saved.
var ErrCouldNotSaveAggregate = errors.New("could not save aggregate")

// EventStore implements an EventStore for MongoDB.
type EventStore struct {
	folder            string
	defaultFolderPerm os.FileMode
	defaultFilePerm   os.FileMode
}

func NewEventStore(folder string) *EventStore {
	return &EventStore{
		folder:            folder,
		defaultFolderPerm: 0777,
		defaultFilePerm:   0644,
	}
}

// Save implements the Save method of the eventhorizon.EventStore interface.
func (s *EventStore) Save(ctx context.Context, events []eh.Event, originalVersion int) error {
	if len(events) == 0 {
		return eh.EventStoreError{
			Err:       eh.ErrNoEventsToAppend,
			Namespace: eh.NamespaceFromContext(ctx),
		}
	}

	// Build all event records, with incrementing versions starting from the
	// original aggregate version.
	dbEvents := make([]dbEvent, len(events))
	firstEvent := events[0]
	aggregateId := firstEvent.AggregateID()
	aggregateType := firstEvent.AggregateType()
	for i, event := range events {
		// Only accept events belonging to the same aggregate.
		if event.AggregateID() != aggregateId {
			return eh.EventStoreError{
				Err:       eh.ErrInvalidEvent,
				Namespace: eh.NamespaceFromContext(ctx),
			}
		}

		// Only accept events that apply to the correct aggregate version.
		if event.Version() != originalVersion+i+1 {
			return eh.EventStoreError{
				Err:       eh.ErrIncorrectEventVersion,
				Namespace: eh.NamespaceFromContext(ctx),
			}
		}

		// Create the event record for the DB.
		e, err := newDBEvent(ctx, event)
		if err != nil {
			return err
		}
		dbEvents[i] = *e
	}

	namespaceFolder := s.buildFolderName(ctx)
	if err := os.MkdirAll(namespaceFolder, s.defaultFolderPerm); err != nil {
		return eh.EventStoreError{
			Err:       ErrCouldNotSaveAggregate,
			BaseErr:   err,
			Namespace: eh.NamespaceFromContext(ctx),
		}
	}

	// Either insert a new aggregate or append to an existing.
	if originalVersion == 0 {
		if err := s.saveAggregate(ctx, namespaceFolder, &aggregateRecord{
			Id:      aggregateId,
			Version: len(dbEvents),
			Type:    aggregateType,
		}); err != nil {
			return err
		}

		if err := s.writeEvents(ctx, namespaceFolder, aggregateId, dbEvents); err != nil {
			return err
		}
	} else {
		// Increment aggregate version on insert of new event record, and
		// only insert if version of aggregate is matching (ie not changed
		// since loading the aggregate).
		if aggregate, err := s.loadAggregate(ctx, namespaceFolder, aggregateId); err != nil {
			return err
		} else {
			if aggregate.Version != originalVersion {
				return eh.EventStoreError{
					Err:       ErrCouldNotSaveAggregate,
					BaseErr:   fmt.Errorf("invalid original version %d", originalVersion),
					Namespace: eh.NamespaceFromContext(ctx),
				}
			} else {
				aggregate.Version += len(dbEvents)
				if err := s.saveAggregate(ctx, namespaceFolder, aggregate); err != nil {
					return err
				}

				if err := s.writeEvents(ctx, namespaceFolder, aggregateId, dbEvents); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *EventStore) writeEvents(
	ctx context.Context, namespaceFolder string, aggregateId uuid.UUID, dbEvents []dbEvent) error {

	aggregateEventsFileName := s.buildAggregateEventsFileName(namespaceFolder, aggregateId)
	if aggregateEventsFile, err :=
		os.OpenFile(aggregateEventsFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, s.defaultFilePerm); err != nil {
		return s.newCouldNotSaveAggregate(ctx, err)
	} else {
		aggregateEventsWriter := bufio.NewWriter(aggregateEventsFile)

		for _, dbEvent := range dbEvents {
			if err := s.writeEvent(ctx, dbEvent, aggregateEventsWriter); err != nil {
				return err
			}
		}
		aggregateEventsWriter.Flush()
		aggregateEventsFile.Close()
	}
	return nil
}

func (s *EventStore) writeEvent(ctx context.Context, dbEvent dbEvent, aggregateEventsWriter *bufio.Writer) error {
	if bytes, err := json.Marshal(dbEvent); err == nil {
		if _, err = aggregateEventsWriter.Write(bytes); err != nil {
			return s.newCouldNotSaveAggregate(ctx, err)
		} else {
			if _, err = aggregateEventsWriter.WriteString("\n"); err != nil {
				return s.newCouldNotSaveAggregate(ctx, err)
			}
		}
	} else {
		return s.newCouldNotSaveAggregate(ctx, err)
	}
	return nil
}

func (s *EventStore) newCouldNotSaveAggregate(ctx context.Context, baseErr error) error {
	return eh.EventStoreError{
		Err:       ErrCouldNotSaveAggregate,
		BaseErr:   baseErr,
		Namespace: eh.NamespaceFromContext(ctx),
	}
}

// Load implements the Load method of the eventhorizon.EventStore interface.
func (s *EventStore) Load(ctx context.Context, id uuid.UUID) (ret []eh.Event, err error) {
	namespaceFolder := s.buildFolderName(ctx)

	var aggregate *aggregateRecord
	if aggregate, err = s.loadAggregate(ctx, namespaceFolder, id); err != nil {
		return
	}

	if aggregate == nil {
		return []eh.Event{}, nil
	}

	eventsFileName := s.buildAggregateEventsFileName(namespaceFolder, id)
	var eventsFile *os.File
	if eventsFile, err = os.Open(eventsFileName); err != nil {
		if os.IsNotExist(err) {
			return []eh.Event{}, nil
		} else {
			return nil, eh.EventStoreError{
				Err:       ErrCouldNotLoadAggregate,
				BaseErr:   err,
				Namespace: eh.NamespaceFromContext(ctx),
			}
		}
	}
	defer eventsFile.Close()

	events := make([]eh.Event, aggregate.Version)
	eventIndex := 0
	scanner := bufio.NewScanner(eventsFile)
	for scanner.Scan() {
		dbEventType := dbEventType{}
		if err = json.Unmarshal(scanner.Bytes(), &dbEventType); err != nil {
			return s.errCouldNotUnmarshalEvent(ctx, err)
		}
		var eventData eh.EventData
		if eventData, err = eh.CreateEventData(dbEventType.EventType); err != nil {
			return s.errCouldNotUnmarshalEvent(ctx, err)
		}

		dbEvent := dbEvent{
			Data: eventData,
		}
		if err = json.Unmarshal(scanner.Bytes(), &dbEvent); err == nil {
			events[eventIndex] = event{
				aggregate: aggregate,
				dbEvent:   dbEvent,
			}
			eventIndex += 1
		} else {
			return nil, eh.EventStoreError{
				Err:       ErrCouldNotUnmarshalEvent,
				BaseErr:   err,
				Namespace: eh.NamespaceFromContext(ctx),
			}
		}

	}
	return events, nil
}

func (s *EventStore) errCouldNotUnmarshalEvent(ctx context.Context, err error) ([]eh.Event, error) {
	return nil, eh.EventStoreError{
		Err:       ErrCouldNotUnmarshalEvent,
		BaseErr:   err,
		Namespace: eh.NamespaceFromContext(ctx),
	}
}

func (s *EventStore) loadAggregate(
	ctx context.Context, namespaceFolder string, id uuid.UUID) (ret *aggregateRecord, err error) {

	aggregateFileName := s.buildAggregateFileName(namespaceFolder, id)

	if data, err := ioutil.ReadFile(aggregateFileName); err == nil {
		ret = &aggregateRecord{}
		if err = json.Unmarshal(data, ret); err != nil {
			err = eh.EventStoreError{
				Err:       ErrCouldNotLoadAggregate,
				BaseErr:   err,
				Namespace: eh.NamespaceFromContext(ctx),
			}
		}
	}
	return
}

func (s *EventStore) saveAggregate(
	ctx context.Context, namespaceFolder string, aggregate *aggregateRecord) (err error) {

	var data []byte
	if data, err = json.MarshalIndent(aggregate, "", " "); err == nil {
		aggregateFileName := s.buildAggregateFileName(namespaceFolder, aggregate.Id)
		err = ioutil.WriteFile(aggregateFileName, data, s.defaultFilePerm)
	}

	if err != nil {
		err = eh.EventStoreError{
			Err:       ErrCouldNotSaveAggregate,
			BaseErr:   err,
			Namespace: eh.NamespaceFromContext(ctx),
		}
	}
	return
}

// Replace implements the Replace method of the eventhorizon.EventStore interface.
func (s *EventStore) Replace(ctx context.Context, event eh.Event) error {
	return nil
}

// RenameEvent implements the RenameEvent method of the eventhorizon.EventStore interface.
func (s *EventStore) RenameEvent(ctx context.Context, from, to eh.EventType) error {
	return nil
}

// Clear clears the event storage.
func (s *EventStore) Clear(ctx context.Context) error {
	if err := os.RemoveAll(s.buildFolderName(ctx)); err != nil {
		return eh.EventStoreError{
			Err:       ErrCouldNotClearDB,
			BaseErr:   err,
			Namespace: eh.NamespaceFromContext(ctx),
		}
	}
	return nil
}

// Close closes the database client.
func (s *EventStore) Close(ctx context.Context) {
}

func (s *EventStore) buildFolderName(ctx context.Context) string {
	return filepath.Join(s.folder, eh.NamespaceFromContext(ctx))
}

func (s *EventStore) buildAggregateFileName(folder string, aggregateId uuid.UUID) string {
	return filepath.Join(folder, aggregateId.String()) + ".json"
}

func (s *EventStore) buildAggregateEventsFileName(folder string, aggregateId uuid.UUID) string {
	return filepath.Join(folder, aggregateId.String()) + "_events.json"
}

// aggregateRecord is the Database representation of an aggregate.
type aggregateRecord struct {
	Id      uuid.UUID        `json:"id"`
	Version int              `json:"version"`
	Type    eh.AggregateType `json:"type"`
}

// dbEvent is the internal event record for the MongoDB event store used
// to save and load events from the DB.

type dbEventType struct {
	EventType eh.EventType `json:"event_type"`
}

type dbEvent struct {
	EventType eh.EventType `json:"event_type"`
	Timestamp time.Time    `json:"timestamp"`
	Version   int          `json:"version"`
	Data      interface{}  `json:"data,omitempty"`
}

// newDBEvent returns a new dbEvent for an event.
func newDBEvent(ctx context.Context, event eh.Event) (ret *dbEvent, err error) {
	ret = &dbEvent{
		EventType: event.EventType(),
		Timestamp: event.Timestamp(),
		Version:   event.Version(),
		Data:      event.Data(),
	}
	return
}

// event is the private implementation of the eventhorizon.Event interface
// for a MongoDB event store.
type event struct {
	aggregate *aggregateRecord
	dbEvent
}

// AggrgateID implements the AggrgateID method of the eventhorizon.Event interface.
func (e event) AggregateID() uuid.UUID {
	return e.aggregate.Id
}

// AggregateType implements the AggregateType method of the eventhorizon.Event interface.
func (e event) AggregateType() eh.AggregateType {
	return e.aggregate.Type
}

// EventType implements the EventType method of the eventhorizon.Event interface.
func (e event) EventType() eh.EventType {
	return e.dbEvent.EventType
}

// Data implements the Data method of the eventhorizon.Event interface.
func (e event) Data() eh.EventData {
	return e.dbEvent.Data
}

// Version implements the Version method of the eventhorizon.Event interface.
func (e event) Version() int {
	return e.dbEvent.Version
}

// Timestamp implements the Timestamp method of the eventhorizon.Event interface.
func (e event) Timestamp() time.Time {
	return e.dbEvent.Timestamp
}

// String implements the String method of the eventhorizon.Event interface.
func (e event) String() string {
	return fmt.Sprintf("%s@%d", e.dbEvent.EventType, e.dbEvent.Version)
}
