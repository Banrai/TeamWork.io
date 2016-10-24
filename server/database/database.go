// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

type DBConnection struct {
	DBName  string
	User    string
	Pass    string
	SSLMode bool
}

// Connect to the database with the given coordinates, and invoke the
// function, which gets passed a map of all the prepared statements
func WithDatabase(dbCoords DBConnection, fn func(map[string]*sql.Stmt)) {
	preparedStatements := []string{PERSON_INSERT,
		PERSON_UPDATE,
		PERSON_DELETE,
		PERSON_LOOKUP_BY_ID,
		PERSON_LOOKUP_BY_EMAIL,
		SESSION_INSERT,
		SESSION_UPDATE,
		SESSION_CLEANUP,
		SESSION_LOOKUP_BY_CODE,
		SESSION_LOOKUP_BY_ID,
		SESSION_LOOKUP_BY_PERSON,
		PK_INSERT,
		PK_UPDATE,
		PK_DELETE,
		PK_LOOKUP,
		MESSAGE_INSERT,
		MESSAGE_DELETE,
		MESSAGE_CLEANUP,
		RECIPIENT_INSERT,
		RECIPIENT_DELETE,
		RECIPIENT_CLEANUP,
		MESSAGES_BY_AUTHOR,
		MESSAGES_BY_RECIPIENT,
		LATEST_MESSAGES,
		LATEST_MESSAGES_INVOLVING_PERSON,
		RECIPIENTS_BY_MESSAGE}

	sslMode := "disable"
	if dbCoords.SSLMode {
		sslMode = "enable"
	}

	db, dbErr := sql.Open("postgres",
		fmt.Sprintf("user=%s dbname=%s password=%s sslmode=%s",
			dbCoords.User,
			dbCoords.DBName,
			dbCoords.Pass,
			sslMode))
	if dbErr != nil {
		log.Fatal(dbErr)
	}
	defer db.Close()

	statements := map[string]*sql.Stmt{}
	for _, p := range preparedStatements {
		stmt, err := db.Prepare(p)
		if err != nil {
			log.Fatal(err)
		} else {
			statements[p] = stmt
		}
	}

	fn(statements)
}
