# Pregxas API

Pregxas is an open-source community-based prayer management system. There is the main Pregxas application, hosted by Treelight Software, but you are free to take the server and applications (or build your own clients!) and use it to keep track of prayer plans, prayer requests, and more for your group, regardless of faith or background. The application assumes it is the host with smaller communities within the site. If you don't want to host your own server, you may signup for a free or paid hosted version with us!

*What does pregxas mean?*

Pregxas (preh-jas) is a verb from Esperanto which means to pray in the present tense. For example, "mi pregxas" (I pray), "ni pregxas" (we pray).

## Status

The current version of Pregxas is written in NodeJS. Treelight Software, the new stewards of the platform, is rewriting the application in Go for a wide variety of reasons, including compilation, speed, and skillset. Individual aspects of the existing software will be migrated as possible and immediately open-sourced. Feel free to help contribute with documentation, tests, features, and bug fixes!

**This software is not currently ready for use in production and will have many breaking changes as it is rebuilt***

## Requirements

The API is written in Go (as of version 2) and compiles to a static binary. It requires the following additional tools or technologies to function:

- MySQL or Compatible (MariaDB, AWS Aurora, etc)

- Mailgun for emails (free plan available)

- [Task](https://github.com/go-task/task) for cross-platform task management (needed only if working with the code directly)

- [Migrate](https://github.com/golang-migrate/migrate) for database migrations (needed only if working with the code directly and you don't want to run the SQL scripts directly)

## Integration

Each instance can function completely independently. You can use our (soon to be released) React application or you can build your own client.

*What's to stop me from running my own host platform?*

Nothing at all. If you want to run your own host platform and handle sub-communities, go for it! It may be a little more complicated to integrate with us, but if it helps you solve a need, then go for it!

## Running

The easiest way to run the application is to use the Docker image (in the process of being migrated). Until that Docker image is migrated, you will simply want to clone the repository, make your changes, run the tests, and open a PR.

### Vendor

We commit our vendor directory. That way we con be confident in builds, regardless of network status.

## Basic Concepts

The `Site` is the single installation. If you are running this on your own, you would configure the site to be however you would like. When the server starts, it will check to see if the Site has been configured. If not, it will generate a passcode that will be used for setting up the Site and configuring it.

Each Site can have several `Communities`. Communities can be free or paid subscriptions (a future plan will allow for toggling plan thresholds and hiding subscription options). Users can request to join communities and admins may invite users to communities.

`Prayer Requests` are single requests for prayers. They are made by a user and can then be joined to specific communities. If the request is marked `private` it will NOT show up in the global feed.

Users may add `prayers` to a request. These are only allowed once within a sliding time window. An email may optionally be sent with a list of Prayer Requests prayed for and updates.

## TODO and New Features

[X] Finish migrating Prayer Requests and Prayers
[ ] Implement emails and reports
[ ] Implement subscriptions for communities
[ ] Implement reporting a prayer
[ ] Build an administrative dashboard
[ ] Migrate web application
[ ] Migrate mobile application
