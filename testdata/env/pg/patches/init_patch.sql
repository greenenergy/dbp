-- id: fece2b8e-cf43-11eb-b7f3-07af1b70a47a
-- description: Initial schema file.

create table users (
	id serial primary key,
	username text not null,
	created timestamp with time zone not null default CURRENT_TIMESTAMP
);

