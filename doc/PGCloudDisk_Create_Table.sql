CREATE DATABASE PGCloudDisk;

use PGCloudDisk;

CREATE TABLE `users` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `created_at` datetime DEFAULT (now()),
  `updated_at` datetime DEFAULT null,
  `deleted_at` datetime DEFAULT null,
  `username` varchar(50) NOT NULL UNIQUE,
  `password` varchar(128) NOT NULL
);

CREATE TABLE `files` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `created_at` datetime DEFAULT (now()),
  `updated_at` datetime DEFAULT null,
  `deleted_at` datetime DEFAULT null,
  `filename` varchar(256) NOT NULL,
  `size` bigint DEFAULT null,
  `location` varchar(1024) NOT NULL,
  `local_addr` varchar(1024),
  `type` char(8) NOT NULL,
  `user_id` bigint NOT NULL
);

ALTER TABLE `files` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`);

CREATE INDEX `users_index_on_deleted_at` ON `users` (`deleted_at`);

CREATE INDEX `users_index_on_username` ON `users` (`username`);

CREATE INDEX `files_index_on_deleted_at` ON `files` (`deleted_at`);

CREATE INDEX `files_index_on_filename` ON `files` (`filename`(128));

CREATE INDEX `files_index_on_location` ON `files` (`location`(128));

CREATE INDEX `files_index_on_user_id` ON `files` (`user_id`);
