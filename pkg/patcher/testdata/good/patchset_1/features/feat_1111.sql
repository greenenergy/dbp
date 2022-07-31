-- id: 071b5a20-d0aa-11eb-9566-f74f518b37a2 
-- prereqs: fece2b8e-cf43-11eb-b7f3-07af1b70a47a,f6b5f306-cf4c-11eb-9e42-bbfce6e80ed7
-- description: add accounts and transactions

create table accounts (
	id serial primary key,
	acct_holder integer not null references users(id),
	balance float
);

create table xactions (
	id serial primary key
);

create table xaction_parts (
	id serial primary key,
	transaction_id int not null references xactions(id),
	account_id int not null references accounts(id),
	amount float,
	reconciled bool not null default 'f'
);


