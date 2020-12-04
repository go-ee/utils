package eh

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
	Err error
	Cmd eh.Command
	Entity eh.Entity
}

func (o CommandError) Error() string {
	return fmt.Sprintf("%v.%v: %v", o.Entity.EntityID(), o.Cmd.CommandType(), o.Err)
}

func (o CommandError) Cause() error {
	return o.Err
}

func NewErrCouldNotMarshalEvent(ctx context.Context, err error) error {
	return eh.EventStoreError{
		Err:       ErrCouldNotMarshalEvent,
		BaseErr:   err,
		Namespace: eh.NamespaceFromContext(ctx),
	}
}

func NewErrCouldNotUnmarshalEvent(ctx context.Context, err error) error {
	return eh.EventStoreError{
		Err:       ErrCouldNotUnmarshalEvent,
		BaseErr:   err,
		Namespace: eh.NamespaceFromContext(ctx),
	}
}

func NewErrCouldNotLoadAggregate(ctx context.Context, err error) error {
	return eh.EventStoreError{Err: ErrCouldNotLoadAggregate,
		BaseErr:   err,
		Namespace: eh.NamespaceFromContext(ctx),
	}
}

func NewErrIncorrectEventVersion(ctx context.Context) error {
	return eh.EventStoreError{
		Err:       eh.ErrIncorrectEventVersion,
		Namespace: eh.NamespaceFromContext(ctx),
	}
}

func NewErrCouldNotSaveAggregate(ctx context.Context, err error) error {
	return eh.EventStoreError{
		Err:       ErrCouldNotSaveAggregate,
		BaseErr:   err,
		Namespace: eh.NamespaceFromContext(ctx),
	}
}
