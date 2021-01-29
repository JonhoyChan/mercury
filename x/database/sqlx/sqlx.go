package sqlx

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"mercury/config"
	"mercury/x/log"
	"reflect"
	"time"

	"github.com/pkg/errors"

	_ "github.com/lib/pq"
)

var (
	ErrNoRows = sql.ErrNoRows
)

type DB struct {
	*sql.DB
	rdbChan chan *sql.DB
}

func IsErrNoRows(err error) bool {
	return err == ErrNoRows
}

func (db *DB) Close() error {
	Close(db.DB)
	return nil
}

type ConfigProvider interface {
	Database() *config.Database
}

func Open(c ConfigProvider) (*DB, error) {
	dbc := c.Database()
	db, err := sql.Open(dbc.Driver, dbc.DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(dbc.Active)
	db.SetMaxIdleConns(dbc.Idle)
	db.SetConnMaxLifetime(dbc.IdleTimeout)

	rdbChan := make(chan *sql.DB, 1)
	rdbChan <- db

	return &DB{db, rdbChan}, nil
}

func Close(closer io.Closer) {
	if err := closer.Close(); err != nil {
		log.Warn("close error", "error", err)
	}
}

func Rollback(tx *Tx) {
	if err := tx.Rollback(); err != nil {
		log.Warn("rollback error", "error", err)
	}
}

func (db *DB) Exec(sql string, affect int64, args ...interface{}) error {
	rs, err := db.DB.Exec(sql, args...)
	if err != nil {
		return err
	}
	if affect > 0 {
		affected, err := rs.RowsAffected()
		if err != nil {
			return err
		}
		if affected != affect {
			return errors.New("invalid rows affected")
		}
	}
	return nil
}

func (db *DB) ExecX(sql string, affect int64, args ...interface{}) (int64, error) {
	rs, err := db.DB.Exec(sql, args...)
	if err != nil {
		return 0, err
	}
	if affect > 0 {
		affected, err := rs.RowsAffected()
		if err != nil {
			return 0, err
		}
		if affected != affect {
			return 0, errors.New("invalid rows affected")
		}
	}
	return rs.LastInsertId()
}

func (db *DB) QueryRow(sql string, args ...interface{}) *Row {
	rdb := <-db.rdbChan
	db.rdbChan <- rdb
	return &Row{rdb.QueryRow(sql, args...)}
}

func (db *DB) Query(sql string, args ...interface{}) (*Rows, error) {
	rdb := <-db.rdbChan
	db.rdbChan <- rdb
	rows, err := rdb.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows}, nil
}

func (db *DB) Begin() (*Tx, error) {
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx}, nil
}

type Row struct {
	*sql.Row
}

func (r *Row) Scan(dest ...interface{}) error {
	nullScanner := newNullScanner(dest...)
	if err := r.Row.Scan(nullScanner.nullType...); err != nil {
		if err == sql.ErrNoRows {
			return err
		}
		return err
	}
	return nullScanner.pop()
}

type Rows struct {
	*sql.Rows
}

func (r *Rows) Scan(dest ...interface{}) error {
	nullScanner := newNullScanner(dest...)
	if err := r.Rows.Scan(nullScanner.nullType...); err != nil {
		return err
	}
	return nullScanner.pop()
}

func (r *Rows) Err() error {
	if err := r.Rows.Err(); err != nil {
		return err
	}
	return nil
}

type Tx struct {
	*sql.Tx
}

func (tx *Tx) Commit() error {
	if err := tx.Tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (tx *Tx) Exec(sql string, affect int64, args ...interface{}) error {
	rs, err := tx.Tx.Exec(sql, args...)
	if err != nil {
		return err
	}
	if affect > 0 {
		affected, err := rs.RowsAffected()
		if err != nil {
			return err
		}
		if affected != affect {
			return errors.New("invalid rows affected")
		}
	}
	return nil
}

func (tx *Tx) ExecX(sql string, affect int64, args ...interface{}) (int64, error) {
	rs, err := tx.Tx.Exec(sql, args...)
	if err != nil {
		return 0, err
	}
	if affect > 0 {
		affected, err := rs.RowsAffected()
		if err != nil {
			return 0, err
		}
		if affected != affect {
			return 0, errors.New("invalid rows affected")
		}
	}
	return rs.LastInsertId()
}

func (tx *Tx) QueryRow(sql string, args ...interface{}) *Row {
	return &Row{tx.Tx.QueryRow(sql, args...)}
}

func (tx *Tx) Query(sql string, args ...interface{}) (*Rows, error) {
	rows, err := tx.Tx.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{rows}, nil
}

func (tx *Tx) Rollback() error {
	if err := tx.Tx.Rollback(); err != nil {
		return err
	}
	return nil
}

type nullScanner struct {
	dest     []interface{}
	nullType []interface{}
}

func newNullScanner(dest ...interface{}) *nullScanner {
	nullType := make([]interface{}, len(dest))
	for i, item := range dest {
		switch item.(type) {
		case *string:
			nullType[i] = &sql.NullString{}
		case *int64, *int32, *int16, *int8, *int,
			*uint64, *uint32, *uint16, *uint8, *uint:
			nullType[i] = &sql.NullInt64{}
		case *float64, *float32:
			nullType[i] = &sql.NullFloat64{}
		case *bool:
			nullType[i] = &sql.NullBool{}
		default:
			switch reflect.ValueOf(item).Elem().Kind() {
			case reflect.Int32:
				nullType[i] = &sql.NullInt64{}
			default:
				log.Error("unsupported type", "item", reflect.TypeOf(item))
			}
		}
	}
	return &nullScanner{dest: dest, nullType: nullType}
}

func (ns *nullScanner) pop() error {
	for i, item := range ns.nullType {
		val, err := item.(driver.Valuer).Value()
		if err != nil {
			return err
		}
		if val != nil {
			dv := reflect.Indirect(reflect.ValueOf(ns.dest[i]))
			switch d := ns.dest[i].(type) {
			case *string:
				*d = val.(string)
			case *bool:
				*d = val.(bool)
			case *int64, *int32, *int16, *int8, *int:
				dv.SetInt(val.(int64))
			case *uint64, *uint32, *uint16, *uint8, *uint:
				dv.SetUint(uint64(val.(int64)))
			case *float64, *float32:
				dv.SetFloat(val.(float64))
			case *time.Time:
				t := val.(time.Time)
				*d = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
			default:
				switch reflect.ValueOf(ns.dest[i]).Elem().Kind() {
				case reflect.Int32:
					dv.SetInt(val.(int64))
				default:
					return errors.New("unsupported dest type")
				}
			}
		}
	}
	return nil
}
