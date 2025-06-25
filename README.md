# Gin Batch Scheduler Service

A robust batch processing service built with **Go** and **Gin Gonic**, capable of reading and processing transaction records in parallel with retry logic, structured logging, and job tracking. This system is highly extensible and ready for production-scale workloads.  

This project implements a scheduled batch job system that reads data from a PostgreSQL-backed `transactions` table, processes it in **parallel using goroutines**, and maintains **execution state and retry history** using **structured logging and batch job tracking tables**.

The scheduler is modular, separated by job frequency (daily, weekly, monthly, duration), and is configured from the entrypoint in `cmd/main.go`.

---


## ✨ Features

The system includes the following key features to ensure performance and reliability::

### 🕒 Flexible Job Scheduling

Define and control when jobs should run (daily, weekly, monthly, or at fixed intervals) using `gocron/v2`. Job schedules are centralized in `pkg/scheduler`, ensuring modular and reusable task definitions.


### 🔁 Intelligent Retry with Exponential Backoff

Failed operations are retried using `cenkalti/backoff`, with exponential backoff, max retries, and retry notification support — minimizing transient failure impact.


### ⚙️ High-Performance Parallel Processing  

Transactions are processed concurrently using multiple goroutines, with a configurable worker pool. Tasks are distributed via buffered channels for optimal throughput.


### 🧾 Reliable Job Execution Tracking

Each job run is logged into `batch_job_executions` with metadata: start/end times, success/failure stats, exit codes, and messages. All failure records are captured in `batch_job_failure_details` for observability and audit.


---

## 🧭 How It Works

The following diagram illustrates the end-to-end flow of how a scheduled batch job is executed by the system. It covers job initiation, transaction batching, parallel processing with retry logic, and result tracking in the database.

```pgsql
┌──────────────────────────────────────────────┐
│      [1] Scheduler Triggers Batch Job        │
│----------------------------------------------│
│ - Triggered via gocron (daily/weekly/etc.)   │
│ - Initialized from cmd/main.go               │
│ - JobType: e.g., "transaction"               │
└──────────────────────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────┐
│ [2] Insert New BatchJobExecution Record      │
│----------------------------------------------│
│ - Status: IN_PROGRESS                        │
│ - StartTime: current timestamp               │
│ - JobType: transaction                       │
└──────────────────────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────┐
│     [3] Read Transactions by Batch Size      │
│----------------------------------------------│
│ - Query: SELECT ... FROM transactions LIMIT 5│
│ - Offset increments per batch iteration      │
│ - Continue until no more data is returned    │
└──────────────────────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────┐
│    [4] Concurrent Batch Processing           │
│----------------------------------------------│
│ - Spawn N workers (goroutines)               │
│ - Each worker reads from a shared channel    │
│ - Each record is processed via task handler  │
└──────────────────────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────┐
│    [5] Retry Failed Transactions             │
│----------------------------------------------│
│ - Use backoff with retry limit               │
│ - Retry delays increase exponentially        │
│ - After max retries, log as permanent fail   │
└──────────────────────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────┐
│    [6] Log Failed Items to Failure Table     │
│----------------------------------------------│
│ - Insert into batch_job_failure_details      │
│ - Includes: BatchID, DataID, error message   │
└──────────────────────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────┐
│    [7] Finalize BatchJobExecution            │
│----------------------------------------------│
│ - Set EndTime                                │
│ - Status: COMPLETED or FAILED                │
│ - Update: NumOfCompleted, NumOfFailed        │
│ - Store exit code and message                │
└──────────────────────────────────────────────┘

```
---


## 🤖 Tech Stack

This project leverages a clean and robust Go-based architecture designed for scheduled batch processing, retryable task execution, and structured logging. Below is an overview of the key tools and libraries used:

| Component                 | Description                                                                                   |
|---------------------------|-----------------------------------------------------------------------------------------------|
| **Language**              | `Go` (Golang) - Statically typed language with built-in support for concurrency and performance |
| **Web Framework**         | `Gin Gonic` - Lightweight HTTP router and middleware for building web APIs                    |
| **ORM**                   | `GORM` - ORM library for Go, used for PostgreSQL integration and query abstraction            |
| **Database Driver**       | `gorm.io/driver/postgres` - PostgreSQL driver for GORM                                        |
| **Scheduler**             | `go-co-op/gocron/v2` - Cron and interval-based job scheduler                                  |
| **Retry Mechanism**       | `cenkalti/backoff` - Retry utility with exponential backoff and max retry control             |
| **Logging**               | `logrus` - Structured and leveled logging for all job and task activity                       |
| **Log Rotation**          | `lumberjack.v2` - Handles automatic log file rotation and size management                     |
| **UUID Generation**       | `github.com/google/uuid` - Generates unique IDs for transactions and batch executions         |
| **Concurrency Handling**  | `Goroutines` and buffered channels - For parallel batch processing across multiple workers    |

---

## 🧱 Architecture Overview

TThis project follows a **modular**, **scalable**, and **testable** architecture based on **Clean Architecture** principles. It separates concerns between business logic, data access, delivery mechanisms, and infrastructure components — enabling easier maintenance and extensibility over time.

Each layer is clearly isolated into packages such as `entity`, `repository`, `service`, and `scheduler`, and the project structure encourages dependency inversion and single responsibility.

```bash
📁 go-batch-jobs-with-retry/
├── 📂cmd/                                  # Entry point of the application (e.g., main.go, scheduler bootstrapping)
├── 📂config/
│   └── 📂database/                         # PostgreSQL configuration (DSN, connection pool, migrations)
├── 📂docker/                               # Docker-related configurations
│   ├── 📂app/                              # Dockerfile for building the Go application
│   └── 📂postgres/                         # PostgreSQL container setup (e.g., Dockerfile, init scripts)
├── 📂internal/                             # Business logic grouped by domain
│   ├── 📂entity/                           # Core domain entities (e.g., Transaction, BatchJobExecution)
│   ├── 📂repository/                       # Abstraction layer for database operations using GORM
│   └── 📂service/                          # Application services coordinating business use cases and flow
├── 📂logs/                                 # Directory for storing rotated log files (info, error, etc.)
└── 📂pkg/                                  # Reusable packages and cross-cutting concerns
    ├── 📂logger/                           # Logrus + Lumberjack setup for structured logging with rotation
    └── 📂scheduler/                        # Custom job schedulers (daily, weekly, duration, monthly) using gocron
```

---

## 🛠️ Installation & Setup  

Follow the instructions below to get the project up and running in your local development environment. You may run it natively or via Docker depending on your preference.  

### ✅ Prerequisites

Make sure the following tools are installed on your system:

| **Tool**                                                      | **Description**                           |
|---------------------------------------------------------------|-------------------------------------------|
| [Go](https://go.dev/dl/)                                      | Go programming language (v1.20+)          |
| [Make](https://www.gnu.org/software/make/)                    | Build automation tool (`make`)            |
| [PostgreSQL](https://www.postgresql.org/)                     | Relational database system (v14+)         |
| [Docker](https://www.docker.com/)                             | Containerization platform (optional)      |

### 🔁 Clone the Project  

Clone the repository:  

```bash
git clone https://github.com/yoanesber/Go-Batch-Jobs-with-Retry.git
cd Go-Batch-Jobs-with-Retry
```

### ⚙️ Configure `.env` File  

Set up your **database**, **Redis**, and **JWT configuration** in `.env` file. Create a `.env` file at the project root directory:  

```properties
# Application configuration
ENV=PRODUCTION
API_VERSION=1.0
PORT=1000
IS_SSL=FALSE

# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASS=P@ssw0rd
DB_NAME=transactions
DB_SCHEMA=public
DB_SSL_MODE=disable
# Options: disable, require, verify-ca, verify-full
DB_TIMEZONE=Asia/Jakarta
DB_MIGRATE=TRUE
DB_SEED=TRUE
DB_SEED_FILE=import.sql
# Set to INFO for development and staging, SILENT for production
DB_LOG=SILENT
```

- **🔐 Notes**:  
  - `DB_TIMEZONE=Asia/Jakarta`: Adjust this value to your local timezone (e.g., `Asia/Jakarta`, etc.).
  - `DB_MIGRATE=TRUE`: Set to `TRUE` to automatically run `GORM` migrations for all entity definitions on app startup.
  - `DB_SEED=TRUE` & `DB_SEED_FILE=import.sql`: Use these settings if you want to insert predefined data into the database using the SQL file provided.
  - `DB_USER=appuser`, `DB_PASS=app@123`: It's strongly recommended to create a dedicated database user instead of using the default postgres superuser.

### 👤 Create Dedicated PostgreSQL User (Recommended)

For security reasons, it's recommended to avoid using the default postgres superuser. Use the following SQL script to create a dedicated user (`appuser`) and assign permissions:

```sql
-- Create appuser and database
CREATE USER appuser WITH PASSWORD 'app@123';

-- Allow user to connect to database
GRANT CONNECT, TEMP, CREATE ON DATABASE transactions TO appuser;

-- Grant permissions on public schema
GRANT USAGE, CREATE ON SCHEMA public TO appuser;

-- Grant all permissions on existing tables
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO appuser;

-- Grant all permissions on sequences (if using SERIAL/BIGSERIAL ids)
GRANT USAGE, SELECT, UPDATE ON ALL SEQUENCES IN SCHEMA public TO appuser;

-- Ensure future tables/sequences will be accessible too
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO appuser;

-- Ensure future sequences will be accessible too
ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO appuser;
```

Update your `.env` accordingly:
```properties
DB_USER=appuser
DB_PASS=app@123
```

---


## 🚀 Running the Application  

This section provides step-by-step instructions to run the application either **locally** or via **Docker containers**.

- **Notes**:  
  - All commands are defined in the `Makefile`.
  - To run using `make`, ensure that `make` is installed on your system.
  - To run the application in containers, make sure `Docker` is installed and running.
  - Ensure you have `Go` installed on your system

### 📦 Install Dependencies

Make sure all Go modules are properly installed:  

```bash
make tidy
```

### 🧪 Run Unit Tests

```bash
make test
```

### 🔧 Run Locally (Non-containerized)

Ensure PostgreSQL are running locally, then:

```bash
make run
```

### 🐳 Run Using Docker

To build and run all services (PostgreSQL, Go app):

```bash
make docker-up
```

To stop and remove all containers:

```bash
make docker-down
```

- **Notes**:  
  - Before running the application inside Docker, make sure to update your environment variables `.env`
    - Change `DB_HOST=localhost` to `DB_HOST=batch-job-postgres`.

### 🟢 Application is Running

Now your application is accessible at:
```bash
http://localhost:1000
```

---

## 🧪 Testing Scenarios  

To ensure reliability and correctness of the batch processing system, the following key testing scenarios are recommended:

### Scheduler Trigger Test

- **Goal**: Validate that scheduled jobs trigger at the expected times.
- **Test**:
  - Use short intervals (e.g., every 10 seconds) during testing.
  - Ensure that `BatchJobExecution` record is created on schedule.
  - Check job log output for timestamp alignment.


### Batch Read Pagination

- **Goal**: Ensure transactions are read in batches (e.g., 5 records per read) until exhausted.
- **Test**:
  - Seed the `transactions` table with more than 5 dummy rows.
  - Assert that each batch contains at most 5 items.
  - Confirm that loop terminates once all records are processed.

### Concurrent Processing Test

- **Goal**: Validate that multiple worker goroutines process records concurrently.
- **Test**:
  - Use a batch with multiple items.
  - Add debug logs to identify worker ID per record.
  - Assert that multiple workers process records in parallel.

### Retry Mechanism Test

- **Goal**: Verify that transient errors are retried with exponential backoff.
- **Test**:
  - Force an error in `handleTransaction` for a specific record.
  - Ensure retry attempts are logged (based on `notify` function).
  - Confirm retries stop after `maxRetries`.

### Failure Record Logging

- **Goal**: Ensure that permanently failed transactions are recorded.
- **Test**:
  - Simulate a processing failure beyond max retries.
  - Assert that an entry is created in `batch_job_failure_details` with appropriate message.

### Job Finalization Test

- **Goal**: Validate that BatchJobExecution is correctly updated when processing ends.
- **Test**:
  - Assert `EndTime`, `Status`, `ExitCode`, and `ExitMessage` are populated.
  - Check `NumOfCompleted` and `NumOfFailed` reflect actual results.
