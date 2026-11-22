package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// PortPool is a hosting daemon's member allocation pool, inclusive. Changing it
// never reassigns existing ports. An unset pool fails closed until configured.
type PortPool struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

func (p PortPool) Validate() error {
	if p.Start < 1024 || p.End > 65535 || p.End < p.Start {
		return fmt.Errorf("port pool must be an inclusive range within 1024..65535")
	}
	return nil
}

type portSettingReader interface{ QueryRow(string, ...any) *sql.Row }

func readPortPool(q portSettingReader) (PortPool, error) {
	var raw string
	p := PortPool{}
	err := q.QueryRow(`SELECT value FROM panel_settings WHERE key='port_pool_json'`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return p, fmt.Errorf("an administrator must configure the random port pool first")
	}
	if err != nil {
		return p, err
	}
	if err = json.Unmarshal([]byte(raw), &p); err != nil {
		return p, err
	}
	return p, p.Validate()
}
func (db *DB) GetPortPool() (PortPool, error) { return readPortPool(db.conn) }
func (db *DB) SetPortPool(p PortPool) error {
	if err := p.Validate(); err != nil {
		return err
	}
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return db.SetPanelSetting("port_pool_json", string(raw))
}
