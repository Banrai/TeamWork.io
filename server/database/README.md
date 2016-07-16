# About

The server is powered by a [PostgreSQL](https://www.postgresql.org/) database.

These installation and setup instructions are specially for [Debian](http://www.debian.org/)/[Ubuntu](http://www.ubuntu.com/) systems, but can be adapted to other flavors of Linux easily.

## Installation

Install PostgreSQL using the default package as root/sudo, along with [postgresql-contrib](https://www.postgresql.org/docs/current/static/contrib.html), since the data model uses universally unique identifiers (UUIDs) for its table primary keys:

```sh
apt-get install -y postgresql postgresql-contrib
```

## Configuration

Once the packages have been installed, create the database user, database itself, and set the access level.

1. Become or login as the <tt>postgres</tt> user:

  ```sh
su - postgres
```

2. Create the database user (have a password ready):

  ```sh
createuser -P teamworkio
```

3. Next, the database itself:

  ```sh
createdb teamworkdb
```

4. Then apply the UUID extension:

  ```sh
psql -d teamworkdb
teamworkdb=# create extension "uuid-ossp";
CREATE EXTENSION
teamworkdb=# \q
exit
```

5. Configure the client authentication (optional) 

  Stop the database server:

  ```sh
/etc/init.d/postgresql stop
```

  Then edit the [pg_hba.conf](https://www.postgresql.org/docs/current/static/auth-pg-hba-conf.html) file with your desired settings.

  Finally, restart the database server:

  ```sh
/etc/init.d/postgresql restart
```

## Installing the data model

With the database created and the postgres server running, install the [tables.sql](tables.sql) file from this folder, as any system user:

```sh
psql -d teamworkdb -U teamworkio < tables.sql
```

The [psql](https://www.postgresql.org/docs/current/static/app-psql.html) tool will prompt for the password used in step 2, above:

```sh
Password for user teamworkio: 
```

The database will respond with:

```sh
CREATE TABLE
CREATE TABLE
CREATE TABLE
CREATE TABLE
CREATE TABLE
```
