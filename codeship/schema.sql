/*! \file schema.sql
	\brief This contains our "base" database sql schema written for cockroach and designed to work 
	with some of our base concepts in our boilerplate api code

	'00000000-0000-0000-0000-000000000000'	default uuid
*/

DROP DATABASE IF EXISTS test;
CREATE DATABASE test;
SET database = test;


-- TABLES ------------------------------------------------------------------------------------------------------------
CREATE TABLE users (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	token       TEXT NOT NULL,
	email       TEXT NOT NOT,
	password    TEXT NOT NULL,
    username    TEXT NOT NULL,
    attrs 		JSONB NOT NULL DEFAULT '{}',
	created   	TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	mask       	INT NOT NULL DEFAULT 0 CHECK (mask >= 0),
	INDEX idx_users_email (email),
    INDEX idx_users_password (password)
);

-- these are recuring things that need to happen over and over at some interval
CREATE TABLE schedules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_type INT NOT NULL CHECK (schedule_type > 0),
    created   	TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    next_date   TIMESTAMPTZ NOT NULL,
    attrs       JSONB NOT NULL DEFAULT '{}',
    "interval"  TEXT NOT NULL,
    INDEX idx_schedules_next (next_date)
);


-- INSERTS ------------------------------------------------------------------------------------------------------------


