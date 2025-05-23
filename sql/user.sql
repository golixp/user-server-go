CREATE TABLE user (
	id bigint unsigned PRIMARY KEY,
	username varchar ( 50 ) NOT NULL,
	nickname varchar ( 50 ) NOT NULL,
	password varchar ( 100 ) NOT NULL,
	login_at datetime NULL,
	login_ip char ( 16 ) NULL,
	created_at datetime NULL,
	updated_at datetime NULL,
	deleted_at datetime NULL
);

create index user_username_index
    on user (username);

create index user_info_deleted_at_index
    on user (deleted_at);