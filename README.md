# TeamWork.io

## About

This is the source code for the service running at [TeamWork.io](https://teamwork.io).

This repository is available for review by prospective users, as well as for those people who wish to clone it and run a version of the service on their own servers.

Contributions, both in the form of [code suggestions](https://github.com/Banrai/TeamWork.io/pulls) and [financial donations](https://teamwork.io/donate), are welcome!

## Core Concept

TeamWork.io is an encrypted message hub, a public bulletin board where you can post hidden messages to anyone, both individuals, or teams of people simultaneously, readable by only you and them.

It uses some of the basic ideas of [public key cryptography](http://www.pgpi.org/doc/pgpintro/#p9): if you know someone's public key, you can generate a message that only they can read, and that it's also possible to encrypt a message for a specific group of people, decipherable only by members of that group. 

## Acknowledgements

* The [OpenPGP JavaScript Implementation](https://openpgpjs.org/) project
* The [Let’s Encrypt](https://letsencrypt.org/) and [Certbot](https://certbot.eff.org/) projects
* [Twitter](https://twitter.com/) for the [Bootstrap framework](http://getbootstrap.com/)
* [Harvest](https://www.getharvest.com/) for the [Chosen plugin](https://harvesthq.github.io/chosen/) 
* [Dave Gandy](https://twitter.com/davegandy) for [Font Awesome](http://fontawesome.io/)
* [Cory LaViska](https://www.abeautifulsite.net/author/claviska) for "[Whipping File Inputs Into Shape with Bootstrap 3](https://www.abeautifulsite.net/whipping-file-inputs-into-shape-with-bootstrap-3)"
* [Stéphane Caron](https://scaron.info/) for his excellent [article on SPF and DKIM](https://scaron.info/blog/debian-mail-spf-dkim.html)
