// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/Banrai/TeamWork.io/server/api"
	"github.com/Banrai/TeamWork.io/server/database"
	"github.com/Banrai/TeamWork.io/server/ui"
	"log"
	"net/http"
)

const (
	// default server definitions
	server       = "teamwork.io"
	port         = 8080
	useSSL       = true
	externalPort = 443
	templates    = "/opt/data/html/templates"

	// default database coordinates
	DBName = "db"
	DBUser = "user"
	DBPass = "pass"
	DBSSL  = true
	WORDS  = "/usr/share/dict/words"

	// generate the static HTML files?
	statics      = false
	staticFolder = "/tmp"
)

func main() {
	var (
		dbName, dbUser, dbPass, serverHost, wordsFile, templatesFolder string
		serverPort, externalServerPort                                 int
		dbSSLMode, useServerSSL                                        bool
	)

	// get server settings from the command line args
	flag.StringVar(&serverHost, "host", server, "The hostname or IP address of the server")
	flag.IntVar(&serverPort, "port", port, "The server port")
	flag.IntVar(&externalServerPort, "extPort", externalPort, "The external server port")
	flag.BoolVar(&useServerSSL, "ssl", useSSL, "Does the server use SSL?")
	flag.StringVar(&templatesFolder, "templates", templates, "Path to html templates and static resources")

	// get database settings from the command line args
	flag.StringVar(&dbUser, "dbUser", DBUser, "The database user")
	flag.StringVar(&dbPass, "dbPass", DBPass, "The database password")
	flag.StringVar(&dbName, "dbName", DBName, "The database name")
	flag.BoolVar(&dbSSLMode, "dbSSL", DBSSL, "Does the database use SSL mode?")
	flag.StringVar(&wordsFile, "words", WORDS, "Dictionary file (for generating random session codes)")

	flag.Parse()

	coords := database.DBConnection{DBName: dbName, User: dbUser, Pass: dbPass, SSLMode: dbSSLMode}
	wordsInit := database.InitializeWords(wordsFile)
	if wordsInit != nil {
		log.Fatal(wordsInit)
	}
	log.Println(coords.DBName) // for now

	// define the external-facing server link
	// for email confirmations, etc.
	var buffer bytes.Buffer
	buffer.WriteString("http")
	if useServerSSL {
		buffer.WriteString("s")
	}
	buffer.WriteString("://")
	buffer.WriteString(serverHost)
	if !useServerSSL || externalServerPort != serverPort {
		// the port matters only if it is non-standard
		// for ssl or if not using ssl at all
		buffer.WriteString(fmt.Sprintf(":%d", externalServerPort))
	}
	serverLink := buffer.String()
	log.Println(serverLink) // for now

	statics := map[string]http.Handler{}

	statics["/css/"] = ui.StaticFolder("css", templatesFolder)
	statics["/images/"] = ui.StaticFolder("images", templatesFolder)

	handlers := map[string]func(http.ResponseWriter, *http.Request){}
	handlers["/browser/"] = ui.UnsupportedBrowserHandler(templatesFolder)

	api.RequestServer(serverHost, api.DefaultServerTransport, serverPort, api.DefaultServerReadTimeout, statics, handlers)
}