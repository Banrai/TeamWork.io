// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package database

import (
	"database/sql"
	"time"
)

const (
	// public key a/u/d
	PK_INSERT = "insert into public_key (person_id, key, nickname, source) values ($1, $2, $3, $4) returning id"
	PK_UPDATE = "update public_key set key = $1, nickname = $2, source = $3 where id = $4"
	PK_DELETE = "delete from public_key where id = $1"

	// public key lookup
	PK_LOOKUP = "select id, key, date_added, nickname, source from public_key where person_id = $1"
)

type PUBLIC_KEY struct {
	Id       string    `json:"id,omitempty"`
	Key      string    `json:"key"`
	Added    time.Time `json:"date_added,omitempty"`
	Nickname string    `json:"name,omitempty"`
	Source   string    `json:"source,omitempty"`
}

func (pk *PUBLIC_KEY) Add(stmt *sql.Stmt, personId string) (string, error) {
	var id sql.NullString
	err := stmt.QueryRow(personId, pk.Key, pk.Nickname, pk.Source).Scan(&id)

	return id.String, err
}

func (pk *PUBLIC_KEY) Delete(stmt *sql.Stmt) error {
	_, err := stmt.Exec(pk.Id)

	return err
}

func (pk *PUBLIC_KEY) Update(stmt *sql.Stmt) error {
	_, err := stmt.Exec(pk.Key, pk.Nickname, pk.Source, pk.Id)

	return err
}
