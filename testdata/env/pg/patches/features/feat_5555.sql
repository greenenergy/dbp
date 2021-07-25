-- id: f6b5f306-cf4c-11eb-9e42-bbfce6e80ed7
-- prereqs: fece2b8e-cf43-11eb-b7f3-07af1b70a47a
-- description: add groups table
create table groups (
	id serial primary key,
	name text not null,
	created timestamp with time zone not null default CURRENT_TIMESTAMP
);

create table group_members (
	id serial primary key,
	user_id int not null references users(id),
	created timestamp with time zone not null default CURRENT_TIMESTAMP
);

