-- PATCH: v0.0.1
-- id: 11151736-cf4d-11eb-9959-cb7e7c3dad34
-- author: cfox@infoblox.com
-- prereqs: f6b5f306-cf4c-11eb-9e42-bbfce6e80ed7
-- description: forgot email column in users table
alter table users add email text;
