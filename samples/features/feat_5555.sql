-- PATCH: v0.0.1
-- id: f6b5f306-cf4c-11eb-9e42-bbfce6e80ed7
-- author: greenenergy@gmail.com
-- prereqs: fece2b8e-cf43-11eb-b7f3-07af1b70a47a
-- description: add users table
create table users (
	id int serial primary key,
	username text not null,
	created timestamp with time zone not null default CURRENT_TIMESTAMP,
);

