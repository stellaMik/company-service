# company-service

## Overview

This project is a RESTful API designed for managing company records, allowing users to create, retrieve, update, and delete company information securely. The API integrates with mySQL a database to handle data persistence and utilizes Kafka for event-driven functionalities, ensuring that actions such as company creation are published to a message broker for further processing.

## Features

- **User Authentication**: The API includes a login endpoint that allows users to authenticate using their username and password. Upon successful login, a JWT (JSON Web Token) is generated for secure access to protected routes.
- **CRUD Operations**:
  - **Create**: Users can create a new company record by providing necessary details such as name, description, number of employees, registration status, and type.
  - **Read**: Fetch existing company records using unique identifiers (UUID).
  - **Update**: Modify existing company details based on the UUID.
  - **Delete**: Remove company records from the database.
- **Input Validation**: The API ensures that inputs for creating or updating company records meet specified criteria, such as name length and employee count.
- **Event Publishing**: Creation, updates, and deletions of company records trigger events published to a Kafka topic, promoting asynchronous processing and decoupled architecture.

## Configuration

The application is configured via environment variables. Important configurations include database credentials, API port, JWT secrets, and Kafka settings. A `.env` file is used to manage these configurations, which include:

- Database Options (User, Password, Name, Host, Port)
- API and User Authentication settings
- Kafka server settings for message publishing

## Technology Stack

- **Programming Language**: Go (Golang)
- **Framework**: Gorilla Mux for routing
- **Database**: MySQL via GORM for ORM
- **Message Broker**: Kafka for event handling

## Usage

### Running with Docker

1. Ensure you have Docker installed on your machine.
2. Create a `.env` file in the project root and populate it with the necessary environment variables. Here’s an example of what it might look like(i have included one for testing reasons):

```env
API_PORT=8080
API_USER=user
API_PASSWORD=pass
DB_USER=myuser
DB_PASSWORD=mypassword
DB_NAME=mydatabase
DB_PORT=3306
JWT_SECRET=mysecret
KAFKA_URL=broker:9092
KAFKA_TOPIC=mytopic
KAFKA_GROUP_ID=mygroup
```

3. Build and run the application and its dependencies (MySQL, Kafka, and Zookeeper) using Docker Compose:

   ```bash
   docker-compose up --build
   ```

## Accessing the API

Once the services are up, you can access the API through `http://localhost:8080/api` or the port that user add to .env file.

## Endpoints

- **POST /login**: Authenticate user and obtain JWT and stores it in a cookie (15 minutes expiration) for secure access to protected routes.
- **POST /companies**: Create a new company entry. Only if user is authenticated.
- **GET /companies/{id}**: Retrieve company details by ID.
- **PATCH /companies/{id}**: Update existing company information. Only if user is authenticated.
- **DELETE /companies/{id}**: Remove a company record. Only if user is authenticated.

# Integration Test for Company Service

This section explains how to run the integration tests for the **Company Service**. The test simulates a series of interactions with the API and checks the integration with the database and Kafka message broker.

## Requirements

Before running the integration test, ensure you have the following prerequisites:

- **Go** installed.
- **Docker** (for running Kafka and database).
- **Database** MySQL running locally or in a Docker container.
- **Kafka** running locally or in a Docker container.
- Can run the docker-compose.yml
- **Ensure** the `.env` file with the appropriate configuration is set up in the project directory.

## Configuration

The integration test uses a configuration file (`config.LoadConfig`) to load environment variables from the `.env` file. Make sure your `.env` file contains the following:

```env
KAFKA_URL=localhost:9092
KAFKA_TOPIC=company-events
KAFKA_GROUP_ID=company-service-consumer-group
USER=testuser
PASSWORD=testpassword
```

You can modify the values based on your local setup.


## Running Test

1. **Start Kafka and the Database**: Run Kafka and database using Docker. You can use the following included `docker-compose.yml` file.
2. **Run the Test**: To execute the test, run the following command in your terminal from the root of the project, be sure that the database is clean and kafka consumer doesn't have any other traffic before the test:
   ```bash
   go test -v integration_test.go
   ```

## Integration Test Flow

The test performs the following steps:

1. **Login**:

   - Simulate a login request with hardcoded credentials and receive a JWT cookie for authentication.
2. **Create Company**:

   - A company is created using the `POST /api/companies` endpoint. The test verifies that the company is successfully created, and the company is published to Kafka with an event type `company_created`.
   - **Kafka Consumer**: The consumer listens for the event and ensures the `company_created` event is successfully consumed from the Kafka topic.
3. **Retrieve Company**:

   - The company is retrieved using the `GET /api/companies/{id}` endpoint, and the test verifies that the response matches the created company's data.
4. **Update Company**:

   - The company is updated using the `PATCH /api/companies/{id}` endpoint. The test checks that the company is updated, and a `company_updated` Kafka event is published.
   - **Kafka Consumer**: The consumer listens for the event and ensures the `company_updated` event is successfully consumed from the Kafka topic.
5. **Delete Company**:

   - The company is deleted using the `DELETE /api/companies/{id}` endpoint. The test checks that the company is deleted and verifies the event `company_deleted` is published to Kafka.
   - **Kafka Consumer**: The consumer listens for the event and ensures the `company_deleted` event is successfully consumed from the Kafka topic.
6. **Verify Deletion**:

   - The test ensures that after the company is deleted, a `GET` request to retrieve the company returns a `404 Not Found`.

## Notes

- The MySQL container exposes port 3306, allowing you to connect to the database if needed. Also the db credentials provided from .env file.
- Kafka and Zookeeper can be accessed through their respective exposed ports (9092 for Kafka).
- Ensure that Docker and Docker Compose are properly configured on your environment in order to run the project seamlessly.

This project provides a robust framework for managing company data while ensuring security through token-based authentication and reliable event handling via Kafka.
