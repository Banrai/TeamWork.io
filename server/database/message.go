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
	// add + delete
	MESSAGE_INSERT  = "insert into message (person_id, message, date_expires) values ($1, $2, $3 at time zone 'UTC') returning id"
	MESSAGE_DELETE  = "delete from message where id = $1"
	MESSAGE_CLEANUP = "select id from message where date_expires <= (now() at time zone 'UTC')"

	RECIPIENT_INSERT  = "insert into message_recipient (message_id, person_id) values ($1, $2)"
	RECIPIENT_DELETE  = "delete from message_recipient where message_id = $1 and person_id = $2"
	RECIPIENT_CLEANUP = "delete from message_recipient where message_id = $1"

	// lookups
	MESSAGES_BY_AUTHOR               = "select id, person_id, message, date_posted, date_expires from message where person_id = $1"
	MESSAGES_BY_RECIPIENT            = "select m.id, m.person_id, m.message, m.date_posted, m.date_expires from message m, message_recipient mr where m.id = mr.message_id and mr.person_id = $1"
	RECIPIENTS_BY_MESSAGE            = "select person_id from message_recipient where message_id = $1"
	LATEST_MESSAGES                  = "select id, person_id, message, date_posted, date_expires from message order by date_posted desc limit $1 offset $2"
	LATEST_MESSAGES_INVOLVING_PERSON = `select distinct m.id, m.person_id, m.date_posted, m.date_expires
	from message m, message_recipient mr
	where m.id = mr.message_id
	and (m.person_id = $1 or mr.person_id = $1)
	order by m.date_posted desc
	limit $2 offset $3`
	MESSAGE_BY_ID = "select id, person_id, message, date_posted, date_expires from message where id = $1 limit $2 offset $3"
)

type MESSAGE struct {
	Id          string    `json:"id"`
	PersonId    string    `json:"person_id"`
	Message     string    `json:"message"`
	DatePosted  time.Time `json:"date_posted"`
	DateExpires time.Time `json:"date_expires"`
}

type MESSAGE_DIGEST struct {
	Message           *MESSAGE
	Preview           string
	Sender            *PERSON
	Recipients        []*PERSON
	InvolvesRequestor bool
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

// Return a list of messages for the given query limit/offset criteria
func RetrieveMessages(stmt *sql.Stmt, uniqueId string, limit, offset int64) ([]*MESSAGE, error) {
	results := make([]*MESSAGE, 0)

	var (
		rows *sql.Rows
		err  error
	)
	if len(uniqueId) > 0 {
		rows, err = stmt.Query(uniqueId, limit, offset)
	} else {
		rows, err = stmt.Query(limit, offset)
	}
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id, person_id, message    sql.NullString
			date_posted, date_expires pq.NullTime
		)
		err := rows.Scan(&id, &person_id, &message, &date_posted, &date_expires)
		if err != nil {
			return results, err
		} else {
			result := new(MESSAGE)
			result.Id = id.String
			result.PersonId = person_id.String
			result.Message = message.String
			result.DatePosted = date_posted.Time
			result.DateExpires = date_expires.Time
			results = append(results, result)
		}
	}

	return results, nil
}

// Fetch the first unique line of the armored message as a preview
func (m *MESSAGE) GetPreview() string {
	for i, line := range strings.Split(m.Message, "\r\n") {
		l := len(line)
		if l > 0 && i > 4 {
			if !strings.HasPrefix(line, "-----BEGIN") && !strings.HasPrefix(line, "Version:") && !strings.HasPrefix(line, "Comment:") {
				until := 35
				if l < until {
					until = l - 1
				}
				return line[0:until]
			}
		}
	}
	return "[no preview available]"
}

// Retrieve the corresponding digest (which includes all the involved Person objects) for this Message
func (m *MESSAGE) GetDigest(personStmt, recipientStmt *sql.Stmt, personId string) (*MESSAGE_DIGEST, error) {
	result := &MESSAGE_DIGEST{Message: m, Preview: m.GetPreview()}
	result.InvolvesRequestor = (m.PersonId == personId)

	var (
		person      *PERSON
		personError error
	)

	person, personError = LookupPerson(personStmt, m.PersonId)
	if personError != nil {
		return result, personError
	}
	result.Sender = person

	rows, err := recipientStmt.Query(m.Id)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	recipients := make([]*PERSON, 0)
	for rows.Next() {
		var id sql.NullString
		err := rows.Scan(&id)
		if err != nil {
			return result, err
		} else {
			person, personError = LookupPerson(personStmt, id.String)
			if personError != nil {
				return result, personError
			}
			if !result.InvolvesRequestor {
				result.InvolvesRequestor = (person.Id == personId)
			}
			recipients = append(recipients, person)
		}
	}
	result.Recipients = recipients

	return result, nil
}

// Return the corresponding digests for this list of messages
func GetMessageDigests(personStmt, recipientStmt *sql.Stmt, messages []*MESSAGE, personId string) ([]*MESSAGE_DIGEST, []error) {
	digests := make([]*MESSAGE_DIGEST, 0)
	errors := make([]error, 0)

	for _, message := range messages {
		digest, err := message.GetDigest(personStmt, recipientStmt, personId)
		digests = append(digests, digest)
		errors = append(errors, err)
	}

	return digests, errors
}
