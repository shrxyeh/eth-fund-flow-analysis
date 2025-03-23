package analyzer

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/shrxyeh/ethereum-fund-flow/internal/etherscan"
	"golang.org/x/sync/errgroup"
)

// BeneficiaryAnalyzer is responsible for analyzing transactions to identify beneficiaries
type BeneficiaryAnalyzer struct {
	etherscanClient *etherscan.Client
	debug           bool
}

// NewBeneficiaryAnalyzer creates a new beneficiary analyzer
func NewBeneficiaryAnalyzer(etherscanClient *etherscan.Client) *BeneficiaryAnalyzer {
	return &BeneficiaryAnalyzer{
		etherscanClient: etherscanClient,
		debug:           true, // Enable debug logging
	}
}

// Beneficiary represents a beneficiary address with transaction details
type Beneficiary struct {
	Address      string               `json:"beneficiary_address"`
	Amount       float64              `json:"amount"`
	Transactions []TransactionDetails `json:"transactions"`
}

// TransactionDetails represents simplified transaction details
type TransactionDetails struct {
	TxAmount      float64 `json:"tx_amount"`
	DateTime      string  `json:"date_time"`
	TransactionID string  `json:"transaction_id"`
}

// AnalyzeBeneficiary analyzes the transaction flow for a given address to identify beneficiaries
func (ba *BeneficiaryAnalyzer) AnalyzeBeneficiary(address string) ([]Beneficiary, error) {
	if ba.debug {
		fmt.Printf("DEBUG: Starting beneficiary analysis for address: %s\n", address)
	}

	// Fetch all transaction types concurrently
	var normalTxs []etherscan.Transaction
	var internalTxs []etherscan.Transaction
	var tokenTransfers []etherscan.TokenTransfer
	var err error

	eg := errgroup.Group{}

	eg.Go(func() error {
		normalTxs, err = ba.etherscanClient.GetNormalTransactions(address)
		if err != nil {
			return fmt.Errorf("error fetching normal transactions: %w", err)
		}
		if ba.debug {
			fmt.Printf("DEBUG: Fetched %d normal transactions\n", len(normalTxs))
		}
		return nil
	})

	eg.Go(func() error {
		internalTxs, err = ba.etherscanClient.GetInternalTransactions(address)
		if err != nil {
			return fmt.Errorf("error fetching internal transactions: %w", err)
		}
		if ba.debug {
			fmt.Printf("DEBUG: Fetched %d internal transactions\n", len(internalTxs))
		}
		return nil
	})

	eg.Go(func() error {
		tokenTransfers, err = ba.etherscanClient.GetTokenTransfers(address)
		if err != nil {
			return fmt.Errorf("error fetching token transfers: %w", err)
		}
		if ba.debug {
			fmt.Printf("DEBUG: Fetched %d token transfers\n", len(tokenTransfers))
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Process transactions to identify beneficiaries
	beneficiaryMap := make(map[string]*Beneficiary)

	// Process normal transactions
	for i, tx := range normalTxs {
		if ba.debug && i < 5 {
			fmt.Printf("DEBUG: Normal tx %d - From: %s, To: %s, Value: %s, Hash: %s, IsError: %s\n", 
				i, tx.From, tx.To, tx.Value, tx.Hash, tx.IsError)
			
			// Convert timestamp
			timestamp, err := stringToInt64(tx.TimeStamp)
			if err == nil {
				formattedTime := time.Unix(timestamp, 0).Format("2006-01-02 15:04:05")
				fmt.Printf("DEBUG: Timestamp: %s -> %s\n", tx.TimeStamp, formattedTime)
			}
		}

		// Only consider outgoing transactions (where this address is the source)
		if strings.EqualFold(tx.From, address) && tx.IsError == "0" {
			if ba.debug {
				fmt.Printf("DEBUG: Processing outgoing normal transaction to %s with value %s\n", tx.To, tx.Value)
			}
			ba.processBeneficiary(beneficiaryMap, tx.To, tx.Value, tx.Hash, tx.TimeStamp)
		}
	}

	// Process internal transactions
	for i, tx := range internalTxs {
		if ba.debug && i < 5 {
			fmt.Printf("DEBUG: Internal tx %d - From: %s, To: %s, Value: %s, Hash: %s, IsError: %s\n", 
				i, tx.From, tx.To, tx.Value, tx.Hash, tx.IsError)
		}

		// Only consider outgoing transactions
		if strings.EqualFold(tx.From, address) && tx.IsError == "0" {
			if ba.debug {
				fmt.Printf("DEBUG: Processing outgoing internal transaction to %s with value %s\n", tx.To, tx.Value)
			}
			ba.processBeneficiary(beneficiaryMap, tx.To, tx.Value, tx.Hash, tx.TimeStamp)
		}
	}

	// Process token transfers
	for i, transfer := range tokenTransfers {
		if ba.debug && i < 5 {
			fmt.Printf("DEBUG: Token transfer %d - From: %s, To: %s, Value: %s, Token: %s (%s), Hash: %s\n", 
				i, transfer.From, transfer.To, transfer.Value, transfer.TokenName, transfer.TokenSymbol, transfer.Hash)
		}

		// Only consider outgoing transfers
		if strings.EqualFold(transfer.From, address) {
			if ba.debug {
				fmt.Printf("DEBUG: Processing outgoing token transfer to %s with value %s of token %s\n", 
					transfer.To, transfer.Value, transfer.TokenSymbol)
			}
			ba.processBeneficiary(beneficiaryMap, transfer.To, transfer.Value, transfer.Hash, transfer.TimeStamp)
		}
	}

	// Convert map to slice
	beneficiaries := make([]Beneficiary, 0, len(beneficiaryMap))
	for _, beneficiary := range beneficiaryMap {
		beneficiaries = append(beneficiaries, *beneficiary)
	}

	if ba.debug {
		fmt.Printf("DEBUG: Found %d beneficiary addresses\n", len(beneficiaries))
		for i, b := range beneficiaries {
			if i < 5 { // Limit to first 5 for brevity
				fmt.Printf("DEBUG: Beneficiary %d - Address: %s, Total: %f ETH, Transactions: %d\n", 
					i, b.Address, b.Amount, len(b.Transactions))
			}
		}
	}

	return beneficiaries, nil
}

// processBeneficiary adds a transaction to the beneficiary map
func (ba *BeneficiaryAnalyzer) processBeneficiary(beneficiaryMap map[string]*Beneficiary, 
	beneficiaryAddr, valueStr, hash, timestampStr string) {
		
	// Convert value to float (from Wei to Ether)
	value := new(big.Float)
	value.SetString(valueStr)
	divisor := new(big.Float)
	divisor.SetString("1000000000000000000") // 10^18 (Wei to Ether)
	value.Quo(value, divisor)

	amount, _ := value.Float64()

	// Format timestamp
	dateTime, err := etherscan.FormatTime(timestampStr)
	if err != nil {
		dateTime = timestampStr // Use original timestamp if formatting fails
		if ba.debug {
			fmt.Printf("DEBUG: Failed to format timestamp %s: %v\n", timestampStr, err)
		}
	}

	if ba.debug {
		timestamp, err := stringToInt64(timestampStr)
		if err == nil {
			fmt.Printf("DEBUG: Raw timestamp: %s -> Unix: %d -> Formatted: %s\n", 
				timestampStr, timestamp, time.Unix(timestamp, 0).Format("2006-01-02 15:04:05"))
		}
	}

	// Create transaction details
	txDetails := TransactionDetails{
		TxAmount:      amount,
		DateTime:      dateTime,
		TransactionID: hash,
	}

	// Add to beneficiary map
	if b, exists := beneficiaryMap[beneficiaryAddr]; exists {
		b.Amount += amount
		b.Transactions = append(b.Transactions, txDetails)
	} else {
		beneficiaryMap[beneficiaryAddr] = &Beneficiary{
			Address:      beneficiaryAddr,
			Amount:       amount,
			Transactions: []TransactionDetails{txDetails},
		}
	}
}

// Helper function to convert string to int64
// Helper function to convert string to int64
func stringToInt64(s string) (int64, error) {
    // Parse the string to an integer
    i, ok := new(big.Int).SetString(s, 10)
    if !ok {
        return 0, fmt.Errorf("failed to parse string to int: %s", s)
    }
    
    // Check if the value fits in an int64
    if !i.IsInt64() {
        return 0, fmt.Errorf("value too large for int64: %s", s)
    }
    
    return i.Int64(), nil
}