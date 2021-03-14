# Pregxas API

NOTE: Not currently production-ready.

Pregxas is an open-source community-based prayer management system. There is the main Pregxas application, hosted by Treelight Software, but you are free to take the server and applications (or build your own clients!) and use it to keep track of prayer plans, prayer requests, and more for your group, regardless of faith or background (as subject to the LICENSE.md terms). The application assumes it is the host with smaller communities within the site. If you don't want to host your own server, you may sign up for a free or paid hosted version with us!

*What does pregxas mean?*

Pregxas (preh-jas) is a verb from Esperanto which means to pray in the present tense. For example, "mi pregxas" (I pray), "ni pregxas" (we pray).

## Why Not Fully Open Source?

We at Treelight Software value free and open sharing of knowledge. However, we also recognize that it would be entirely possible, and even ethical, for a competitor to take our hard work and offer it as a service under their brand. This is a challenge, and one that has, unfortunately, become more prevalent. After a lot of thought and "soul searching", we decided on a compromise. The code base remains open for contributions but cannot be resold. We allow organizations to host the platform under very narrow rules, related to size and use case.

If you have concerns, want to discuss specific licensing, or simply want to talk about it, feel free to contact us at support@treelightsoftware.com

## Status

The current version of Pregxas is written in NodeJS. Treelight Software, the new stewards of the platform, is rewriting the application in Go for a wide variety of reasons, including compilation, speed, and skillset. Individual aspects of the existing software will be migrated as soon as possible. Feel free to help contribute with documentation, tests, features, and bug fixes!

**This software is not currently ready for use in production and will have many breaking changes as it is rebuilt**

## Requirements

The API is written in Go (as of version 2) and compiles to a static binary. It requires the following additional tools or technologies to function:

- MySQL or Compatible (MariaDB, AWS Aurora, etc)

- Mailgun for emails

- [Task](https://github.com/go-task/task) for cross-platform task management (needed only if working with the code directly)

- [Migrate](https://github.com/golang-migrate/migrate) for database migrations (needed only if working with the code directly and you don't want to run the SQL scripts directly)

### Other Services

We run the following additional services directly in Docker containers, but you are welcome to run them however you would like. However, Pregxas requires access to:

- Redis - Used for caching (not yet implemented)

- RabbitMQ - Used for message queuing (not yet implemented)

- Sonic - Used for searches. Configuration files are found in the `sonic/` directory. Feel free to tailor for your use. To run locally, for example:

  - `docker run -p 1491:1491 -v $PWD/sonic/config.cfg:/etc/sonic.cfg -v $PWD/sonic/store:/var/lib/sonic/store valeriansaliou/sonic:v1.3.0`

## Integration

Each instance can function completely independently. You can use our (soon to be released) React application or you can build your own client.

*What's to stop me from running my own host platform?*

As long as it meets the requirements in the license (for example, a small prayer group), nothing at all. If you want to run your own host platform and handle sub-communities, go for it! It may be a little more complicated to integrate with us, but if it helps you solve a need, then go for it!

## Running

The easiest way to run the application is to use the Docker image (in the process of being migrated). Until that Docker image is migrated, you will simply want to clone the repository, make your changes, run the tests, and open a PR. A good example is the `docker-compose.yml` file, which shows the services needed.

### Vendor

We commit our vendor directory. That way we con be confident in builds, regardless of network status.

## Authentication and Authorization

Pregxas will support two authentication mechanisms. The first is session-based access/refresh with local login. The access and refresh tokens will be returned after logon view Secure, HTTPOnly cookies to prevent the need for JS to interact with them or store them (such as in local storage, which has security implications). However, the tokens are still returned in the body for non-web applications, such as mobile.

On the roadmap is support for OAuth2 to allow third-party integration as well. This is to be determined.

## Basic Concepts

The `Site` is the single installation. If you are running this on your own, you would configure the site to be however you would like. When the server starts, it will check to see if the Site has been configured. If not, it will generate a passcode that will be used for setting up the Site and configuring it.

Each Site can have several `Communities`. Communities can be free or paid subscriptions (a future plan will allow for toggling plan thresholds and hiding subscription options). Users can request to join communities and admins may invite users to communities.

A public community is listed publicly but has several options for handling membership. Auto approval allows anyone to join. Admins can also choose to only approve requests manually. A third option exists and is listed below.

A private community is not listed in the public feed. It can still be joined by sending invitations. Users cannot directly request to join a private community; they must be invited.

Both private and public communities can be set up with a `shortCode`. This `shortCode` allows a user to join a community with a specified code. If the user requests to join with the correct code, they are automatically joined to the group as a member.

`Prayer Requests` are single requests for prayers. They are made by a user and can then be joined to specific communities. If the request is marked `private` it will NOT show up in the global feed.

Users may add `prayers` to a request. These are only allowed once within a sliding time window. An email may optionally be sent with a list of Prayer Requests prayed for and updates.

## I'm New, How Can I Help

There are many ways to help, and we are a very open group that has no problem helping along newer contributors, whether new to the industry or just new to Go!

If you don't have any programming background, that's fine! We could always use help with our documentation (both comments and Open API Standard 3 docs) or with new tutorials or guides. If you create an external resource, such as a tutorial or guide, we would love to link to it from here!

If you are experienced in Quality Assurance, we would love any feedback on bugs, typos, or things that could be improved. You can open a ticket or simply email us. Whatever works for you!

If you are a programmer with Go experience, we would love to have your help with this project. If you aren't sure where we need the most help, and the To Do list below seems daunting, you can always help with tests, clarifications, and QA.

Speak multiple languages? Although the API is mostly neutral in that regard, our web and mobile apps could always use a translation. Since none exist at this point, email us and we can talk about the best path forward here!

## TODO and New Features

An `X` represents a recently released feature that could use some testing but is, in theory, completed. A `-` represents something that is in progress but will likely span many different pull requests (such as adding OAS3 docs for every endpoint).

[X] Finish migrating Prayer Requests and Prayers

[ ] Implement emails and administrative reports

[X] Implement Prayer Lists

[ ] Implement subscriptions for communities

[X] Implement reporting a prayer

[-] Create OpenAPI 3 compliant docs for all end points (see the [API Docs Repo](https://github.com/TreelightSoftware/pregxas-api-docs) folder and the `Config.go` file for `TODO` stubs on which endpoints need it)

[ ] Build an administrative dashboard

[ ] Migrate web application

[ ] Migrate mobile application
