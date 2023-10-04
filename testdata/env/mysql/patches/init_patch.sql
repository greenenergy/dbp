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


