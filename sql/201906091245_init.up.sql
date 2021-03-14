CREATE TABLE `Users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `firstName` varchar(128) NOT NULL DEFAULT '',
  `lastName` varchar(128) NOT NULL DEFAULT '',
  `email` varchar(256) NOT NULL DEFAULT '', -- can be blank for cross-posted, since we odn't want to leak emails
  `password` varchar(128) NOT NULL DEFAULT '', 
  `created` datetime NOT NULL,
  `status` enum('pending','verified') DEFAULT 'pending',
  `username` varchar(32) NOT NULL,
  `updated` datetime NOT NULL DEFAULT '1969-01-01 00:00:00',
  `lastLogin` datetime NOT NULL DEFAULT '1969-01-01 00:00:00',
  `platformRole` enum("member", "admin") DEFAULT "member",
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Communities` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL DEFAULT '',
  `description` varchar(1024) NOT NULL DEFAULT '',
  `shortCode` varchar(24) NOT NULL DEFAULT '',
  `joinCode` varchar(24) NOT NULL DEFAULT '',
  `created` datetime NOT NULL,
  `privacy` ENUM('private', 'public') NOT NULL DEFAULT 'private',
  `userSignupStatus` ENUM('none', 'short_code', 'approval_required', 'auto_accept') NOT NULL DEFAULT 'auto_accept', 
  `plan` ENUM('free','basic','pro') NOT NULL DEFAULT 'free', -- the plan for the community, which essentially just places limits on what they can do
  `planPaidThrough` date NOT NULL DEFAULT '1970-01-01',
  `planDiscountPercent` int NOT NULL DEFAULT 0,
  `stripeSubscriptionId` varchar(32) NOT NULL DEFAULT '',
  PRIMARY KEY (`id`),
  UNIQUE KEY `shortCode` (`shortCode`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `CommunityPayments` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `communityId` int(11) NOT NULL,
  `paymentDate` date NOT NULL,
  `status` ENUM('paid','declined') NOT NULL DEFAULT 'paid',
  `amountDue` int(11) NOT NULL DEFAULT 0,
  `amountPaid` int(11) NOT NULL DEFAULT 0,
  `notes` varchar(512) NOT NULL,
  PRIMARY KEY (`id`),
  KEY (`communityId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `CommunityUserLinks` (
  `userId` int(11) DEFAULT NULL,
  `communityId` int(11) DEFAULT NULL,
  `role` enum('member','admin') NOT NULL DEFAULT 'member',
  `status` enum('invited','requested','accepted','declined') NOT NULL DEFAULT 'invited',
  `shortCode` varchar(24) NOT NULL DEFAULT '', -- this is used in the invitation email
  UNIQUE KEY `user-community` (`userId`,`communityId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `UserTokens` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `token` varchar(64) DEFAULT NULL,
  `userId` int(11) NOT NULL,
  `created` datetime NOT NULL,
  `tokenType` enum('email','password_reset','refresh') NOT NULL DEFAULT 'email',
  PRIMARY KEY (`id`),
  KEY `userId` (`userId`),
  KEY `userId_type` (`userId`, `tokenType`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `PrayerRequests` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(256) DEFAULT '',
  `body` text,
  `createdBy` int(11) NOT NULL,
  `privacy` enum('public','private') DEFAULT 'private',
  `created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` enum('pending','answered','not_answered','unknown') DEFAULT 'pending',
  PRIMARY KEY (`id`),
  KEY `createdBy` (`createdBy`),
  KEY `privacy` (`privacy`),
  KEY `created` (`created`),
  KEY `createdBy_created` (`createdBy`,`created`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `PrayerRequestCommunityLinks` (
  `prayerRequestId` int(11) NOT NULL,
  `communityId` int(11) NOT NULL,
  UNIQUE KEY `link` (`prayerRequestId`, `communityId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Prayers` (
  `prayerRequestId` int(11) NOT NULL,
  `userId` int(11) NOT NULL,
  `whenPrayed` dateTime NOT NULL,
  KEY `request_user` (`prayerRequestId`, `userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `PrayerRequestTags` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `tag` varchar(32) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `tag` (`tag`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `PrayerRequestTagLinks` (
  `tagId` int(11) NOT NULL,
  `prayerRequestId` int(11) NOT NULL,
  UNIQUE `tagPR` (`tagId`, `prayerRequestId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Site` (
  `name` varchar(128) NOT NULL DEFAULT 'Pregxas',
  `description` varchar(1024) NOT NULL DEFAULT '',
  `secretKey` varchar(32) NOT NULL,
  `status` enum('pending_setup','active') NOT NULL DEFAULT 'pending_setup',
  `logoLocation` varchar(256) NOT NULL DEFAULT ''
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Reports` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `requestId` int(11) NOT NULL,
  `reporterId` int(11) NOT NULL,
  `reason` enum('offensive','threat','copyright','other') DEFAULT 'other',
  `reasonText` varchar(2048) NOT NULL DEFAULT '',
  `reported` datetime NOT NULL,
  `updated` datetime NOT NULL,
  `status` enum('open','closed_no_action','closed_deleted','follow_up') NOT NULL DEFAULT 'open',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `PrayerLists` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `userId` int(11) NOT NULL,
  `title` varchar(256) NOT NULL,
  `updateFrequency` enum('daily','weekly','never') NOT NULL DEFAULT 'never',
  `created` datetime NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `PrayerRequestPrayerListLinks` (
  `prayerRequestId` int(11) NOT NULL,
  `listId` int(11) NOT NULL,
  `added` datetime NOT NULL,
  UNIQUE KEY `request_list` (`prayerRequestId`,`listId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;