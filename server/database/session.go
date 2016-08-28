// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package database

import (
	"bufio"
	"database/sql"
	"github.com/lib/pq"
	"math/rand"
	"os"
	"strings"
	"time"
)

var WordTokens []string

const (
	// session a/u/d
	SESSION_INSERT  = "insert into session (session_code, person_id, date_expires) values ($1, $2, $3) returning id"
	SESSION_UPDATE  = "update session set verified = $1, date_verified = (now() at time zone 'UTC') where id = $2"
	SESSION_CLEANUP = "delete from session where date_expires <= (now() at time zone 'UTC')"

	// session lookup
	SESSION_LOOKUP_BY_CODE   = "select id, person_id, date_created, verified, date_verified, date_expires from session where session_code = $1"
	SESSION_LOOKUP_BY_PERSON = "select id, session_code, date_created, verified, date_verified, date_expires from session where person_id = $1"
)

type SESSION struct {
	Id           string    `json:"id"`
	PersonId     string    `json:"person_id"`
	Code         string    `json:"session_code"`
	DateCreated  time.Time `json:"date_created"`
	Verified     bool      `json:"verified"`
	DateVerified time.Time `json:"date_verified"`
	DateExpires  time.Time `json:"date_expires"`
}

func (s *SESSION) Add(stmt *sql.Stmt, PersonId string, codeSize int, duration time.Duration) (string, error) {
	var id sql.NullString

	code := generateSessionCode(codeSize)
	expires := time.Now().UTC().Add(duration)

	err := stmt.QueryRow(code, PersonId, expires).Scan(&id)

	return id.String, err
}

func (s *SESSION) Delete(stmt *sql.Stmt) error {
	_, err := stmt.Exec(s.Id)

	return err
}

func (s *SESSION) Update(stmt *sql.Stmt) error {
	_, err := stmt.Exec(s.Verified, s.Id)

	return err
}

func LookupSession(stmt *sql.Stmt, code string) (*SESSION, error) {
	result := new(SESSION)

	rows, err := stmt.Query(code)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id, person_id                             sql.NullString
			verified                                  sql.NullBool
			date_created, date_verified, date_expires pq.NullTime
		)
		err := rows.Scan(&id, &person_id, &date_created, &verified, &date_verified, &date_expires)
		if err != nil {
			return result, err
		} else {
			result.Code = code
			result.Id = id.String
			result.PersonId = person_id.String
			result.DateCreated = date_created.Time
			result.Verified = verified.Bool
			result.DateVerified = date_verified.Time
			result.DateExpires = date_expires.Time
			break
		}
	}

	return result, nil
}

func CleanupSessions(stmt *sql.Stmt) error {
	_, err := stmt.Exec()

	return err
}

func generateSessionCode(size int) string {
	var codes []string
	l := len(WordTokens)
	for i := 0; i < size; i++ {
		codes = append(codes, WordTokens[rand.Intn(l)])
	}
	return strings.Join(codes, " ")
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func InitializeWords(wordsFile string) error {
	var err error
	WordTokens, err = readLines(wordsFile)
	return err
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
