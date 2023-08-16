CREATE TABLE attachments
(
    id bigint NOT NULL AUTO_INCREMENT,
    todo_id bigint NOT NULL,
    path varchar(255) not null,
    attachment_order int not null,
    timestamp timestamp default current_timestamp,
    PRIMARY KEY (id),
    FOREIGN KEY (todo_id) REFERENCES todolists(id) ON DELETE CASCADE
);
