package ehu

import (
	"context"
	"errors"
	"fmt"
	eh "github.com/looplab/eventhorizon"
)

var ErrAggregateDeleted = errors.New("the aggregate has already been deleted")

var ErrCouldNotClearDB = errors.New("could not clear database")

var ErrCouldNotMarshalEvent = errors.New("could not marshal dbEvent")

var ErrCouldNotUnmarshalEvent = errors.New("could not unmarshal dbEvent")

var ErrCouldNotLoadAggregate = errors.New("could not load aggregate")

var ErrCouldNotSaveAggregate = errors.New("could not save aggregate")

type CommandError struct {
	Err    error
	Cmd    eh.Command
	Entity eh.Entity
}

func (o CommandError) Error() string {
	return fmt.Sprintf("%v.%v: %v", o.Entity.EntityID(), o.Cmd.CommandType(), o.Err)
}

func (o CommandError) Cause() error {
	return o.Err
}

func NewErrCouldNotMarshalEvent(_ context.Context, err error) error {
	return &eh.EventStoreError{
		Err: fmt.Errorf("%v: %v", ErrCouldNotMarshalEvent, err),
	}
}

func NewErrCouldNotUnmarshalEvent(_ context.Context, err error) error {
	return &eh.EventStoreError{
		Err: fmt.Errorf("%v: %v", ErrCouldNotUnmarshalEvent, err),
	}
}

func NewErrCouldNotLoadAggregate(_ context.Context, err error) error {
	return &eh.EventStoreError{
		Err: fmt.Errorf("%v: %v", ErrCouldNotLoadAggregate, err),
	}
}

func NewErrIncorrectEventVersion(_ context.Context) error {
	return &eh.EventStoreError{
		Err: eh.ErrIncorrectEventVersion,
	}
}

func NewErrCouldNotSaveAggregate(_ context.Context, err error) error {
	return &eh.EventStoreError{
		Err: fmt.Errorf("%v: %v", ErrCouldNotSaveAggregate, err),
	}
}
