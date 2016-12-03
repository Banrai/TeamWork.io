// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"github.com/Banrai/TeamWork.io/server/api"
	"github.com/Banrai/TeamWork.io/server/database"
	"github.com/Banrai/TeamWork.io/server/ui"
	"log"
	"net/http"
)

const (
	// default server definitions
	hostname  = "teamwork.io"
	server    = "localhost"
	port      = 8080
	useSSL    = true
	templates = "/opt/data/html/templates"

	// default database coordinates
	DBName = "db"
	DBUser = "user"
	DBPass = "pass"
	DBSSL  = true
	WORDS  = "/usr/share/dict/words"

	// process donations with stripe.com
	stripeDefaultPK = "pk_test_"
	stripeDefaultSK = "sk_test_"

	// generate the static HTML files?
	statics      = false
	staticFolder = "/tmp"
)

func main() {
	var (
		dbName, dbUser, dbPass, hostName, serverHost, wordsFile, templatesFolder, staticOutputFolder, stripePK, stripeSK string
		serverPort                                                                                                       int
		dbSSLMode, useServerSSL, makeStaticFiles                                                                         bool
	)

	// get server settings from the command line args
	flag.StringVar(&hostName, "host", hostname, "The (externally-facing) name of the server")
	flag.StringVar(&serverHost, "ip", server, "The hostname or IP address of the server")
	flag.IntVar(&serverPort, "port", port, "The server port")
	flag.BoolVar(&useServerSSL, "ssl", useSSL, "Does the server use SSL?")
	flag.StringVar(&templatesFolder, "templates", templates, "Path to html templates and static resources")

	// get database settings from the command line args
	flag.StringVar(&dbUser, "dbUser", DBUser, "The database user")
	flag.StringVar(&dbPass, "dbPass", DBPass, "The database password")
	flag.StringVar(&dbName, "dbName", DBName, "The database name")
	flag.BoolVar(&dbSSLMode, "dbSSL", DBSSL, "Does the database use SSL mode?")
	flag.StringVar(&wordsFile, "words", WORDS, "Dictionary file (for generating random session codes)")

	// get the payment coordinates
	flag.StringVar(&stripePK, "stripePK", stripeDefaultPK, "The Stripe Public Key")
	flag.StringVar(&stripeSK, "stripeSK", stripeDefaultSK, "The Stripe Secret Key")

	// versus static file generation and exit
	flag.BoolVar(&makeStaticFiles, "staticHtml", statics, "Generate the static HTML files? (if yes, does not start the server)")
	flag.StringVar(&staticOutputFolder, "staticHtmlFolder", staticFolder, "Output folder for the static HTML files")

	flag.Parse()

	coords := database.DBConnection{DBName: dbName, User: dbUser, Pass: dbPass, SSLMode: dbSSLMode}
	wordsInit := database.InitializeWords(wordsFile)
	if wordsInit != nil {
		log.Fatal(wordsInit)
	}

	// define the external-facing server link
	// for email confirmations, etc.
	var buffer bytes.Buffer
	buffer.WriteString("http")
	if useServerSSL {
		buffer.WriteString("s")
	}
	buffer.WriteString("://")
	buffer.WriteString(hostName)

	serverLink := make([]interface{}, 1)
	serverLink[0] = buffer.String()

	statics := map[string]http.Handler{}
	statics["/css/"] = ui.StaticFolder("css", templatesFolder)
	statics["/js/"] = ui.StaticFolder("js", templatesFolder)
	statics["/fonts/"] = ui.StaticFolder("fonts", templatesFolder)
	statics["/images/"] = ui.StaticFolder("images", templatesFolder)

	ui.InitializeTemplates(templatesFolder)

	handlers := map[string]func(http.ResponseWriter, *http.Request){}
	handlers["/browser/"] = ui.UnsupportedBrowserHandler(templatesFolder)
	handlers["/addpost"] = ui.MakeHTMLHandler(ui.PostMessage, coords)
	handlers["/session"] = ui.MakeHTMLHandler(ui.CreateSession, coords)
	handlers["/confirm"] = ui.MakeHTMLHandler(ui.ConfirmSession, coords)
	handlers["/upload"] = ui.MakeHTMLHandler(ui.UploadKey, coords)
	handlers["/posts"] = ui.MakeHTMLHandler(ui.DisplayPosts, coords)
	handlers["/download"] = ui.MakeHTMLHandler(ui.DownloadMessage, coords, serverLink[0])

	// payment processing requires some additional parameters
	stripeVals := make([]interface{}, 2)
	stripeVals[0] = stripePK
	stripeVals[1] = stripeSK
	handlers["/donate"] = ui.MakeHTMLHandler(ui.ProcessDonation, coords, stripeVals[0], stripeVals[1])

	handlers["/searchPublicKeys"] = func(w http.ResponseWriter, r *http.Request) {
		lookup := func(w http.ResponseWriter, r *http.Request) string {
			return api.SearchPersonPublicKeys(r, coords)
		}
		api.Respond("application/json", "utf-8", lookup)(w, r)
	}

	if makeStaticFiles {
		ui.GenerateStaticFiles(templatesFolder, staticFolder)
	} else {
		api.RequestServer(serverHost, api.DefaultServerTransport, serverPort, api.DefaultServerReadTimeout, statics, handlers)
	}

}
