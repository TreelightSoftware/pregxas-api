CREATE TABLE `Users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `firstName` varchar(128) NOT NULL DEFAULT '',
  `lastName` varchar(128) NOT NULL DEFAULT '',
  `email` varchar(256) NOT NULL DEFAULT '', -- can be blank for cross-posted, since we odn't want to leak emails
  `password` varchar(128) NOT NULL DEFAULT '', 
  `dateCreated` datetime NOT NULL,
  `status` enum('pending','verified') DEFAULT 'pending',
  `username` varchar(32) NOT NULL,
  `updated` datetime NOT NULL DEFAULT '1969-01-01 00:00:00',
  `lastLogin` datetime NOT NULL DEFAULT '1969-01-01 00:00:00',
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `Communities` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(128) NOT NULL DEFAULT '',
  `description` varchar(1024) NOT NULL DEFAULT '',
  `shortCode` varchar(24) NOT NULL DEFAULT '',
  `created` datetime NOT NULL,
  `privacy` ENUM('private', 'public') NOT NULL DEFAULT 'private',
  `userSignupStatus` ENUM('none', 'approval_required', 'auto_accept') NOT NULL DEFAULT 'auto_accept', 
  -- none = no user can signup; approval_required = admin must approve; auto_accept = users automatically accepted
  `plan` ENUM('free','basic','pro') NOT NULL DEFAULT 'free', -- the plan for the community, which essentially just places limits on what they can do
  `planPaidThrough` date NOT NULL DEFAULT '1970-01-01',
  `planDiscountPercent` int NOT NULL DEFAULT 0,
  `stripeChargeToken` varchar(32) NOT NULL DEFAULT '',
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
  `createdOn` datetime NOT NULL,
  `tokenType` enum('email','password_reset') NOT NULL DEFAULT 'email',
  PRIMARY KEY (`id`),
  KEY `userId` (`userId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `PrayerRequests` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `title` varchar(256) DEFAULT '',
  `body` text,
  `createdBy` int(11) NOT NULL,
  `privacy` enum('public','private') DEFAULT 'private',
  `dateCreated` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `status` enum('pending','answered','not_answered','unknown') DEFAULT 'pending',
  PRIMARY KEY (`id`),
  KEY `createdBy` (`createdBy`),
  KEY `privacy` (`privacy`),
  KEY `dateCreated` (`dateCreated`),
  KEY `createdBy_dateCreated` (`createdBy`,`dateCreated`)
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