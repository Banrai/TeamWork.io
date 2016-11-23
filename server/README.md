# Server Installation

## Go

This version is written primarily in [Go](https://golang.org/), so the server needs a working installation.

Follow the instructions here: [https://golang.org/doc/install](https://golang.org/doc/install)

## Dependencies

The server code has these dependencies, which can be fulfilled using <tt>go get</tt>:

```sh
$ go get github.com/lib/pq
$ go get golang.org/x/crypto/openpgp
$ go get github.com/stripe/stripe-go
```

[Stripe](https://stripe.com/) is used to process donations on the [TeamWork.io](https://teamwork.io) site, which is why their library is necessary here too.

It is possible to run this service without a Stripe account, but for the time being, the library is still needed to build the server binary (at some point in the future, this dependency will be removed).

With the above dependencies installed, fetch this package:

```sh
$ go get github.com/Banrai/TeamWork.io
```

There will be a warning about no buildable files, like this, which can be ignored:

```sh
package github.com/Banrai/TeamWork.io: no buildable Go source files in $GOPATH/src/github.com/Banrai/TeamWork.io
```

## Build 

Finally, back in this project's root folder (i.e., inside <tt>$GOPATH/src/github.com/Banrai/TeamWork.io</tt>), use the [Makefile](../Makefile) to create the server binary:

```sh
$ make all
go build -o $GOPATH/src/github.com/Banrai/TeamWork.io/server/TeamWorkServer $GOPATH/src/github.com/Banrai/TeamWork.io/server/main.go
```

Which should produce a <tt>TeamWorkServer</tt> binary file in the <tt>$GOPATH/src/github.com/Banrai/TeamWork.io</tt> folder.

## Generate the static HTML files (optional)

The default package comes with a few static files for running in combination with a web server, but are not needed for using the core service.

```sh
$ cd $GOPATH/src/github.com/Banrai/TeamWork.io/server/
$ ./TeamWorkServer -staticHtml=true -templates=$GOPATH/src/github.com/Banrai/TeamWork.io/html/templates
```

This produces the static files in <tt>/tmp</tt>, which can be moved to <tt>/var/www/html</tt> or wherever the web server document root is.

The output target can be changed from <tt>/tmp</tt> by setting the <tt>-staticHtmlFolder</tt> option to another folder instead.

## Run

Use the <tt>--help</tt> option to see the configuration settings:

```sh
$ ./TeamWorkServer --help
Usage of ./TeamWorkServer:
  -dbName string
    	The database name (default "db")
  -dbPass string
    	The database password (default "pass")
  -dbSSL
    	Does the database use SSL mode? (default true)
  -dbUser string
    	The database user (default "user")
  -extPort int
    	The external server port (default 443)
  -host string
    	The hostname or IP address of the server (default "teamwork.io")
  -port int
    	The server port (default 8080)
  -ssl
    	Does the server use SSL? (default true)
  -staticHtml
    	Generate the static HTML files? (if yes, does not start the server)
  -staticHtmlFolder string
    	Output folder for the static HTML files (default "/tmp")
  -stripePK string
    	The Stripe Public Key (default "pk_test_")
  -stripeSK string
    	The Stripe Secret Key (default "sk_test_")
  -templates string
    	Path to html templates and static resources (default "/opt/data/html/templates")
  -words string
    	Dictionary file (for generating random session codes) (default "/usr/share/dict/words")
```

Saving these in an [LSBInitScript](init.d/README.md) and running from <tt>/etc/init.d</tt> is recommended.
