# auth-go

This is a Go-based web application that provides user authentication and management services. It uses the Gin framework for routing and handling HTTP requests, and PostgreSQL for the database.

## Features

- **User Authentication:**
  - Register new users
  - Login with email and password
  - Logout
  - Password reset
  - Email verification
- **OAuth 2.0:**
  - Login with third-party providers (e.g., Google, Github)
- **User Management:**
  - Get user information
  - Update user information
  - Delete users
- **JWT Support:**
  - Uses JSON Web Tokens for secure API authentication

## Getting Started

### Prerequisites

- Go 1.16+
- PostgreSQL
- Git

### Installation

1.  Clone the repository:
    ```sh
    git clone https://github.com/Jonathan0823/auth-go.git
    ```
2.  Install dependencies:
    ```sh
    go mod tidy
    ```
3.  Set up the database:
    - Create a PostgreSQL database
    - Set the environment variables in a `.env` file (see Configuration section)
4.  Run the application:
    ```sh
    go run main.go
    ```

## Usage

The application exposes a RESTful API for user authentication and management.

### API Endpoints

- `POST /api/auth/register`: Register a new user
- `POST /api/auth/login`: Login with email and password
- `POST /api/auth/logout`: Logout the current user
- `POST /api/auth/refresh`: Refresh the JWT token
- `POST /api/auth/forgot-password`: Request a password reset
- `POST /api/auth/reset-password`: Reset the password
- `GET /api/auth/verify/email`: Verify the user's email
- `POST /api/auth/verify/email/resend`: Resend the email verification link
- `GET /api/auth/:provider`: Initiate OAuth 2.0 login with a provider
- `GET /api/auth/:provider/callback`: Handle the OAuth 2.0 callback
- `GET /api/user/me`: Get the current user's information
- `GET /api/user/:id`: Get user information by ID
- `GET /api/user/get-all`: Get all users
- `GET /api/user/email`: Get user information by email
- `PATCH /api/user/update`: Update the current user's information
- `DELETE /api/user/delete/:id`: Delete a user by ID

## Configuration

The application is configured using environment variables. Create a `.env` file in the root of the project with the following variables:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
DB_SSL=disable

JWT_SECRET=your_jwt_secret
JWT_REFRESH_SECRET=your_jwt_refresh_secret

GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback

GITHUB=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
GITHUB_REDIRECT_URL=http://localhost:8080/api/auth/github/callback

EMAIL_HOST=your_email_host
EMAIL_PORT=your_email_port
EMAIL_USER=your_email_user
EMAIL_PASSWORD=your_email_password
```

## Dependencies

- [Gin](https://github.com/gin-gonic/gin): HTTP web framework
- [pq](https://github.com/lib/pq): PostgreSQL driver
- [jwt-go](https://github.com/golang-jwt/jwt): JSON Web Token implementation
- [goth](https://github.com/markbates/goth): OAuth 2.0 library
- [godotenv](https://github.com/joho/godotenv): Environment variable loader
- [validator](https://github.com/go-playground/validator): Input validation
- [sessions](https://github.com/gorilla/sessions): Session management
- [gomail](https://github.com/go-gomail/gomail): Email sending

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
