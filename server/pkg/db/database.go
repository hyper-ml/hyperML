package db

import (
	"database/sql"
	"fmt"
	"github.com/dgraph-io/badger"
	// Postgres may not be used
	_ "github.com/lib/pq"
	"time"

	"github.com/hyper-ml/hyperml/server/pkg/config"
)

// Body :
type Body map[string]interface{}

// DatabaseContext :
type DatabaseContext interface {
	GetSequence(kind string, cache uint64) (uint64, error)
	Get(key string) ([]byte, error)
	Insert(key string, value interface{}) error
	InsertAndTrack(key string, value interface{}, templates ...interface{}) error
	Update(key string, value interface{}) error
	UpdateAndTrack(key string, value interface{}, templates ...interface{}) error
	Upsert(key string, value interface{}) error

	List(prefix []byte) ([][]byte, error)
	ListKeys(prefix []byte) ([][]byte, error)

	DeleteIfExists(key string) error
	Delete(key string) error
	SoftDelete(key string) error
	KeyExists(key string) bool
	GetListener() ChangeListener
	Close()
}

// NewDatabaseContext :
func NewDatabaseContext(c *config.DBConfig) (DatabaseContext, error) {

	switch c.Driver {
	case config.Postgres:
		return NewPGContext(c.Name, c.User, c.Pass)
	case config.Badger:
		return NewBadgerContext(c.Name, c.DataDirPath, c.EventBuffer)
	}

	return nil, fmt.Errorf("Failed to initiate database context")
}

// BadgerContext :
type BadgerContext struct {
	Name     string
	db       *badger.DB
	LastCall time.Time
	listener ChangeListener
	seqList  map[string]*badger.Sequence
	// add listener active flag to avoid multiple listener
}

// NewBadgerContext : Creates a new badger conn
func NewBadgerContext(name, datapath string, eventBuf int) (DatabaseContext, error) {
	opts := badger.DefaultOptions
	opts.Dir = datapath
	opts.ValueDir = datapath
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	seqList := make(map[string]*badger.Sequence)

	return &BadgerContext{
		Name:     name,
		db:       db,
		LastCall: time.Now(),
		listener: NewChangeListener(eventBuf),
		seqList:  seqList,
	}, nil

}

// PGContext :
type PGContext struct {
	Name     string
	conn     *sql.DB
	LastCall time.Time
	Listener ChangeListener
	// add listener active flag to avoid multiple listener
}

// NewPGContext :
func NewPGContext(name, user, pass string) (DatabaseContext, error) {
	// var err error
	// var conn_str string

	// if pass != "" {
	// 	conn_str = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
	// 		user, pass, name)
	// } else {
	// 	conn_str = fmt.Sprintf("user=%s dbname=%s sslmode=disable",
	// 		user, name)
	// }

	// db, err := sql.Open("postgres", conn_str)
	// if err != nil {
	//	return nil, err
	//}

	// TODO: add config for listener
	// change_lnr := NewChangeListener(25)

	//return &PGContext{Name: name, conn: db, LastCall: time.Now(), Listener: change_lnr}, nil
	return nil, fmt.Errorf("Unimplemented")
}
