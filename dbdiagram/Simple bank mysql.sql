CREATE TABLE `accounts` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `owner` varchar(255) NOT NULL,
  `balance` bigint NOT NULL,
  `currency` varchar(255) NOT NULL,
  `created_at` timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE `entries` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `acc_id` bigint,
  `amount` bigint NOT NULL,
  `created_at` timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE `transfers` (
  `id` int PRIMARY KEY AUTO_INCREMENT,
  `from_account_id` bigint,
  `to_account_id` bigint,
  `amount` bigint NOT NULL,
  `created_at` timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX `accounts_index_0` ON `accounts` (`owner`);

CREATE INDEX `transfers_index_1` ON `transfers` (`from_account_id`);

CREATE INDEX `transfers_index_2` ON `transfers` (`to_account_id`);

CREATE INDEX `transfers_index_3` ON `transfers` (`from_account_id`, `to_account_id`);

ALTER TABLE `entries` ADD FOREIGN KEY (`acc_id`) REFERENCES `accounts` (`id`);

ALTER TABLE `transfers` ADD FOREIGN KEY (`from_account_id`) REFERENCES `accounts` (`id`);

ALTER TABLE `transfers` ADD FOREIGN KEY (`to_account_id`) REFERENCES `accounts` (`id`);
