package store

import (
	"context"
	"github.com/edwardsb/secureworks/model"
)

//Storer is responsible for writing new events to the data store, and retrieving preceding and subsequent access
type Storer interface {
	Put(ctx context.Context, record *model.Record) (int64, error)
	PrecedingAccess(ctx context.Context, user string, timestamp int64) (*model.Record, error)
	SubsequentAccess(ctx context.Context, user string, timestamp int64) (*model.Record, error)
}
