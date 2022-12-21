# User Management System API

This is the documentation for the User Management System API. You can use this API to perform various actions such as login, logout, refresh token, get a list of users, get a specific user by ID, update a user, and delete a user.

The system follows the clean architecture principles, with a separation of concerns between the different layers. https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html


### API Documentation

https://www.postman.com/irvankadhafi/workspace/irvankadhafi-deall/collection/10454328-d1ea77ec-c419-41c4-b60f-1cc92955a17a?ctx=documentation


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