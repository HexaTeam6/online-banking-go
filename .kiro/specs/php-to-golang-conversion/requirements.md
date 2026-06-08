# Requirements Document

## Introduction

This document defines the requirements for converting the existing PHP online banking application to a modern Go (Golang) REST API backend. The conversion replaces the legacy PHP monolith (which renders HTML directly) with a clean-architecture Go service exposing RESTful endpoints. The new system addresses critical security vulnerabilities (plaintext passwords, SQL injection, no CSRF protection) and introduces proper authentication, input validation, structured logging, and testable architecture following SOLID principles.

## Glossary

- **API_Server**: The Go HTTP server that exposes RESTful endpoints for the banking application
- **Auth_Service**: The service responsible for customer and admin authentication, session/token management
- **Account_Service**: The service responsible for managing customer bank accounts and account types
- **Transfer_Service**: The service responsible for processing money transfers between accounts
- **Request_Service**: The service responsible for managing money request operations between customers
- **Transaction_Repository**: The data access layer for transaction records
- **Customer_Repository**: The data access layer for customer profile data
- **Admin_Repository**: The data access layer for admin user data
- **Password_Hasher**: The component responsible for hashing and verifying passwords using bcrypt
- **Input_Validator**: The component responsible for validating and sanitizing all external input
- **Token_Manager**: The component responsible for issuing and validating JWT authentication tokens
- **Database**: The MySQL/MariaDB relational database storing all banking data
- **Account_Number**: A unique 9-digit integer identifier for each customer bank account
- **Quick_Transfer**: A money transfer operation between two accounts with amount limits of 500 to 20,000

## Requirements

### Requirement 1: Customer Authentication

**User Story:** As a customer, I want to securely log in to my bank account, so that I can access my banking features with confidence that my credentials are protected.

#### Acceptance Criteria

1. WHEN a customer submits a valid username (between 3 and 64 characters) and a valid password (between 8 and 128 characters) that matches the bcrypt hash stored in the Database, THE Auth_Service SHALL return a JWT access token with an expiration time of 15 minutes
2. WHEN a customer submits invalid credentials (non-matching username or password), THE Auth_Service SHALL return an authentication error indicating that the credentials are invalid, without revealing which specific field (username or password) is incorrect
3. WHEN a customer logs in successfully, THE Auth_Service SHALL record the login timestamp and client IP address in the login history
4. WHEN a customer logs out, THE Auth_Service SHALL invalidate the current token and record the logout timestamp in the login history
5. IF an authentication request is missing the username or password field, THEN THE Input_Validator SHALL return a validation error specifying the missing fields
6. THE Password_Hasher SHALL store all new passwords using bcrypt with a minimum cost factor of 10
7. IF a customer submits invalid credentials 5 consecutive times within a 15-minute window, THEN THE Auth_Service SHALL lock the account for 30 minutes and return an error indicating the account is temporarily locked
8. IF the Auth_Service receives a request with an expired or invalidated JWT token, THEN THE Auth_Service SHALL return an authentication error indicating the session has expired

### Requirement 2: Customer Registration

**User Story:** As a new user, I want to register for an online banking account, so that I can begin using the banking services.

#### Acceptance Criteria

1. WHEN a customer submits a registration form with all required fields (first name, last name, gender, birth date, mobile, email, address, city, state, zip code, username, password, and account type), THE Account_Service SHALL create records in the account, customer, address, account_type, and balance tables within a single database transaction
2. WHEN a customer registers, THE Password_Hasher SHALL hash the password using bcrypt with a minimum cost factor of 10 before storage
3. IF a registration request uses a username that already exists, THEN THE Account_Service SHALL return a conflict error indicating the username is taken
4. IF a registration request contains an email that does not conform to a valid email address format, THEN THE Input_Validator SHALL return a validation error identifying the email field
5. IF a registration request contains a mobile number that is not between 10 and 15 digits, THEN THE Input_Validator SHALL return a validation error identifying the mobile field
6. THE Account_Service SHALL generate a unique 9-digit account number for each new registration
7. THE Account_Service SHALL initialize the new account balance to zero
8. IF a registration request is missing any required field or contains a password shorter than 8 characters or a username shorter than 3 characters or longer than 30 characters, THEN THE Input_Validator SHALL return a validation error specifying each invalid field
9. IF a registration request contains an account type other than "SAVING" or "CURRENT", THEN THE Input_Validator SHALL return a validation error indicating the allowed account types
10. IF the database transaction fails during registration, THEN THE Account_Service SHALL roll back all changes and return an error indicating registration could not be completed

### Requirement 3: Customer Dashboard

**User Story:** As a customer, I want to view my transaction summary and account balance, so that I can monitor my financial activity.

#### Acceptance Criteria

1. WHEN an authenticated customer requests the dashboard, THE API_Server SHALL return the all-time total number of transactions, total credit amount, total debit amount, and current balance for that customer's account
2. WHEN an authenticated customer requests the dashboard, THE API_Server SHALL return monthly statistics including transaction count, credit sum, and debit sum for the current calendar month
3. WHEN an authenticated customer requests transaction history, THE API_Server SHALL return a paginated list of transactions ordered by date descending, with a default page size of 10 and a maximum page size of 50, where each transaction entry includes the transaction type, amount, purpose, date, and resulting balance
4. IF an unauthenticated request is made to the dashboard endpoint, THEN THE API_Server SHALL return a 401 Unauthorized response
5. IF the authenticated customer has no transactions, THEN THE API_Server SHALL return zero for all summary totals, the current balance, and an empty transaction list
6. IF the page number requested exceeds the total available pages, THEN THE API_Server SHALL return an empty transaction list with correct pagination metadata indicating zero results for that page

### Requirement 4: Quick Transfer

**User Story:** As a customer, I want to transfer money to another account quickly, so that I can send funds to other bank customers.

#### Acceptance Criteria

1. WHEN an authenticated customer submits a transfer request containing a destination account number, transfer amount, and purpose (maximum 100 characters), THE Transfer_Service SHALL debit the sender account, credit the receiver account, update both account balances, and record both transactions atomically within a single database transaction
2. IF the transfer amount is less than 500 or greater than 20,000, THEN THE Transfer_Service SHALL reject the transfer with a limit violation error
3. IF the destination account number does not exist in the Database, THEN THE Transfer_Service SHALL reject the transfer with an invalid account error
4. IF the sender account number matches the destination account number, THEN THE Transfer_Service SHALL reject the transfer with a self-transfer error
5. IF the sender account balance is less than the transfer amount, THEN THE Transfer_Service SHALL reject the transfer with an insufficient balance error
6. IF the destination account number is not a valid 9-digit integer, THEN THE Transfer_Service SHALL reject the transfer with a validation error indicating invalid account number format
7. THE Transfer_Service SHALL record the transaction with type, amount, purpose, date, and updated balance for both accounts
8. IF the database transaction fails during transfer processing, THEN THE Transfer_Service SHALL roll back all changes and return an error indicating the transfer could not be completed
9. IF an unauthenticated request is made to the transfer endpoint, THEN THE API_Server SHALL return a 401 Unauthorized response

### Requirement 5: Money Request

**User Story:** As a customer, I want to request money from another account holder, so that I can receive funds from other bank customers.

#### Acceptance Criteria

1. WHEN a customer submits a money request with a valid target account number, an amount between 500 and 20,000 inclusive, and a message between 1 and 200 characters, THE Request_Service SHALL create a request record with status "PENDING", the request amount, message, requester account number, target account number, and timestamp
2. IF the request amount is less than 500 or greater than 20,000, THEN THE Request_Service SHALL reject the request with a limit violation error
3. IF the target account number does not exist in the Database, THEN THE Request_Service SHALL reject the request with an invalid account error
4. IF the requester account number matches the target account number, THEN THE Request_Service SHALL reject the request with a self-request error
5. IF the message field is empty or exceeds 200 characters, THEN THE Request_Service SHALL reject the request with a validation error indicating the message constraint
6. WHEN a customer views received requests, THE Request_Service SHALL return a paginated list of requests sent to that customer ordered by date descending, including request id, requester account number, amount, message, status, hasViewed flag, and timestamp
7. WHEN a customer views a request, THE Request_Service SHALL mark the request as viewed by setting the hasViewed flag to true
8. IF an unauthenticated request is made to any money request endpoint, THEN THE API_Server SHALL return a 401 Unauthorized response

### Requirement 6: Customer Profile Management

**User Story:** As a customer, I want to view and manage my profile information, so that I can keep my personal details up to date.

#### Acceptance Criteria

1. WHEN an authenticated customer requests their profile, THE API_Server SHALL return the customer details including full name, gender, birth date, mobile, email, address, and account type
2. WHEN an authenticated customer submits a profile update, THE API_Server SHALL accept changes only to the following fields: full name, gender, birth date, mobile, email, and address, and SHALL persist the validated updates to the Database
3. IF a profile update contains invalid data, THEN THE Input_Validator SHALL return validation errors identifying each invalid field and the reason for rejection, applying the following rules: full name must be 1 to 100 characters, email must match a valid email format, mobile must be 10 to 15 digits, birth date must be a valid past date, and gender must be one of "Male", "Female", or "Other"
4. IF an unauthenticated request is made to the profile endpoint, THEN THE API_Server SHALL return a 401 Unauthorized response
5. IF a profile update request contains no updatable fields, THEN THE API_Server SHALL return a validation error indicating that at least one field must be provided

### Requirement 7: Customer Feedback

**User Story:** As a customer, I want to submit feedback about the banking service, so that I can share my experience with the bank.

#### Acceptance Criteria

1. WHEN an authenticated customer submits feedback with text between 1 and 1000 characters and a rating between 1 and 5, THE API_Server SHALL store the feedback text, rating (hearts), and submission timestamp linked to the customer account
2. IF the feedback text is empty or exceeds 1000 characters, THEN THE Input_Validator SHALL return a validation error indicating the text length constraint
3. IF the rating value is not an integer between 1 and 5 inclusive, THEN THE Input_Validator SHALL return a validation error indicating the allowed rating range
4. IF an unauthenticated request is made to the feedback endpoint, THEN THE API_Server SHALL return a 401 Unauthorized response

### Requirement 8: Admin Authentication

**User Story:** As a bank administrator, I want to securely log in to the admin dashboard, so that I can manage bank operations.

#### Acceptance Criteria

1. WHEN an admin submits valid credentials, THE Auth_Service SHALL verify the password against the bcrypt hash stored in the Database and return a JWT token with admin role claims and an expiration time of 30 minutes
2. WHEN an admin submits invalid credentials, THE Auth_Service SHALL return an authentication error without revealing which field is incorrect
3. THE Token_Manager SHALL include role-based claims in admin tokens that distinguish admin users from customers
4. IF a customer token is used to access an admin endpoint, THEN THE API_Server SHALL return a 403 Forbidden response
5. IF an admin authentication request is missing required fields, THEN THE Input_Validator SHALL return a validation error specifying the missing fields
6. IF an admin JWT token has exceeded its expiration time, THEN THE API_Server SHALL return a 401 Unauthorized response

### Requirement 9: Admin Account Management

**User Story:** As a bank administrator, I want to manage customer accounts and balances, so that I can support banking operations.

#### Acceptance Criteria

1. WHEN an admin requests the customer list, THE API_Server SHALL return a paginated list of all customer profiles with a default page size of 20 and a maximum page size of 100
2. WHEN an admin submits a balance adjustment with a valid account number, operation type (credit or debit), and an amount between 0.01 and 999,999,999.99, THE API_Server SHALL update the customer balance, record a transaction entry with type, amount, purpose, date, and updated balance, and persist the changes atomically
3. IF an admin submits a balance adjustment with an account number that does not exist in the Database, THEN THE API_Server SHALL return an invalid account error without modifying any balance
4. IF an admin submits a debit balance adjustment and the customer account balance is less than the specified amount, THEN THE API_Server SHALL return an insufficient balance error without modifying any balance
5. WHEN an admin views transactions, THE API_Server SHALL return a paginated list of all transactions across all accounts with a default page size of 20 and a maximum page size of 100
6. WHEN an admin views money requests, THE API_Server SHALL return a paginated list of all requests across all accounts with a default page size of 20 and a maximum page size of 100
7. WHEN an admin views feedback, THE API_Server SHALL return a paginated list of all customer feedback entries with a default page size of 20 and a maximum page size of 100
8. WHEN an admin views login history, THE API_Server SHALL return a paginated list of login/logout records with a default page size of 20 and a maximum page size of 100

### Requirement 10: Input Validation and Security

**User Story:** As a system operator, I want all inputs validated and queries parameterized, so that the system is protected against injection attacks and malformed data.

#### Acceptance Criteria

1. THE API_Server SHALL use parameterized queries for all database operations
2. THE Input_Validator SHALL validate all request body fields against defined type and length constraints before processing, enforcing a maximum of 255 characters for general string fields and 1000 characters for text fields unless a field-specific limit is defined elsewhere
3. IF a request has a content type other than application/json, THEN THE API_Server SHALL return a 415 Unsupported Media Type response without processing the request body
4. THE API_Server SHALL apply rate limiting to authentication endpoints, permitting a maximum of 5 requests per 60-second window per client IP address
5. IF the rate limit is exceeded on an authentication endpoint, THEN THE API_Server SHALL return a 429 Too Many Requests response with a Retry-After header indicating the number of seconds until the next request is permitted
6. IF a request body exceeds 1 MB in size, THEN THE API_Server SHALL return a 413 Payload Too Large response without processing the request body
7. THE API_Server SHALL include CORS headers configurable via environment variables, supporting at minimum Access-Control-Allow-Origin, Access-Control-Allow-Methods, and Access-Control-Allow-Headers

### Requirement 11: Application Configuration and Infrastructure

**User Story:** As a DevOps engineer, I want the application configured via environment variables with structured logging, so that I can deploy and monitor the system reliably.

#### Acceptance Criteria

1. THE API_Server SHALL read all configuration values from environment variables including database connection string, JWT secret, server port, and CORS origins
2. IF a required environment variable (database connection string, JWT secret) is missing or empty at startup, THEN THE API_Server SHALL log an error message indicating the missing variable name and exit with a non-zero status code
3. WHEN a request is processed, THE API_Server SHALL emit a structured JSON log entry containing timestamp (ISO 8601), level (DEBUG, INFO, WARN, ERROR), message, request method, request path, response status code, and request duration in milliseconds
4. THE API_Server SHALL expose a health check endpoint that verifies database connectivity and returns HTTP 200 with a JSON body indicating status "healthy" when the database is reachable, or HTTP 503 with status "unhealthy" when it is not
5. IF the database connection is unavailable at startup, THEN THE API_Server SHALL retry the connection up to 3 times with a 5-second interval between attempts, then log the error and exit with a non-zero status code
6. THE API_Server SHALL apply context-based timeout handling to all HTTP requests with a default timeout of 30 seconds, configurable via an environment variable
7. THE API_Server SHALL follow the clean architecture project layout with /cmd, /internal, and /pkg directories

### Requirement 12: Database Migration

**User Story:** As a developer, I want the existing database schema migrated to support the new security requirements, so that the Go application can operate with proper data integrity.

#### Acceptance Criteria

1. THE Database SHALL include a migration that alters the password column in both tbl_account and tbl_admin from char(25) to varchar(72) to accommodate bcrypt hashes (60 characters)
2. THE Database SHALL include a migration that adds foreign key constraints referencing tbl_account(account_no) on the following tables: tbl_account_type, tbl_address, tbl_balance, tbl_customer, tbl_feedback, tbl_login_history, tbl_requests, and tbl_transaction, with ON DELETE RESTRICT and ON UPDATE CASCADE behavior
3. THE Database SHALL include a migration that adds non-unique indexes on the following columns: tbl_transaction.account_no, tbl_transaction.trans_date, tbl_login_history.account_no, and tbl_feedback.account_no
4. THE Database SHALL include a migration script that hashes all existing plaintext passwords in tbl_account and tbl_admin in-place using bcrypt with a cost factor of 12, skipping any password value that is already 60 characters in length
5. THE Database SHALL preserve all existing data during migration, verified by matching row counts in every table before and after migration execution
6. IF a migration step fails, THEN THE Database SHALL roll back all changes from that step and report an error message indicating which migration step failed
7. WHEN the migration is executed more than once, THE Database SHALL skip already-applied steps without producing errors or duplicating schema changes

### Requirement 13: API Response Format

**User Story:** As a frontend developer, I want consistent API response formats, so that I can reliably parse responses from the Go backend.

#### Acceptance Criteria

1. THE API_Server SHALL return all successful responses with Content-Type "application/json" in a JSON envelope with a "data" field containing the response payload and a "meta" field containing a "timestamp" value indicating when the response was generated
2. THE API_Server SHALL return all error responses with Content-Type "application/json" in a JSON envelope with an "error" object containing "code" (a string matching the HTTP status reason), "message" (a human-readable description of maximum 500 characters), and a "details" field present only when the error relates to multiple validation failures
3. THE API_Server SHALL return appropriate HTTP status codes: 200 for success, 201 for creation, 400 for validation errors, 401 for authentication failures, 403 for authorization failures, 404 for not found, and 500 for server errors
4. WHEN returning paginated data, THE API_Server SHALL include pagination metadata in the "meta" field with total count, current page (starting at 1), page size (default 20, maximum 100), and total pages
5. IF a pagination request specifies a page number greater than total pages, THEN THE API_Server SHALL return an empty "data" array with pagination metadata reflecting zero results for the requested page

### Requirement 14: Login History Tracking

**User Story:** As a customer, I want to view my login history, so that I can detect unauthorized access to my account.

#### Acceptance Criteria

1. WHEN a customer requests login history, THE API_Server SHALL return a paginated list of login sessions for that customer, each entry containing the login timestamp, logout timestamp, and IP address, ordered by login timestamp descending with a default page size of 20 items
2. WHEN a customer authenticates successfully, THE Auth_Service SHALL record the login timestamp and the client IP address for the new login session
3. WHEN a customer logs out or a token expires, THE Auth_Service SHALL record the logout timestamp for the corresponding login session
4. IF a login session has no recorded logout timestamp, THEN THE API_Server SHALL return that entry with a null logout timestamp indicating the session is still active or was not explicitly terminated
5. IF a customer requests login history and no login records exist, THEN THE API_Server SHALL return an empty list with pagination metadata showing zero total count
