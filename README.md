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

- **POST /login**: Authenticate user and obtain JWT and stores it in a cookie for secure access to protected routes.
- **POST /companies**: Create a new company entry.
- **GET /companies/{id}**: Retrieve company details by ID.
- **PATCH /companies/{id}**: Update existing company information.
- **DELETE /companies/{id}**: Remove a company record.

## Notes

- The MySQL container exposes port 3306, allowing you to connect to the database if needed. Also the db credentials provided from .env file.
- Kafka and Zookeeper can be accessed through their respective exposed ports (9092 for Kafka).
- Ensure that Docker and Docker Compose are properly configured on your environment in order to run the project seamlessly.

This project provides a robust framework for managing company data while ensuring security through token-based authentication and reliable event handling via Kafka.
