-- PATCH: v0.0.1
-- id: 61c44b25-4cea-49f3-a7c7-6ac930d33d56
-- author: 
-- prereqs: fc8b6bf5-1ba1-44e9-aa05-a23e5dccf40d
-- description: 

insert into users(name, email) values ('cfox', 'greenenergy@gmail.com'); -- 1
insert into orgs(name) values ('supplier 1'); -- 1
insert into orgs(name) values ('client 1'); -- 2

insert into account_holders(user_id) values (1); -- 1
insert into account_holders(org_id) values (1); -- 2
insert into account_holders(org_id) values (2); -- 3

insert into accounts(actype, holder, amount) values ('deposit', 1, 0); -- 1
insert into accounts(actype, holder, amount) values ('chequing', 1, 0); -- 2
insert into accounts(actype, holder, amount) values ('AP', 2, 0); -- 3
insert into accounts(actype, holder, amount) values ('AR', 2, 0); -- 4

insert into accounts(actype, holder, amount) values ('Purchase', 3, 0);

-- Create a $10 transaction, from 'supplier 1' to cfox's deposit account.
insert into transactions(description) values ('test1'); -- 1
insert into transaction_parts(account_id, amount) values (3, -10);
insert into transaction_parts(account_id, amount) values (1, 10);
