package etherscan

import (
	"encoding/json"
	"fmt"
	"io"
	// "log"
	"net/http"
	"strconv"
	"time"
)

const (
	baseURL = "https://api.etherscan.io/api"
)

// Client is the Etherscan API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	debug      bool
}

// NewClient creates a new Etherscan client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // Increase timeout to 60 seconds
		},
		debug: true, // Enable debug logging
	}
}

// GetNormalTransactions fetches normal transactions for an address with pagination
func (c *Client) GetNormalTransactions(address string) ([]Transaction, error) {
	// Only fetch a limited number of transactions to avoid timeout
	// Most beneficiary/payer analysis would focus on recent transactions anyway
	endpoint := fmt.Sprintf("%s?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&page=1&offset=100&sort=desc&apikey=%s",
		baseURL, address, c.apiKey)
	
	if c.debug {
		fmt.Printf("DEBUG: Fetching normal transactions for address: %s\n", address)
		fmt.Printf("DEBUG: API endpoint: %s\n", endpoint)
	}
		
	return c.fetchTransactions(endpoint)
}

// GetInternalTransactions fetches internal transactions for an address with pagination
func (c *Client) GetInternalTransactions(address string) ([]Transaction, error) {
	endpoint := fmt.Sprintf("%s?module=account&action=txlistinternal&address=%s&startblock=0&endblock=99999999&page=1&offset=100&sort=desc&apikey=%s",
		baseURL, address, c.apiKey)
	
	if c.debug {
		fmt.Printf("DEBUG: Fetching internal transactions for address: %s\n", address)
		fmt.Printf("DEBUG: API endpoint: %s\n", endpoint)
	}
		
	return c.fetchTransactions(endpoint)
}

// GetTokenTransfers fetches token transfers (ERC-20, ERC-721, ERC-1155) for an address with pagination
func (c *Client) GetTokenTransfers(address string) ([]TokenTransfer, error) {
	endpoint := fmt.Sprintf("%s?module=account&action=tokentx&address=%s&startblock=0&endblock=99999999&page=1&offset=100&sort=desc&apikey=%s",
		baseURL, address, c.apiKey)
	
	if c.debug {
		fmt.Printf("DEBUG: Fetching token transfers for address: %s\n", address)
		fmt.Printf("DEBUG: API endpoint: %s\n", endpoint)
	}
		
	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("error fetching token transfers: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if c.debug {
		// Print the first 500 characters of the response for debugging
		responsePreview := string(body)
		if len(responsePreview) > 1000 {
			responsePreview = responsePreview[:1000] + "... (truncated)"
		}
		fmt.Printf("DEBUG: Token transfers response: %s\n", responsePreview)
	}

	// First try to parse as a standard response with array result
	var result TokenTransferResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Handle error response where result might be a string instead of an array
	if result.Status == "0" {
		// Check if "No transactions found" - this is not really an error
		if result.Message == "No transactions found" {
			return []TokenTransfer{}, nil
		}
		
		// Try to parse as error response where Result is a string
		var errorResult struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Result  string `json:"result"`
		}
		
		if jsonErr := json.Unmarshal(body, &errorResult); jsonErr == nil && errorResult.Result != "" {
			// Handle rate limit errors
			if errorResult.Result == "Max rate limit reached" {
				return nil, fmt.Errorf("etherscan API rate limit exceeded, please try again later")
			}
			return nil, fmt.Errorf("etherscan API error: %s", errorResult.Result)
		}
		
		// Generic error
		return nil, fmt.Errorf("etherscan API error: %s", result.Message)
	}

	// Basic analysis of results for debugging
	if c.debug && len(result.Result) > 0 {
		fmt.Printf("DEBUG: Received %d token transfers\n", len(result.Result))
		if len(result.Result) > 0 {
			firstTx := result.Result[0]
			fmt.Printf("DEBUG: First token transfer - From: %s, To: %s, Token: %s, Value: %s\n", 
				firstTx.From, firstTx.To, firstTx.TokenSymbol, firstTx.Value)
		}
	}

	return result.Result, nil
}

// fetchTransactions is a helper function to fetch and parse transaction data
func (c *Client) fetchTransactions(endpoint string) ([]Transaction, error) {
	// Add retry with exponential backoff
	maxRetries := 3
	var resp *http.Response
	var err error
	
	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			// Wait before retrying (exponential backoff)
			wait := time.Duration(retry * retry) * time.Second
			time.Sleep(wait)
		}
		
		resp, err = c.httpClient.Get(endpoint)
		if err == nil {
			break
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("error fetching transactions after %d retries: %w", maxRetries, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if c.debug {
		// Print the first 500 characters of the response for debugging
		responsePreview := string(body)
		if len(responsePreview) > 1000 {
			responsePreview = responsePreview[:1000] + "... (truncated)"
		}
		fmt.Printf("DEBUG: Transaction response: %s\n", responsePreview)
	}

	// First try to parse as a standard response with array result
	var result TransactionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		// Check if it's a response with a non-array result (like error messages)
		var errorResult struct {
			Status  string `json:"status"`
			Message string `json:"message"`
			Result  string `json:"result"`
		}
		
		if jsonErr := json.Unmarshal(body, &errorResult); jsonErr == nil && errorResult.Result != "" {
			// Handle rate limit errors
			if errorResult.Result == "Max rate limit reached" {
				return nil, fmt.Errorf("etherscan API rate limit exceeded, please try again later")
			}
			return nil, fmt.Errorf("etherscan API error: %s", errorResult.Result)
		}
		
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Handle no results case more gracefully
	if result.Status == "0" && result.Message == "No transactions found" {
		if c.debug {
			fmt.Println("DEBUG: No transactions found in response")
		}
		return []Transaction{}, nil
	}

	// Check if we got an error instead of transactions
	if result.Status == "0" {
		return nil, fmt.Errorf("etherscan API error: %s", result.Message)
	}

	// Basic analysis of results for debugging
	if c.debug && len(result.Result) > 0 {
		fmt.Printf("DEBUG: Received %d transactions\n", len(result.Result))
		if len(result.Result) > 0 {
			firstTx := result.Result[0]
			fmt.Printf("DEBUG: First transaction - From: %s, To: %s, Value: %s, Hash: %s, Timestamp: %s\n", 
				firstTx.From, firstTx.To, firstTx.Value, firstTx.Hash, firstTx.TimeStamp)
			
			// Convert timestamp to human-readable date
			timestamp, err := strconv.ParseInt(firstTx.TimeStamp, 10, 64)
			if err == nil {
				formattedTime := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
				fmt.Printf("DEBUG: Timestamp converted: %s\n", formattedTime)
			}
		}
	}

	return result.Result, nil
}

// GetLatestBlockNumber fetches the latest block number
func (c *Client) GetLatestBlockNumber() (int, error) {
	endpoint := fmt.Sprintf("%s?module=proxy&action=eth_blockNumber&apikey=%s", baseURL, c.apiKey)
	
	if c.debug {
		fmt.Printf("DEBUG: Fetching latest block number\n")
		fmt.Printf("DEBUG: API endpoint: %s\n", endpoint)
	}
	
	resp, err := c.httpClient.Get(endpoint)
	if err != nil {
		return 0, fmt.Errorf("error fetching latest block: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %w", err)
	}

	if c.debug {
		fmt.Printf("DEBUG: Latest block response: %s\n", string(body))
	}

	var result struct {
		JsonRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Result  string `json:"result"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Convert hex string to int
	blockNumber, err := strconv.ParseInt(result.Result[2:], 16, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing block number: %w", err)
	}

	if c.debug {
		fmt.Printf("DEBUG: Latest block number: %d\n", blockNumber)
	}

	return int(blockNumber), nil
}