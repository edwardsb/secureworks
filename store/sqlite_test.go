package store

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/edwardsb/secureworks/model"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSqliteStorer_Put(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	store := SqliteStorer{db: sqlx.NewDb(db, "sqlite3")}

	timestamp := time.Now().Unix()
	mock.ExpectExec("INSERT INTO events").
		WithArgs("foo", timestamp, 27.950575, -82.457176, 50, "10.24.1.22", false).
		WillReturnResult(sqlmock.NewResult(int64(12), 1))

	record := &model.Record{
		UserName:  "foo",
		Timestamp: timestamp,
		IP:        "10.24.1.22",
		Anonymous: false,
		Geo: model.Geo{
			Lat:    27.950575,
			Lon:    -82.457176,
			Radius: 50,
		},
	}
	id, err := store.Put(context.Background(), record)
	require.NoError(t, err)
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
	require.Equal(t, id, int64(12))
}
