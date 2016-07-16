# TeamWork.io

## About

This is the source code for the service soon to be up and running at [https://teamwork.io](https://teamwork.io).

This repository is available for review by prospective users, as well as for those people who wish to clone it and run a version of the service on their own servers.

Contributions are welcome.

## Core Concept

TeamWork.io is an encrypted message hub, a public bulletin board where you can post hidden messages to anyone, both individuals, or teams of people simultaneously, readable by only you and them.

It uses some of the basic ideas of [public key cryptography](http://www.pgpi.org/doc/pgpintro/#p9): if you know someone's public key, you can generate a message that only they can read, and that it's also possible to encrypt a message for a specific group of people, decipherable only by members of that group. 