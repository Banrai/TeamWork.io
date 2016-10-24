// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package database

import (
	"database/sql"
	"github.com/lib/pq"
	"strings"
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
	err := stmt.QueryRow(strings.ToLower(p.Email)).Scan(&id)

	return id.String, err
}

func (p *PERSON) Delete(stmt *sql.Stmt) error {
	_, err := stmt.Exec(p.Id)

	return err
}

func (p *PERSON) Update(stmt *sql.Stmt) error {
	_, err := stmt.Exec(strings.ToLower(p.Email), p.Verified, p.Enabled, p.Id)

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

func (p *PERSON) LookupSessions(stmt *sql.Stmt) ([]*SESSION, error) {
	results := make([]*SESSION, 0)

	rows, err := stmt.Query(p.Id)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id, person_id, session_code               sql.NullString
			verified                                  sql.NullBool
			date_created, date_verified, date_expires pq.NullTime
		)
		err := rows.Scan(&id, &person_id, &session_code, &date_created, &verified, &date_verified, &date_expires)
		if err != nil {
			return results, err
		} else {
			result := new(SESSION)
			result.PersonId = person_id.String
			result.Id = id.String
			result.Code = session_code.String
			result.DateCreated = date_created.Time
			result.Verified = verified.Bool
			result.DateVerified = date_verified.Time
			result.DateExpires = date_expires.Time
			results = append(results, result)
		}
	}

	return results, nil
}

func (p *PERSON) LookupPublicKeys(stmt *sql.Stmt) ([]*PUBLIC_KEY, error) {
	results := make([]*PUBLIC_KEY, 0)

	rows, err := stmt.Query(p.Id)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id, public_key, nickname, source sql.NullString
			date_added                       pq.NullTime
		)
		err := rows.Scan(&id, &public_key, &date_added, &nickname, &source)
		if err != nil {
			return results, err
		} else {
			result := new(PUBLIC_KEY)
			result.Id = id.String
			result.Key = public_key.String
			result.Added = date_added.Time
			if nickname.Valid {
				result.Nickname = nickname.String
			}
			if source.Valid {
				result.Source = source.String
			}
			results = append(results, result)
		}
	}

	return results, nil
}

func (p *PERSON) LookupMessages(stmt *sql.Stmt, usePersonId bool, limit, offset int64) ([]*MESSAGE, error) {
	if usePersonId {
		return RetrieveMessages(stmt, p.Id, limit, offset)
	} else {
		return RetrieveMessages(stmt, "", limit, offset)
	}
}

// Return a list of all messages, regardless of involvement by this person
func (p *PERSON) LookupLatestMessages(stmt *sql.Stmt, limit, offset int64) ([]*MESSAGE, error) {
	return p.LookupMessages(stmt, false, limit, offset)
}

// Return a list of messages originated by this person
func (p *PERSON) LookupAuthoredMessages(stmt *sql.Stmt, limit, offset int64) ([]*MESSAGE, error) {
	return p.LookupMessages(stmt, true, limit, offset)
}

// Return a list of messages in which this person was a recipient
func (p *PERSON) LookupRecipientMessages(stmt *sql.Stmt, limit, offset int64) ([]*MESSAGE, error) {
	return p.LookupMessages(stmt, true, limit, offset)
}

// Return a list of messages in which this person was involved, either as an origniator or a recipient
func (p *PERSON) LookupInvolvedMessages(stmt *sql.Stmt, limit, offset int64) ([]*MESSAGE, error) {
	return p.LookupMessages(stmt, true, limit, offset)
}

// create a new Person in the db, and associate these public keys
func AddPersonWithKeys(pStmt, pkStmt *sql.Stmt, email string, pkList []*PUBLIC_KEY) error {
	p := new(PERSON)
	p.Email = email
	pId, pErr := p.Add(pStmt)
	if pErr != nil {
		return pErr
	}

	for _, key := range pkList {
		_, keyErr := key.Add(pkStmt, pId)
		if keyErr != nil {
			return keyErr
		}
	}

	return nil
}
