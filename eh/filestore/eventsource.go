package filestore

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	esh "github.com/go-ee/utils/eh"
	"github.com/go-ee/utils/eio"
	"github.com/google/uuid"
	eh "github.com/looplab/eventhorizon"
	"io"
	"os"
	"path/filepath"
	"time"
)

type EventStore struct {
	*Base
}

func NewEventStore(folder string) *EventStore {
	return &EventStore{
		Base: NewBase(folder),
	}
}

// Save implements the Save method of the eventhorizon.EventStore interface.
func (s *EventStore) Save(ctx context.Context, events []eh.Event, originalVersion int) (err error) {
	if len(events) == 0 {
		return &eh.EventStoreError{
			Err: fmt.Errorf("no events to append for '%v'", esh.ContextGetNamespace(ctx)),
		}
	}

	namespaceFolder := s.buildFolderName(ctx)
	if err = os.MkdirAll(namespaceFolder, s.defaultFolderPerm); err != nil {
		return esh.NewErrCouldNotSaveAggregate(ctx, err)
	}

	firstEvent := events[0]
	aggregateId := firstEvent.AggregateID()

	aggregateEventsFileName := buildEventsFileName(namespaceFolder, aggregateId)
	if err = checkAggregateVersion(ctx, aggregateEventsFileName, originalVersion); err != nil {
		return
	}

	var aggregateEventsFile *os.File
	if aggregateEventsFile, err =
		os.OpenFile(aggregateEventsFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, s.defaultFilePerm); err != nil {
		return esh.NewErrCouldNotSaveAggregate(ctx, err)
	}
	defer aggregateEventsFile.Close()

	// Build all dbEvent records, with incrementing versions starting from the
	// original aggregate version.
	dbEvents := make([]dbEvent, len(events))

	for i, event := range events {
		// Only accept events belonging to the same aggregate.
		if event.AggregateID() != aggregateId {
			return &eh.EventStoreError{
				Err: fmt.Errorf("invalid event for '%v'", esh.ContextGetNamespace(ctx)),
			}
		}

		// Only accept events that apply to the correct aggregate version.
		if event.Version() != originalVersion+i+1 {
			return esh.NewErrIncorrectEventVersion(ctx)
		}

		// Create the dbEvent record for the DB.
		e, err := newDbEvent(event)
		if err != nil {
			return err
		}
		dbEvents[i] = *e
	}

	aggregateEventsWriter := bufio.NewWriter(aggregateEventsFile)

	for _, dbEvent := range dbEvents {
		if err := writeEvent(ctx, dbEvent, aggregateEventsWriter); err != nil {
			return err
		}
	}
	aggregateEventsWriter.Flush()
	aggregateEventsFile.Close()

	return
}

func (s *EventStore) Load(ctx context.Context, id uuid.UUID) (ret []eh.Event, err error) {
	namespaceFolder := s.buildFolderName(ctx)

	eventsFileName := buildEventsFileName(namespaceFolder, id)
	var eventsFile *os.File
	if eventsFile, err = os.Open(eventsFileName); err != nil {
		if os.IsNotExist(err) {
			return []eh.Event{}, nil
		} else {
			return nil, esh.NewErrCouldNotLoadAggregate(ctx, err)
		}
	}
	defer eventsFile.Close()

	eventIndex := 0
	scanner, err := eio.NewReverseScannerFile(eventsFile)
	if err != nil {
		return nil, esh.NewErrCouldNotLoadAggregate(ctx, err)
	}
	// read last line
	if !scanner.Scan() {
		if scanner.ScanErr() == io.EOF {
			return s.noEvents()
		}
		return nil, esh.NewErrCouldNotLoadAggregate(ctx, err)
	}

	var lastEvent *dbEvent
	if lastEvent, err = parseEvent(ctx, scanner.Bytes()); err != nil {
		return
	}
	var events = make([]eh.Event, lastEvent.Version())
	eventIndex = lastEvent.Version() - 1
	events[eventIndex] = lastEvent

	for scanner.Scan() {
		eventIndex -= 1
		if events[eventIndex], err = parseEvent(ctx, scanner.Bytes()); err != nil {
			return
		}
	}
	return events, nil
}

func (s *EventStore) noEvents() ([]eh.Event, error) {
	return []eh.Event{}, nil
}

/*
func (s *EventStore) Replace(ctx context.Context, event eh.Event) error {
	return nil
}

func (s *EventStore) RenameEvent(ctx context.Context, from, to eh.EventType) error {
	return nil
}
*/

func (s *EventStore) Clear(ctx context.Context) error {
	if err := os.RemoveAll(s.buildFolderName(ctx)); err != nil {
		return &eh.EventStoreError{
			Err: fmt.Errorf("%v: %v", esh.ErrCouldNotClearDB, err),
		}
	}
	return nil
}

func (s *EventStore) buildFolderName(ctx context.Context) string {
	return filepath.Join(s.folder, esh.ContextGetNamespace(ctx))
}

func scanLastEvent(ctx context.Context, scanner *eio.ReverseScanner) (ret *dbEvent, err error) {
	if !scanner.Scan() {
		if scanner.ScanErr() != io.EOF {
			err = esh.NewErrCouldNotLoadAggregate(ctx, scanner.ScanErr())
		}
	} else {
		ret, err = parseEvent(ctx, scanner.Bytes())
	}
	return
}

func checkAggregateVersion(
	ctx context.Context, eventsFileName string, originalVersion int) (err error) {

	var lastEvent *dbEvent
	if lastEvent, err = loadLastEvent(ctx, eventsFileName); err != nil {
		return esh.NewErrCouldNotLoadAggregate(ctx, err)
	}
	var aggregateVersion int
	if lastEvent == nil {
		aggregateVersion = 0
	} else {
		aggregateVersion = lastEvent.Version()
	}

	// Increment aggregate version on insert of new dbEvent record, and
	// only insert if version of aggregate is matching (ie not changed
	// since loading the aggregate).
	if aggregateVersion != originalVersion {
		err = &eh.EventStoreError{
			Err: fmt.Errorf("%v, invalid original version %d", esh.ErrCouldNotSaveAggregate, originalVersion),
		}
	}
	return
}

func loadLastEvent(ctx context.Context, eventsFileName string) (ret *dbEvent, err error) {
	var aggregateEventsFile *os.File
	if aggregateEventsFile, err = os.Open(eventsFileName); err != nil {
		return nil, nil
	}
	defer aggregateEventsFile.Close()

	var scanner *eio.ReverseScanner
	if scanner, err = eio.NewReverseScannerFile(aggregateEventsFile); err != nil {
		err = esh.NewErrCouldNotLoadAggregate(ctx, err)
		return
	}

	ret, err = scanLastEvent(ctx, scanner)
	return
}

func writeEvent(ctx context.Context, dbEvent dbEvent, aggregateEventsWriter *bufio.Writer) error {
	if bytes, err := json.Marshal(dbEvent); err == nil {
		if _, err = aggregateEventsWriter.Write(bytes); err != nil {
			return esh.NewErrCouldNotSaveAggregate(ctx, err)
		} else {
			if _, err = aggregateEventsWriter.WriteString("\n"); err != nil {
				return esh.NewErrCouldNotSaveAggregate(ctx, err)
			}
		}
	} else {
		return esh.NewErrCouldNotMarshalEvent(ctx, err)
	}
	return nil
}

func parseEvent(ctx context.Context, data []byte) (ret *dbEvent, err error) {
	dbEventType := dbEventType{}
	if err = json.Unmarshal(data, &dbEventType); err != nil {
		err = esh.NewErrCouldNotUnmarshalEvent(ctx, err)
		return
	}

	dbEvent := dbEvent{}

	if eventData, eventDataErr := eh.CreateEventData(dbEventType.EventType); eventDataErr == nil {
		dbEvent.Data_ = eventData
	}

	if err = json.Unmarshal(data, &dbEvent); err == nil {
		ret = &dbEvent
	} else {
		err = esh.NewErrCouldNotUnmarshalEvent(ctx, err)
	}
	return
}

// dbEvent is the internal dbEvent record for the MongoDB dbEvent store used
// to save and load events from the DB.

type dbEventType struct {
	EventType eh.EventType `json:"event_type"`
}

type dbEvent struct {
	AggregateID_   uuid.UUID              `json:"aggregate_id"`
	AggregateType_ eh.AggregateType       `json:"aggregate_type"`
	EventType_     eh.EventType           `json:"event_type"`
	Timestamp_     time.Time              `json:"timestamp"`
	Version_       int                    `json:"version"`
	Data_          interface{}            `json:"data,omitempty"`
	Metadata_      map[string]interface{} `json:"metadata"`
}

func newDbEvent(ehEvent eh.Event) (ret *dbEvent, err error) {
	ret = &dbEvent{
		AggregateID_:   ehEvent.AggregateID(),
		AggregateType_: ehEvent.AggregateType(),
		EventType_:     ehEvent.EventType(),
		Timestamp_:     ehEvent.Timestamp(),
		Version_:       ehEvent.Version(),
		Data_:          ehEvent.Data(),
		Metadata_:      ehEvent.Metadata(),
	}
	return
}

func (e dbEvent) AggregateID() uuid.UUID {
	return e.AggregateID_
}

func (e dbEvent) AggregateType() eh.AggregateType {
	return e.AggregateType_
}

func (e dbEvent) EventType() eh.EventType {
	return e.EventType_
}

func (e dbEvent) Data() eh.EventData {
	return e.Data_
}

func (e dbEvent) Version() int {
	return e.Version_
}

func (e dbEvent) Timestamp() time.Time {
	return e.Timestamp_
}

func (e dbEvent) Metadata() map[string]interface{} {
	return e.Metadata_
}

func (e dbEvent) String() string {
	return fmt.Sprintf("%s@%d", e.EventType_, e.Version_)
}

func buildEventsFileName(folder string, id uuid.UUID) string {
	return filepath.Join(folder, id.String()) + ".json"
}
