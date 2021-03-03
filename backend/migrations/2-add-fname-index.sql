-- +migrate Up
CREATE INDEX `f_name_idx` USING BTREE ON `users` (`first_name`);

-- +migrate Down
ALTER TABLE `users`
    DROP INDEX `f_name_idx`;