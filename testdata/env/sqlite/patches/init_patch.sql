-- id: fc8b6bf5-1ba1-44e9-aa05-a23e5dccf40d
-- prereqs: 
-- description: 

create table users (
	id integer primary key,
	name text not null,
	password blob,
	email text not null
);

create table orgs (
	id integer primary key,
	name text not null
);

create table account_holders (
	id integer primary key,
	user_id integer references users(id),
	org_id integer references orgs(id)
);

create table accounts (
	id integer primary key,
	actype text not null,
	holder integer not null references account_holders(id),
	amount float not null default 0
);

create table transactions (
	id integer primary key,
	created timestamp not null default CURRENT_TIMESTAMP,
	description text
);

create table transaction_parts (
	id integer primary key,
	account_id integer not null references accounts(id),
	amount float not null default 0.00,
	reconciled bool not null default 'f'
);

