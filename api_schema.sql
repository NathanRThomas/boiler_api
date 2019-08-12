/*! \file api_schema.sql
	\brief This contains our "base" database sql schema written for cockroach and designed to work 
	with some of our base concepts in our boilerplate api code

	'00000000-0000-0000-0000-000000000000'	default uuid
*/

CREATE DATABASE test;
SET database = test;

-- Table creates
CREATE TABLE users (
	id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	token       TEXT NOT NULL,
	email       TEXT NOT NULL,
	password    TEXT NOT NULL,
	attrs 		JSONB NOT NULL DEFAULT '{}',
	created   	TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	mask       	INT NOT NULL DEFAULT 0 CHECK (mask >= 0),
	INDEX idx_users_email (email)
);


-- Primary Inserts
