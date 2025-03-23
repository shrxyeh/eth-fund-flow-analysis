</thinking>

# Ethereum Fund Flow Analysis API

A Go-based API for analyzing the flow of funds on the Ethereum blockchain. This tool helps identify beneficiaries (where funds are going to) and payers (where funds are coming from) for any Ethereum address.

## Features

- **Beneficiary Analysis**: Identify final recipients of funds from a given Ethereum address
- **Payer Analysis**: Identify sources of funds received by a given Ethereum address
- **Multiple Transaction Types**:
  - Normal Ethereum transactions
  - Internal transactions (contract interactions)
  - Token transfers (ERC-20, ERC-721, ERC-1155)
- **Command-Line Interface**: Specify addresses and analysis modes via command line
- **Web Interface**: Interactive web UI to visualize and explore results
- **Robust Error Handling**: Handles API timeouts, rate limits, and pagination
- **Detailed Transaction Information**: View amount, timestamp, and transaction hash

## Technologies Used

- **Go** (Golang 1.19+)
- **Etherscan API** for blockchain data retrieval
- **Gorilla Mux** for HTTP routing
- **Concurrent API Calls** for efficient data processing
- **JSON Response Format** for standardized data interchange

## Installation

### Prerequisites

- Go 1.19 or higher
- Etherscan API key (get one for free at [https://etherscan.io/apis](https://etherscan.io/apis))

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/shrxyeh/ethereum-fund-flow.git
   cd ethereum-fund-flow
   ```

2. Create a `.env` file with your Etherscan API key:
   ```
   ETHERSCAN_API_KEY=your_etherscan_api_key_here
   PORT=8080
   ```

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Build the application:
   ```bash
   go build -o bin/api ./cmd/api
   ```

## Usage

### Command Line Arguments

The application accepts the following command-line arguments:

- `-address`: Ethereum address to analyze (default: WETH contract if not specified)
- `-mode`: Analysis mode: "beneficiary", "payer", or "both" (default: "both")
- `-port`: Server port to listen on (overrides .env PORT)
- `-help`: Show usage information

Examples:

```bash
# Run with default settings
./bin/api

# Analyze a specific address for beneficiaries
./bin/api -address=0x7a250d5630b4cf539739df2c5dacb4c659f2488d -mode=beneficiary

# Analyze a specific address for payers
./bin/api -address=0xb8901acb165ed027e32754e0ffe830802919727f -mode=payer

# Run on a different port
./bin/api -port=9090

# Show help
./bin/api -help
```

### Web Interface

After starting the server, access the web interface at `http://localhost:8080/` (or your configured port).

The interface provides:
- Access to API endpoints with documentation
- Examples with current command-line settings
- Health check status
- Sample Ethereum addresses for testing

## API Endpoints

### Health Check

```
GET /health
```

Returns `OK` if the API is running.

### Beneficiary Analysis

```
GET /beneficiary?address={ethereum_address}
```

Identifies where funds are flowing to from the given address.

Example Response:
```json
{
  "message": "success",
  "data": [
    {
      "beneficiary_address": "0x6032de3d44b46cdbca9f8e078cf534c96b3e2f12",
      "amount": 0.000072888245889635,
      "transactions": [
        {
          "tx_amount": 0.000072888245889635,
          "date_time": "2023-03-23 12:01:23",
          "transaction_id": "0x3f1a19ffd94a6bdeee14187b5040bb5b5cc77ede2a8733e1169a180e0143db10"
        }
      ]
    }
  ]
}
```

### Payer Analysis

```
GET /payer?address={ethereum_address}
```

Identifies where funds are coming from to the given address.

Example Response:
```json
{
  "message": "success",
  "data": [
    {
      "payer_address": "0x742d35cc6634c0532925a3b844bc454e4438f44e",
      "amount": 0.8,
      "transactions": [
        {
          "tx_amount": 0.8,
          "date_time": "2023-03-15 14:22:10",
          "transaction_id": "0x1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0t1u2v3w4x5y6z7a8b9c0d1e2f"
        }
      ]
    }
  ]
}
```

## Architecture

The application follows a clean, layered architecture:

### Project Structure

```
ethereum-fund-flow/
├── cmd/
│   └── api/
│       └── main.go           # Application entry point
├── internal/
│   ├── api/
│   │   ├── handler.go        # HTTP request handlers
│   │   ├── router.go         # HTTP router setup
│   │   └── server.go         # HTTP server
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── etherscan/
│   │   ├── client.go         # Etherscan API client
│   │   └── models.go         # Etherscan data models
│   ├── analyzer/
│   │   ├── beneficiary.go    # Beneficiary analysis logic
│   │   └── payer.go          # Payer analysis logic
├── pkg/
│   └── logger/
│       └── logger.go         # Logging functionality
├── .env                      # Environment variables
├── .gitignore
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
└── README.md                 # Project documentation
```

### Components

1. **Etherscan Client**: Communicates with the Etherscan API to retrieve transaction data
2. **Analyzers**: Process transaction data to identify beneficiaries and payers
3. **API Handlers**: Handle HTTP requests and responses
4. **Router**: Defines API endpoints and routes
5. **Server**: Manages the HTTP server lifecycle
6. **Configuration**: Handles environment variables and command-line flags
7. **Logger**: Provides structured logging

## Debugging

The application includes debug logging to help understand the flow of data and troubleshoot issues. Debug mode can be enabled in the Etherscan client and analyzers.

Common debug information includes:
- API requests and responses
- Transaction processing
- Time formatting and conversion
- Beneficiary and payer identification

## Performance Considerations

- For addresses with many transactions (like popular contracts), the API uses pagination to limit results to the most recent 100 transactions
- The HTTP client timeout is set to 60 seconds to accommodate larger requests
- Concurrent API calls improve performance when fetching different transaction types
- Exponential backoff retry logic handles temporary API failures

## Troubleshooting

### Common Issues

1. **Timeout Errors**:
   - Occurs with very popular addresses (like WETH) that have millions of transactions
   - Solution: Use a more specific address or increase the client timeout

2. **API Rate Limits**:
   - Etherscan limits API calls for free accounts
   - Solution: Implement caching or upgrade to a paid Etherscan API plan

3. **No Transactions Found**:
   - Ensure the address is correct and has transaction history
   - Try both beneficiary and payer analysis modes

4. **Date Formatting Issues**:
   - Timestamps from Etherscan are Unix timestamps
   - The application converts these to human-readable dates

## Future Improvements

- Add transaction caching to reduce API calls
- Implement pagination for large result sets
- Add filtering by date range or transaction type
- Visualize fund flow as a network graph
- Support other blockchain explorers beyond Etherscan
- Add support for other EVM-compatible chains

## Sample Ethereum Addresses for Testing

- `0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2` (WETH Contract)
- `0x7a250d5630b4cf539739df2c5dacb4c659f2488d` (Uniswap V2 Router)
- `0xb8901acb165ed027e32754e0ffe830802919727f` (Sample address with various transactions)

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Etherscan API](https://etherscan.io/apis) for providing blockchain data
- [Gorilla Mux](https://github.com/gorilla/mux) for HTTP routing
- [godotenv](https://github.com/joho/godotenv) for environment variable management