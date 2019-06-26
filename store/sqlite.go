package store

import (
	"context"
	"database/sql"
	"github.com/edwardsb/secureworks/model"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" //import sqlite3 dialect
	"log"
)

var schema = `create table if not exists events
(
	id INTEGER
		constraint events_pk
			primary key autoincrement,
	username text,
	timestamp int,
	lat real,
	lon real,
	radius int,
	ip text,
	anonymous boolean,
	constraint events_pk_2
		unique (username, ip)
);`


const insert = `INSERT INTO events(username, timestamp, lat, lon, radius, ip, anonymous)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(username,ip) DO UPDATE SET timestamp=excluded.timestamp
                                       WHERE excluded.timestamp > events.timestamp;`
const subsequent = `SELECT *
FROM events
WHERE username = ? AND timestamp > ?
LIMIT 1;`

const preceeding = `SELECT *
FROM events
WHERE username = ? AND timestamp < ?
LIMIT 1;`

//SqliteStorer satisfies the Storer interface, but is specific to Sqlite
type SqliteStorer struct {
	db *sqlx.DB
}

//NewSqliteDb is a constructor that takes a *sql.DB, this is so you can modify the driver before it gets here.
//For example you could decorate the driver with an OpenTracing db driver.
func NewSqliteDb(db *sql.DB) *SqliteStorer {
	return &SqliteStorer{db: sqlx.NewDb(db, "sqlite3")}
}

//Open is a very cheap way to run migrations, normally you'd just ping the database
//maybe even have an exponential backoff, in case it takes a little bit to establish connection
func (s *SqliteStorer) Open() error {
	log.Println("opening sqlite store")
	_, err := s.db.Exec(schema)
	return err
}

//Close closes the underlying db
func (s *SqliteStorer) Close() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	log.Println("closing sqlite store")
	return nil
}

//Put will store the user login event into the database, it will also update the record model with the ID that was inserted
func (s *SqliteStorer) Put(ctx context.Context, record *model.Record) (int64, error) {
	result, err := s.db.ExecContext(ctx, insert, record.UserName, record.Timestamp, record.Lat, record.Lon, record.Radius, record.IP, record.Anonymous)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

//PrecedingAccess gets the access that happened before timestamp for the specified user
func (s *SqliteStorer) PrecedingAccess(ctx context.Context, user string, timestamp int64) (*model.Record, error) {
	return s.getAccess(ctx, user, timestamp, preceeding)
}

//SubsequentAccess gets the access that happened after timestamp for the specified user
func (s *SqliteStorer) SubsequentAccess(ctx context.Context, user string, timestamp int64) (*model.Record, error) {
	return s.getAccess(ctx, user, timestamp, subsequent)
}

func (s *SqliteStorer) getAccess(ctx context.Context, user string, timestamp int64, query string) (*model.Record, error) {
	record := &model.Record{}
	err := s.db.GetContext(ctx, record, query, user, timestamp)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return record, nil
}
