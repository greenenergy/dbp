-- id: init_patch
-- prereqs: 
-- description: 

CREATE TABLE users (
    username VARCHAR(255) NOT NULL,
    passwordhash CHAR(60) NOT NULL,
    fullname VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    creation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (username),
	UNIQUE (email)
);

 create table orgs (
 	orgname varchar(255) not null,
 	creator varchar(255) not null,
 	fullname varchar(255),
     creation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
 	foreign key(creator) references users(username)
 		on delete cascade
 		on update cascade,
 	primary key(orgname)
 );
 
 create table org_members (
 	id int auto incremement,
 	username varchar(255) not null,
 	orgname varchar(255) not null,
     creation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	primary key id,
 	foreign key(username) references users(username)
 		on delete cascade,
 	foreign key(orgname) references orgs(orgname)
 		on delete cascade,
 	unique (username, orgname)
 );
 