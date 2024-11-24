```
/secretsanta
├── main.go
├── go.mod
├── go.sum
├── config
│   └── config.go
├── controllers
│   ├── auth_controller.go
│   ├── group_controller.go
│   └── user_controller.go
├── models
│   ├── group.go
│   ├── participant.go
│   └── user.go
├── routes
│   └── routes.go
├── utils
│   ├── hash.go
│   └── jwt.go
└── database
    └── database.go
```

- **main.go**: The entry point of the application. It initializes the server and routes.
- **go.mod** and **go.sum**: Go modules files for dependency management.
- **config/config.go**: Configuration settings for the application.
- **controllers/**: Contains the handler functions for different endpoints.
  - **auth_controller.go**: Handles authentication-related endpoints (e.g., sign in, sign up).
  - **group_controller.go**: Handles group-related endpoints (e.g., create group, add participant).
  - **user_controller.go**: Handles user-related endpoints (e.g., create user, get user).
- **models/**: Contains the data models for the application.
  - **group.go**: Defines the Group model.
  - **participant.go**: Defines the Participant model.
  - **user.go**: Defines the User model.
- **routes/routes.go**: Defines the routes and associates them with the corresponding handlers.
- **utils/**: Contains utility functions.
  - **hash.go**: Functions for hashing passwords.
  - **jwt.go**: Functions for generating and validating JWT tokens.
- **database/database.go**: Contains the database connection and initialization logic.







