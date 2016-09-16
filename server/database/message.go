// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package database

import (
	"database/sql"
	"time"
)

const (
	// add + delete
	MESSAGE_INSERT  = "insert into message (person_id, message, date_expires) values ($1, $2, $3) returning id"
	MESSAGE_DELETE  = "delete from message where id = $1"
	MESSAGE_CLEANUP = "select id from message where date_expires <= (now() at time zone 'UTC')"

	RECIPIENT_INSERT  = "insert into message_recipient (message_id, person_id) values ($1, $2)"
	RECIPIENT_DELETE  = "delete from message_recipient where message_id = $1 and person_id = $2"
	RECIPIENT_CLEANUP = "delete from message_recipient where message_id = $1"

	// lookups
	MESSAGES_BY_AUTHOR    = "select id, message, date_posted, date_expires from message where person_id = $1"
	MESSAGES_BY_RECIPIENT = "select m.id, m.message, m.date_posted, m.date_expires from message m, message_recipient mr where m.id = mr.message_id and mr.person_id = $1"
	RECIPIENTS_BY_MESSAGE = "select person_id from message_recipient where message_id = $1"
)

type MESSAGE struct {
	Id          string    `json:"id"`
	PersonId    string    `json:"person_id"`
	Message     string    `json:"message"`
	DatePosted  time.Time `json:"date_posted"`
	DateExpires time.Time `json:"date_expires"`
}

func (m *MESSAGE) Add(stmt *sql.Stmt, duration time.Duration) (string, error) {
	var id sql.NullString

	expires := time.Now().UTC().Add(duration)
	err := stmt.QueryRow(m.PersonId, m.Message, expires).Scan(&id)

	return id.String, err
}

func (m *MESSAGE) ProcessRecipients(stmt *sql.Stmt, recipients []*PERSON) []error {
	results := make([]error, 0)
	for _, recipient := range recipients {
		_, err := stmt.Exec(m.Id, recipient.Id)
		results = append(results, err)
	}

	return results
}

func (m *MESSAGE) AddRecipients(stmt *sql.Stmt, recipients []*PERSON) []error {
	return m.ProcessRecipients(stmt, recipients)
}

func (m *MESSAGE) Delete(stmt *sql.Stmt) error {
	_, err := stmt.Exec(m.Id)

	return err
}

func (m *MESSAGE) DeleteRecipients(stmt *sql.Stmt, recipients []*PERSON) []error {
	return m.ProcessRecipients(stmt, recipients)
}

// Return a list of message ids that are now expired
func ExpiredMessages(stmt *sql.Stmt) ([]string, error) {
	results := make([]string, 0)

	rows, err := stmt.Query()
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var id sql.NullString
		err := rows.Scan(&id)
		if err != nil {
			return results, err
		} else {
			results = append(results, id.String)
		}
	}

	return results, nil
}

// Identify all expired messages, and remove them and their recipient list
func CleanupMessages(idStmt, msgStmt, recipientStmt *sql.Stmt) []error {
	results := make([]error, 0)

	expired, err := ExpiredMessages(idStmt)
	results = append(results, err)
	if err != nil {
		for _, id := range expired {
			_, recipientErr := recipientStmt.Exec(id)
			results = append(results, recipientErr)

			_, msgErr := msgStmt.Exec(id)
			results = append(results, msgErr)
		}
	}

	return results
}
