package store

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/edwardsb/secureworks/model"
	"github.com/guregu/dynamo"
)

type DynamoStorer struct {
	db *dynamo.DB
}

func NewDynamoStore() *DynamoStorer {
	config := &aws.Config{Endpoint: aws.String("localhost:8000"), Region: aws.String("local-test")}
	sess, _ := session.NewSession(config)
	return &DynamoStorer{db: dynamo.New(sess, config)}
}

func (s *DynamoStorer) Put(ctx context.Context, record *model.Record) error {
	return s.db.Table("events").Put(record).RunWithContext(ctx)
}


func (s *DynamoStorer) PrecedingAccess(ctx context.Context, user string, timestamp int64) (*model.Record, error) {
	return s.getEvent(ctx, user, timestamp, dynamo.Less, dynamo.Descending)
}

func (s *DynamoStorer) SubsequentAccess(ctx context.Context, user string, timestamp int64) (*model.Record, error) {
	return s.getEvent(ctx, user, timestamp, dynamo.Greater, dynamo.Ascending)
}

func (s *DynamoStorer) getEvent(ctx context.Context, user string, ts int64, op dynamo.Operator, order dynamo.Order) (*model.Record, error){
	var result model.Record
	var err error
	err = s.db.Table("events").
		Get("username", user).
		Range("ts", op, ts).
		Order(order).Limit(1).OneWithContext(ctx, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
