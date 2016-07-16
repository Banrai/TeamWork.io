
CREATE TABLE person (
	id            uuid primary key DEFAULT uuid_generate_v4(),
	email         text NOT NULL,
	date_added    timestamp with time zone DEFAULT (now() at time zone 'UTC'),
	verified      boolean DEFAULT false,
	date_verified timestamp with time zone,
	enabled       boolean DEFAULT false,
	UNIQUE(email, id)
);

CREATE TABLE public_key (
	id         uuid primary key DEFAULT uuid_generate_v4(),
	person_id  uuid references person(id),
	key        text NOT NULL,
	date_added timestamp with time zone DEFAULT (now() at time zone 'UTC'),
	nickname   text,
	source     text,
	UNIQUE(person_id, nickname)
);

CREATE TABLE message (
	id           uuid primary key DEFAULT uuid_generate_v4(),
	person_id    uuid references person(id), -- author
	message      text NOT NULL,
	date_posted  timestamp with time zone DEFAULT (now() at time zone 'UTC'),
	date_expires  timestamp with time zone,
	UNIQUE(person_id, message, date_posted)
);

CREATE TABLE message_recipient (
	id           uuid primary key DEFAULT uuid_generate_v4(),
	message_id   uuid references message(id),
	person_id     uuid references person(id),
	UNIQUE(message_id, person_id)
);

CREATE TABLE session (
	id            uuid primary key DEFAULT uuid_generate_v4(),
	person_id     uuid references person(id),
	session_code  text NOT NULL,
	date_created  timestamp with time zone DEFAULT (now() at time zone 'UTC'),
	verified      boolean DEFAULT false,
	date_verified timestamp with time zone,
	date_expires  timestamp with time zone,
	UNIQUE(person_id, session_code)
);
