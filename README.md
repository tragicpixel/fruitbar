![Fruitbar Logo](fruitbar_logo.jpg)

Overview
=========
Synopsis
--------
Fruitbar is a multi-stage containerized web application written in Go and ReactJS.

The nitty gritty: It's a microservice CRUD application with user authentication.

*It's meant as a sample project to be viewed by potential employers.*

- **Back-end microservices: Go**
- **Front-end UI: ReactJS**
- **Major libraries used: gorillamux** and **gorm** (Go), as well as **react-router-dom** (ReactJS)
- 80% unit test coverage
- All code is formatted using godoc and golint
- OpenAPI (2.0) documentation is provided for each service endpoint
- A **postgres** database is used, but is easily changed in code, and an alternative "file" repository is provided if you want to try it out
- Support for user accounts with secure password storage.
- Machine-readable logging.
- **Docker** is used for containment, and a complete **Kubernetes manifest** is provided
- **The entire application can be run, built, and viewed with a single command from a single Makefile**
- A **Jenkinsfile** is provided for CI/CD integration.

![Screenshot of the cart page](fruitbar_cart.jpg)

Use cases
---------
Fruitbar use cases:
- Place an order of fruit
- Manage existing orders
- Manage the inventory of fruit
- Manage user accounts for employees, customers, and admins.

![Screenshot of the inventory page](fruitbar_inventory.jpg)

Architecture
------------
There are four services:
- 🧑‍🤝‍🧑 **Users** service : Register new user accounts, authenticate existing accounts, authorize user tokens.
- 🛍️ **Orders** service : Create/Read/Update/Delete customer orders.
- 🍓 **Products** service : Create/Read/Update/Delete products available for the customer to purchase. (product menu)
- 🖥️ **User Interface** service: User interface for the application.

*🩺 All services have a health check endpoint.*

![Architecture Diagram](fruitbar_archdiagram.jpg)

Sending requests to the API
---------------------------
There are three APIs. All three are JSON APIs:
- **/orders/** : CRUD operations for orders placed by customers.
- **/users/** : register new user, log in, and authenticate user tokens when an API request is made.
- **/products/** : CRUD operations for inventory items.

All APIs have the following endpoint:
- **/health/** : Health check.

Before you can make requests to the API, you to need to register a user account and log in to recieve your authentication token (JSON Web Token):

- **/users/register** : register new user.
- **/users/login** : login and recieve an authentication token.
- **/users/authorization** : All requests to the other APIs are routed through this one. It checks that an authentication token is valid.

Once you have a valid token, you can use the APIs as you see fit!

Check the API documentation for more information. (see the [Quickstart section](#quickstart))

Navigating the application in a web browser
-------------------------------------------
Once the application is up and running, navigate to [localhost:xxxx](http://localhost:3320) to view it.

There are five pages:
- **/login/**: Log in to your user account.
- **/cart/** : Place an order of fruit. (requires Customer role)
- **/dashboard/** : Manage existing orders. (requires Employee or Admin role)
- **/inventory/** : Manage the fruit that is available to buy. (requires Employee or Admin role)
- **/users/** : Manage the existing user accounts. (requires Admin role)

Upon initially visiting any page of the site, or the root directory, if you are not logged in, you will be routed to the login page.

Upon successful login, you will be routed to the page you originally visited, or the cart page for customers, and the dashboard page for employees or admins.

Visiting a page your role is not authorized to view will show an error message, but you can navigate away using the navigation bar.

Logging out will send you back to the login page.

Dev-ops: Testing, Building, and Deploying
-----------------------------------------
...

Quickstart
==========
Build & run the application
---------------------------
1. Download the project.
2. Navigate to the top level directory of the project. (x/x)
3. Run the following command:

`build, run, open a window to the main page`

View the API documentation
--------------------------
1. Navigate to the top level directory of the project in your shell.
2. Run the following command:

`make swagger`

Pre-loaded user credentials
---------------------------
If you want to get right into the action, the following user credentials have been pre-loaded on the database image:

- Customer role -- **username**: customer **password:** fruitbar
- Employee role -- **username**: employee **password:** fruitbar2
- Admin role -- **username**: admin **password:** admin

Design Decisions
================
Third Party Libraries
---------------------
- **gorillamux**: HTTP routing.
- **logrus**: logging.
- **jwt-go**: generate and decode JSON Web Tokens.
- **gorm**: object-relational management against database.
- **zapgorm2**: structured logging for gorm.
- **go-swagger**: generate OpenAPI documentation.
- **create-react-app**: generate boilerplate for the front-end.
- **react-router-dom**: HTTP routing for the front-end.

Model/Repository/Handler/Driver Structure
-----------------------------------------
- `Model` models the data. `Repository` and `Handler` depend on this implementation.
- `Repository` represents any persistent storage of the data (could be database, file, etc.). It implements the repository operations using different kinds of databases/ORMs.
- `Handler` handles any operations done on the data. It doesn't depend on the implementation details of the `Repository`.
- `Driver` handles the database connection/setup for different types of databases/ORMs.

You can easily switch out which kind of underlying database/ORM is being used, without modifying any of the internal handler logic--refactoring to use a different underlying database implementation would be a breeze.

Services defined as types, separate main.go for each service
------------------------------------------------------------
- Keeps all configuration, setup, and health check logic for a service contained in one location.
- Keeps main.go files a lot more concise and readable -- you can worry about how the application is started here, other concerns are in the service type.
- Separate main.go makes it easier to define targets when building each service.

JSON API
--------
- The API only returns JSON or an empty page: easy for consumer applications to deal with.
- Responses in JSON follow the [Google JSON Style Guide](https://google.github.io/styleguide/jsoncstyleguide.xml): provides a standardized style guide for JSON responses that consumers and developers can use.
- Health check uses a simpler format: you only need to ever return one piece of data for the health check ("ok":"yes").
- Does not implement the JSON PATCH API but supports partial updates via the 'fields' parameter in PUT requests. (http://jsonpatch.com/)

Logging
-------
- Using `logrus` go module: Chosen for ease of use. `zerolog` might be a better pick in the real-world if logging performance is a concern.
- Using `zapgorm2` go module: Chosen to give structured logging from Gorm. (why re-invent the wheel when your true innovation is creating the tire?)
- Logs are writting in the [JSON Lines](http://jsonlines.org) format: works well with log aggregators i.e. elasticsearch in most real-world implementations.
- Option to switch between JSON and text formats: JSON can be hard to read for a human, easier to examine the logs when developing/debugging on a local machine with no log aggregation.
- Log level is set based on the environment: Adjustable to reduce logging volume in production environments where there may be heavy traffic.

### What is being logged:
- Logs **all traffic to HTTP endpoints**: Useful for gathering statistics, debugging, security, etc. -- you should have this in your CRUD app.
- Logs when **any database operation is performed** (attempted,completed,etc): Useful for gathering statistics, debugging.
- Logs when **any CRUD operation is performed** (attempted,completed,etc): Useful for gathering statistics, debugging.
- Logs when **any service moves to a different phase in its application life cycle**: Set up, operation, shutting down, etc. -- useful for debugging, good to know.

Using ReactJS as the front-end framework
----------------------------------------
- Widely used front-end framework, so other developers can easily understand it
- All the front-end code will be in javascript.

Single makefile, dockerfile, multi-stage builds
-----------------------------------------------
- Uses a single makefile and dockerfile: Keep everything in one spot for the entire application. Easy to notice differences in the builds for each, easy to keep things straight for each service when refactoring the build.
- Multi-stage builds are used for all the services: Build with different options for dev, qa, and prod environments.
- Any given build must pass testing in order to be built: Prevents bugs and regression.

### k8s manifest included
- Kubernetes is widely used tool for managing containers, so can easily to plug into an existing kubernetes cluster.

### Jenkinsfile included
- It's a widely used open source tool for CI/CD, so can easily plug in to an existing Jenkins implementation.

Room for improvement
--------------------
I recognize that the following things could be improved, and should be, for a real-world project, but are out-of-scope for this sample project:
- Using Open API 3.0 specification + documentation generators
- Set the services up as a swarm so I can use docker secrets to hide passwords
- Use zerolog instead of logrus: better log performance.
- Add a register screen for users. (currently must use the API or directly edit the database)
- Use generics (very recently introduced in go 1.18) to cut down on some of the boilerplate (service, handler, repository packages)
- Add an audit trail with username + timestamp, for create/update/delete operations on orders and products.
- Use the "soft delete" feature of gorm in combination with a scheduled stored procedure to archive "deleted" orders/users/products, and use another stored procedure to delete archived items after some time. (in compliance with state laws for example)
- Use a permissions set up for user roles, instead of strictly defined roles.
- Dynamically retrieve sales tax based on the zip code in the payment information. (although, this could depend on the tax laws of the state the server resides in)
- Implement the JSON PATCH API (partial updates are already supported via 'fields' parameter to PUT requests, but implementing the official standard would be more ideal, and also, more work) perhaps using the still maintained json-patch library (https://github.com/evanphx/json-patch)