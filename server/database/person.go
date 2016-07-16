// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package database

import (
	"database/sql"
	"github.com/lib/pq"
	"time"
)

const (
	// person a/u/d
	PERSON_INSERT = "insert into person (email) values ($1) returning id"
	PERSON_UPDATE = "update person set email = $1, verified = $2, date_verified = (now() at time zone 'UTC'), enabled = $3 where id = $4"
	PERSON_DELETE = "delete from person where id = $1"

	// person lookup
	PERSON_LOOKUP_BY_ID    = "select id, email, date_added, verified, date_verified, enabled from person where id = $1"
	PERSON_LOOKUP_BY_EMAIL = "select id, email, date_added, verified, date_verified, enabled from person where email = $1"
)

type PERSON struct {
	Id           string    `json:"id"`
	Email        string    `json:"email"`
	DateAdded    time.Time `json:"date_joined"`
	Verified     bool      `json:"verified"`
	DateVerified time.Time `json:"date_verified,omitempty"`
	Enabled      bool      `json:"enabled"`
}

func (p *PERSON) Add(stmt *sql.Stmt) (string, error) {
	var id sql.NullString
	err := stmt.QueryRow(p.Email).Scan(&id)

	return id.String, err
}

func (p *PERSON) Delete(stmt *sql.Stmt) error {
	_, err := stmt.Exec(p.Id)

	return err
}

func (p *PERSON) Update(stmt *sql.Stmt) error {
	_, err := stmt.Exec(p.Email, p.Verified, p.Enabled, p.Id)

	return err
}

func LookupPerson(stmt *sql.Stmt, param string) (*PERSON, error) {
	result := new(PERSON)

	rows, err := stmt.Query(param)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id, email                 sql.NullString
			date_added, date_verified pq.NullTime
			verified, enabled         sql.NullBool
		)

		err := rows.Scan(&id, &email, &date_added, &verified, &date_verified, &enabled)
		if err != nil {
			return result, err
		} else {
			result.Id = id.String
			result.Email = email.String
			result.DateAdded = date_added.Time
			result.Verified = verified.Bool
			result.DateVerified = date_verified.Time
			result.Enabled = enabled.Bool

			break
		}
	}

	return result, nil
}
