![Fruitbar Logo](fruitbar_logo.jpg)

What is Fruitbar?
=================
Fruitbar allows you to place and manage delicious orders of fruit.

Fruitbar is a **data entry web application** designed under a **micro-service architecture** using **Go** for the backend and **ReactJS** for the frontend. Under the hood, it uses **Docker** for containerization and **Postgres** for the database. It utilizes a **multi-stage build pipeline** with **Jenkins** as the CI/CD tool. It also provides scripts to generate **OpenAPI (2.0)** specs and **godoc**s.

**It is meant to be a sample project for my resume.**

Feature Highlights
------------------
- CRUD operations on user accounts, products, and orders
- User accounts with secure password storage
- Authorization via JWT
- Machine-readable logging
- A health check endpoint for every service
- Generate documentation with a single command
- Test, build, and deploy with a single command
- Integration with Jenkins for CI/CD
- ?% unit test coverage (>80%)
- Database pre-seeded with data

Architecture
============
![Architecture Diagram](fruitbar_archdiagram.jpg)

Services:
---------
- **Users API**: CRUD operations on user accounts, user authorization
- **Products API**: CRUD operations on products (i.e. the various fruits for sale)
- **Orders API**: CRUD operations on orders (i.e. an order of fruit)
- **Fruitbar UI**: Web interface
- **Database**: Data repository

API docs: [Link to OpenAPI 2.0 documentation]()
Thunderclient (built-in to MS visual studio/VS code) API tests: [Link to thunderclient tests]()

### Usage instructions
1. Create a new customer account for yourself (create user API or via web UI)
2. Log in using your new account (user API or web UI) OR use one of the pre-existing credentials:
	- username: **admin** password: **admin** (admin role)
	- username: **employee** password: **employee** (employee role)
	- username: **customer** password: **customer** (customer role)
3. Perform your desired operations (orders, products API or web UI)

#### Brief explanation of user roles:
- admin
	- perform any operation on users/products/orders
- employee
	- perform any operation on orders you own AND any customer order
	- perform any operation on your user account AND any customer user account
	- read the list of products
- customer
	- perform any operation on orders you own
	- perform any operation on your user account
	- read the list of products

Deployment
----------
The deployment is managed via Jenkins. (jenkins stuff here) The scripts themselves are in the Makefile.

### Usage instructions
- Run tests: `make test`
- Build the application: `make build`
- Run the application: `make run`
- Stop the application: `make stop`

Design Decisions
================
Key Decisions
-------------
- Write the backend in Golang: development in golang is fast, less error-prone than writing in a lower-level languages, and performant! Go is also popular enough that a replacement developer could be found if needed.
- Write the frontend in ReactJS: development is fast and it is popular-enough that a replacement developer could be found if needed.
- Use a JSON API for all operations following the [Google JSON Style Guide](https://google.github.io/styleguide/jsoncstyleguide.xml): Easy for consumers to deal with, flexible for developers to deal with, and best of all--human readable.
	- Does not implement the JSON PATCH API but supports partial updates via the 'fields' parameter in PUT requests. (http://jsonpatch.com/)
- Generate machine-readable logging in JSON: easy to plug it into your ELK stack.
	- Logs whenever there is HTTP traffic to any endpoint (track activity)
	- Logs whenever any database operation is performed (track database calls)
	- Logs whenever any operation on the data set is completed
	- Logs a few other informational calls (service startup, etc.)
	- (you don't want to log too much!)
- Use Gorm as the ORM system: reduces boilerplate and gives us a lot of easy and optimized databaase operations, with less code! (and no writing SQL)
- Implement a health check for each API service and the web UI: easy to plug it into your enterprise monitoring solution.

Room for improvement
--------------------
In the real world, sometimes features are simply out-of-scope, but it is still important to recognize what could be improved in the future.

### Features
- Generate OpenAPI 3.0 spec (it's the future)
- Audit trail with username/timestamp for operations on data
- Undelete capability utilizing "soft deletes" and a database stored procedure to archive (delete or move to another table) data that is too old

### Implementation details
- Use zerolog instead of logrus: better performance
- Use a more sophisticated permissions system instead of simple roles: easier for admins to customize how the system will be used.
- Implement the JSON PATCH API (perhaps using this still well-maintained json-patch library (https://github.com/evanphx/json-patch)): a more complete and maintainable way of performing partial updates to data
- Use (forgot name) to automatically generate JSON annotation values instead of relying on hardcoding them or writing custom Marshal functions: more maintainable in the long-term
- Implement multi-threading in the HTTP handlers: improved performance
- Use Go generics (VERY recently released in Go 1.18): reduced boilerplate in the code