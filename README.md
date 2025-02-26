# Gymble IMU Backend

## Overview
Gymble IMU Backend is a Go-based application that provides an API endpoint to handle IMU (Inertial Measurement Unit) data. It stores the received data into InfluxDB.

## Prerequisites
- Docker and Docker Compose installed on your machine.
- Environment variables for InfluxDB configuration (`INFLUXDB_TOKEN`, `INFLUXDB_ORG`, `INFLUXDB_BUCKET`, `INFLUXDB_USER`, `INFLUXDB_PASSWORD`).

## Services

### App Service
The main application service that handles incoming IMU data and stores it into InfluxDB.
- **Build**: Uses the Dockerfile in the root directory.
- **Ports**: Maps port 8080 of the container to port 8080 on the host.
- **Environment Variables**:
  - `INFLUXDB_URL`: The URL of the InfluxDB instance (default: `http://influxdb:8086`).
  - `INFLUXDB_TOKEN`: The token for authenticating with InfluxDB.
  - `INFLUXDB_ORG`: The organization name in InfluxDB.
  - `INFLUXDB_BUCKET`: The bucket name where the data will be stored.

### InfluxDB Service
An instance of InfluxDB used to store the IMU data.
- **Image**: Uses the official InfluxDB image with tag `2.7`.
- **Ports**: Maps port 8086 of the container to port 8086 on the host.
- **Volumes**: Uses a Docker volume named `influxdb-data` to persist InfluxDB data.
- **Environment Variables**:
  - `DOCKER_INFLUXDB_INIT_MODE`: Set to `setup` for initial setup.
  - `DOCKER_INFLUXDB_INIT_USERNAME`, `DOCKER_INFLUXDB_INIT_PASSWORD`, `DOCKER_INFLUXDB_INIT_ORG`, `DOCKER_INFLUXDB_INIT_BUCKET`, and `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN`: Used for initializing InfluxDB.

## Docker Compose

To start the application and InfluxDB services, run:

```bash
docker-compose up --build
```

This command will build the images if they do not exist and start both services on their respective ports.

## API Endpoints

### POST /imu
Handles IMU data sent to the `/imu` endpoint. The received data is then stored in InfluxDB using the provided configuration.

## Environment Variables

- `INFLUXDB_TOKEN`: Required. Token for authenticating with InfluxDB.
- `INFLUXDB_ORG`: Required. Organization name in InfluxDB.
- `INFLUXDB_BUCKET`: Required. Bucket name where the data will be stored.
- `INFLUXDB_USER`: Required for initializing InfluxDB.
- `INFLUXDB_PASSWORD`: Required for initializing InfluxDB.

## Development

To run the application locally, ensure you have Go installed and set up your environment variables. Then, build and run:

```bash
go build -o gymble .
./gymble
```

Ensure that an instance of InfluxDB is running accessible via `http://localhost:8086` with the appropriate credentials.

## Dependencies

The project dependencies are managed using Go modules. They can be found in the [go.mod](go.mod) file.