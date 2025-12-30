The goal of this project is to use ai to format pictures of receipts into a format that can be further processed.

# Setup
``` bash
go get ./...
```

# Configuration

Copy the example environment file and configure your API keys:
```bash
cp .env.example .env
```

Edit `.env` and set:
- `API_KEY`: Your GenAI API key
- `BANK_API_KEY`: Your Mercury Bank API key  
- `BASIC_AUTH_USERNAME`: Username for HTTP basic authentication
- `BASIC_AUTH_PASSWORD`: Password for HTTP basic authentication
- `PORT`: Server port (defaults to 8080)

# Running the Server

```bash
# Load environment variables and run
source .env && go run src/main.go
```

# API Endpoints

## Health Check (No Auth Required)
```bash
curl http://localhost:8080/health
```

## Fetch Transactions (Basic Auth Required)
```bash
# Default: last 7 days
curl -u username:password http://localhost:8080/transactions

# Custom date range (specify days)
curl -u username:password http://localhost:8080/transactions?days=30
```

## Run AI Demo (Basic Auth Required)  
```bash
curl -u username:password http://localhost:8080/demo
```

Replace `username:password` with your actual credentials from the `.env` file.