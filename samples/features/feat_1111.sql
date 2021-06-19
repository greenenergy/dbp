-- PATCH: v0.0.1
-- id: 071b5a20-d0aa-11eb-9566-f74f518b37a2 
-- author: greenenergy@gmail.com
-- prereqs: fece2b8e-cf43-11eb-b7f3-07af1b70a47a,f6b5f306-cf4c-11eb-9e42-bbfce6e80ed7
-- description: add users table
create table accounts (
	id int serial primary key,
	acct_holder integer not null references users(id),
	balance float,
);

create table transactions (
	id int serial primary key
);

create table transaction_parts {
	id int serial primary key,
	transaction_id int not null transactions(id),
	amount float,
	reconciled bool not null default 'f',
};
