package db

import (
	"database/sql"
	"encoding/json"
	"github.com/hyper-ml/hyperml/server/pkg/base"
)

//TODO: Add entity checks

// Insert : DB Insert
func (p *PGContext) Insert(key string, value interface{}) error {
	tx, err := p.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO ws_collections(key, value) VALUES ($1, $2)")
	if err != nil {
		return err
	}

	//defer stmt.Close()
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if _, err = stmt.Exec(key, jsonValue); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Upsert :
func (p *PGContext) Upsert(key string, value interface{}) error {

	if p.KeyExists(key) {
		return p.Update(key, value)
	}

	return p.Insert(key, value)

}

// DeleteIfExists :
func (p *PGContext) DeleteIfExists(key string) error {
	tx, err := p.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("DELETE FROM ws_collections WHERE key=$1")
	if err != nil {
		return err
	}

	if _, err = stmt.Exec(key); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Delete : TODO: throw error if rec doesnt exist
func (p *PGContext) Delete(key string) error {
	tx, err := p.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("DELETE FROM ws_collections WHERE key=$1")
	if err != nil {
		return err
	}

	if _, err = stmt.Exec(key); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// SoftDelete :
func (p *PGContext) SoftDelete(key string) error {
	tx, err := p.conn.Begin()
	if err != nil {
		return err
	}
	trashKey := "deleted:" + key

	stmt, err := tx.Prepare("UPDATE ws_collections SET key = $1 WHERE key=$2")
	if err != nil {
		return err
	}

	if _, err = stmt.Exec(trashKey, key); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// KeyExists :
func (p *PGContext) KeyExists(key string) bool {
	var err error
	var data []byte

	rec, err := p.conn.Query("SELECT 1 FROM ws_collections WHERE key=$1", key)

	if err != nil {
		base.Log("Something wrong with database connection. Key Exists Error: %s", err)
		return false
	}

	defer rec.Close()

	for rec.Next() {
		err = rec.Scan(&data)
	}

	if data != nil {
		base.Info("Updating file info for key: ", key)
		return true
	}

	return false
}

// Get :
func (p *PGContext) Get(key string) ([]byte, error) {
	var data []byte
	var err error

	rows, err := p.conn.Query("SELECT value FROM ws_collections WHERE key=$1", key)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecNotFound{}
		}

		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&data)
	}

	if data == nil {
		base.Error("[PGContext.Get] No data found for key", key, data)
		return nil, ErrRecNotFound{}
	}

	return data, err
}

// Update :
// TODO: raises error if key is missing
func (p *PGContext) Update(key string, value interface{}) error {
	tx, err := p.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE ws_collections SET value=$1 WHERE key=$2")
	if err != nil {
		return err
	}

	//defer stmt.Close()
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if _, err = stmt.Exec(jsonValue, key); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// UpdateAndTrack : a bad hack to enable tracking of  updates
//
func (p *PGContext) UpdateAndTrack(key string, value interface{}, template interface{}) error {
	tx, err := p.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("UPDATE ws_collections SET value=$1 WHERE key=$2")
	if err != nil {
		return err
	}

	//defer stmt.Close()
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if _, err = stmt.Exec(jsonValue, key); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	if p.Listener != nil && template != nil {
		p.Listener.TrackEvent(value, template)
	}

	return nil
}

// GetListener :
func (p *PGContext) GetListener() ChangeListener {
	return p.Listener
}

// Close :
func (p *PGContext) Close() {

	if p.Listener != nil {
		p.Listener.Close()
	}
	defer p.conn.Close()
}
