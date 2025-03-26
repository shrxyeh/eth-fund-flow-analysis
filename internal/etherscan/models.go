package etherscan

import (
	"fmt"
	"strconv"
	"time"
)

// TransactionResponse represents the response from Etherscan API for transactions
type TransactionResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Result  []Transaction `json:"result"`
}

// TokenTransferResponse represents the response from Etherscan API for token transfers
type TokenTransferResponse struct {
	Status  string         `json:"status"`
	Message string         `json:"message"`
	Result  []TokenTransfer `json:"result"`
}

// Transaction represents a normal or internal Ethereum transaction
type Transaction struct {
	Hash              string `json:"hash"`
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	From              string `json:"from"`
	To                string `json:"to"`
	Value             string `json:"value"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	IsError           string `json:"isError"`
	TxReceiptStatus   string `json:"txreceipt_status"`
	Input             string `json:"input"`
	ContractAddress   string `json:"contractAddress"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	GasUsed           string `json:"gasUsed"`
	Confirmations     string `json:"confirmations"`
}

// TokenTransfer represents an ERC-20/ERC-721/ERC-1155 token transfer
type TokenTransfer struct {
	Hash              string `json:"hash"`
	BlockNumber       string `json:"blockNumber"`
	TimeStamp         string `json:"timeStamp"`
	From              string `json:"from"`
	To                string `json:"to"`
	Value             string `json:"value"`
	TokenName         string `json:"tokenName"`
	TokenSymbol       string `json:"tokenSymbol"`
	TokenDecimal      string `json:"tokenDecimal"`
	ContractAddress   string `json:"contractAddress"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	GasUsed           string `json:"gasUsed"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	Confirmations     string `json:"confirmations"`
}

// FormatTime formats the timestamp from the Etherscan API
func FormatTime(timestamp string) (string, error) {
	fmt.Printf("DEBUG: FormatTime called with timestamp: %s\n", timestamp)
	
	unixTime, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return "", fmt.Errorf("error parsing timestamp: %w", err)
	}
	
	t := time.Unix(unixTime, 0)
	
	formatted := t.Format("2006-01-02 15:04:05")
	
	fmt.Printf("DEBUG: Formatted timestamp: %s -> %s\n", timestamp, formatted)
	
	return formatted, nil
}