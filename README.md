# Timber

Timber is a lightweight, self-hosted web-based log viewer and file browser. It allows you to securely expose and view files from your server through a clean and intuitive web interface.

## Features

* **Web-based UI:** A simple and fast web interface to browse and view your files.
* **Authentication:** Protect your files with basic authentication.
* **Access Control:** Fine-grained access control using glob patterns to specify which users can access which files.
* **File Viewing:**
  * `cat`: View the entire content of a file.
  * `head`: View the first N lines of a file.
  * `tail`: View the last N lines of a file.
  * `follow`: Real-time log tailing (`tail -f`).
* **Download:** Download files directly from the web interface.
* **JSON Viewer:** Automatically parse line-delimited JSON files and display them in a structured table.

## Getting Started

### Prerequisites

* Go (for building from source)
* Docker (for running in a container)

### Building from Source

1. Clone the repository:

   ```bash
   git clone https://github.com/fmotalleb/timber.git
   cd timber
   ```

2. Build the binary:

   ```bash
   go build -o timber main.go
   ```

3. Run the server:

   ```bash
   ./timber -c config.toml
   ```

### Running with Docker

You can use the provided `Dockerfile` to build and run Timber in a Docker container.

1. Build the Docker image:

   ```bash
   docker build -t timber .
   ```

2. Run the container with your configuration file:

   ```bash
   docker run -p 8080:8080 -v $(pwd)/config.toml:/app/config.toml timber
   ```

## Configuration

Timber is configured using a TOML file. Here is an example `config.toml`:

```toml
listen = "0.0.0.0:8080"

# Define users and their access rights
# Format: "user:password@access_group1,access_group2"
users = [
  "admin:supersecret@all_logs",
  "guest:guest@app_logs"
]

# or 
[[users]]
name = "admin"
pass = supersecret
access = [
  "all_logs",
]

[access]
# Define access groups with file paths (glob patterns are supported)
[access.all_logs]
path = ["/var/log/*.log", "/var/log/app/**/*.log"]

[access.app_logs]
path = "/var/log/app/application.log"
```

### Configuration Details

* **`listen`**: The address and port the server will listen on. (Default: `127.0.0.1:8080`, Env: `LISTEN`)
* **`users`**: A list of users.
  * You can define users as a list of strings in the format `"username:password@access_group1,access_group2,..."`.
  * Alternatively, you can use a more structured format:
        ```toml
        [[users]]
        name = "admin"
        password = "supersecret"
        access = ["all_logs"]
        ```
* **`access`**: A map of access groups to file paths.
  * The key is the name of the access group.
  * `path` can be a single string or a list of strings containing file paths. Glob patterns are supported.

## Usage

1. Start the Timber server with your configuration file.
2. Open your web browser and navigate to the address specified in your configuration (e.g., `http://localhost:8080`).
3. Log in with the credentials you defined in the configuration.
4. The main application screen will show a list of files you have access to.
5. For each file, you can:
   * **cat**: View the entire file.
   * **head**: View the beginning of the file.
   * **tail**: View the end of the file.
   * **follow**: Stream the file in real-time.
   * **download**: Download the file.
   * **View as JSON**: If the file contains line-delimited JSON, this will display it in a table.

## API Endpoints

The following API endpoints are available. All endpoints require Basic Authentication.

* `GET /me`: Returns information about the currently authenticated user.
* `GET /filesystem/ls`: Lists the files the user has access to.
* `GET /filesystem/cat?path=<path>`: Returns the content of the specified file.
* `GET /filesystem/head?path=<path>&lines=<n>`: Returns the first `n` lines of the file.
* `GET /filesystem/tail?path=<path>&lines=<n>&follow=<true|false>`: Returns the last `n` lines of the file. If `follow=true`, it will stream the file.
