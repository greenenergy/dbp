-- PATCH: v0.0.1
-- id: 
-- author: greenenergy@gmail.com
-- description: Initial schema file.

create table users (
	id int serial primary key,
	username text not null,
	created timestamp with time zone not null default CURRENT_TIMESTAMP,
);

