# User Management System API

This is the documentation for the User Management System API. You can use this API to perform various actions such as login, logout, refresh token, get a list of users, get a specific user by ID, update a user, and delete a user.

The system follows the clean architecture principles, with a separation of concerns between the different layers. https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html

<p align="center">
  <img src="./diagram.png" />
</p>

## Steps
1. Clone the repository: `git clone git@github.com:irvankadhafi/user-management-system.git`
2. Run `docker-compose up`
3. Migrate database: `cd user-service && go run . migrate`
4. Seed user data `cd user-service && go run . seeder`
5. Seed rbac `cd user-service && go run . go run . migrate-rbac-permission`
6. Run server `make run`

## TODO
- Implement all unit tests
- Decomposing rbac mechanism to new service named auth-service
- Using MongoDB as database
- Create kubernetes deployment config

## API Documentation

https://www.postman.com/irvankadhafi/workspace/irvankadhafi-deall/collection/10454328-d1ea77ec-c419-41c4-b60f-1cc92955a17a?ctx=documentation

## User Login Credential (After seed)
- Admin:
    ```
    email: irvankadhafi@mail.com
    password: 123456 
    ```
- Member
     ```
    email: johndoe@mail.com
    password: 123456 
    ```


## Technology
The following technologies were used to build this system:
- Golang for the backend server
- PostgreSQL for the database
- Redis for cache

## Features
- Login user with email and password
- Logout user
- Refresh token
- Admin and member roles, with different levels of access to CRUD functionality
- Redis cache to improve performance
- Logging for easier debugging