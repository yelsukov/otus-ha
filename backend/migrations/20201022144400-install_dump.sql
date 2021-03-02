-- +migrate Up
CREATE TABLE `users`
(
    `id`            int(10) unsigned NOT NULL AUTO_INCREMENT,
    `username`      varchar(30)      NOT NULL,
    `first_name`    varchar(20)               DEFAULT NULL,
    `last_name`     varchar(30)               DEFAULT NULL,
    `age`           tinyint(3)                DEFAULT NULL,
    `gender`        char(1)          NOT NULL,
    `city`          varchar(80)               DEFAULT NULL,
    `password_hash` varchar(255)     NOT NULL,
    `created_at`    timestamp        NOT NULL DEFAULT current_timestamp(),
    `interests`     varchar(255)              DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `username` (`username`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

ALTER TABLE users
    ADD UNIQUE INDEX un_idx (username);

INSERT INTO `users`
VALUES (1, 'janitor', 'Roger', 'Wilco', 25, 'm', 'Unknown',
        '$2a$10$sS1EKXztHWywwsQr5xCERe92goE2UIUuOXF.yrabdH1aRGxbIx2J.', NOW(), 'Do nothing, Get troubles'),
       (2, 'madscie', 'Sludge', 'Vohaul', 125, 'm', 'Hidden Space Base',
        '$2a$10$sS1EKXztHWywwsQr5xCERe92goE2UIUuOXF.yrabdH1aRGxbIx2J.', NOW(),
        'Experiments, Evil Plans, Conquer the Galaxy');

CREATE TABLE `friends`
(
    `user_id`    int(10)   NOT NULL,
    `friend_id`  int(10)   NOT NULL,
    `created_at` timestamp NOT NULL DEFAULT current_timestamp(),
    PRIMARY KEY (`user_id`, `friend_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

-- +migrate Down
DROP TABLE users;
DROP TABLE friends;