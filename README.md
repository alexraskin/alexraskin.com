# alexraskin.com

A simple Go web server that returns a personal website. It has two response types:
- A formatted plain text card for terminal clients (curl, HTTPie)
- HTML content for web browsers

## Running Locally

```bash
# Navigate to the project directory
cd alexraskin.com

# Run the server
go run main.go
```

The server will start on port 8080 by default. You can change this by setting the PORT environment variable.

## Docker

You can also run the application using Docker:

```bash
# Build the Docker image
docker build -t alexraskin-website .

# Run the container
docker run -p 8080:8080 alexraskin-website
```

## Testing the Terminal Output

To see the terminal output, use curl:

```bash
curl http://localhost:8080
```

## Accessing via Browser

Simply navigate to http://localhost:8080 in your web browser.

