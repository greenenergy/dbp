-- PATCH: v0.0.1
-- id: fece2b8e-cf43-11eb-b7f3-07af1b70a47a
-- author: greenenergy@gmail.com
-- description: Phones table

create table phones (
	id int serial primary key,
	created timestamp with time zone not null default CURRENT_TIMESTAMP,
	phonenum text not null
);

