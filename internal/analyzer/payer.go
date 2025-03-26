package analyzer

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/shrxyeh/ethereum-fund-flow/internal/etherscan"
	"golang.org/x/sync/errgroup"
)

// responsible for analyzing transactions to identify payers
type PayerAnalyzer struct {
	etherscanClient *etherscan.Client
}

// creates a new payer analyzer
func NewPayerAnalyzer(etherscanClient *etherscan.Client) *PayerAnalyzer {
	return &PayerAnalyzer{
		etherscanClient: etherscanClient,
	}
}

// Payer represents a payer address with transaction details
type Payer struct {
	Address      string               `json:"payer_address"`
	Amount       float64              `json:"amount"`
	Transactions []TransactionDetails `json:"transactions"`
}

// analyzes the transaction flow for a given address to identify payers
func (pa *PayerAnalyzer) AnalyzePayer(address string) ([]Payer, error) {
	// Fetch all transaction types concurrently
	var normalTxs []etherscan.Transaction
	var internalTxs []etherscan.Transaction
	var tokenTransfers []etherscan.TokenTransfer
	var err error

	eg := errgroup.Group{}

	eg.Go(func() error {
		normalTxs, err = pa.etherscanClient.GetNormalTransactions(address)
		if err != nil {
			return fmt.Errorf("error fetching normal transactions: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		internalTxs, err = pa.etherscanClient.GetInternalTransactions(address)
		if err != nil {
			return fmt.Errorf("error fetching internal transactions: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		tokenTransfers, err = pa.etherscanClient.GetTokenTransfers(address)
		if err != nil {
			return fmt.Errorf("error fetching token transfers: %w", err)
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Process transactions to identify payers
	payerMap := make(map[string]*Payer)

	// Process normal transactions
	for _, tx := range normalTxs {
		// Only consider incoming transactions (where this address is receiving)
		if strings.EqualFold(tx.To, address) && tx.IsError == "0" {
			pa.processPayer(payerMap, tx.From, tx.Value, tx.Hash, tx.TimeStamp)
		}
	}

	// Process internal transactions
	for _, tx := range internalTxs {
		// Only consider incoming transactions
		if strings.EqualFold(tx.To, address) && tx.IsError == "0" {
			pa.processPayer(payerMap, tx.From, tx.Value, tx.Hash, tx.TimeStamp)
		}
	}

	// Process token transfers
	for _, transfer := range tokenTransfers {
		// Only consider incoming transfers
		if strings.EqualFold(transfer.To, address) {
			pa.processPayer(payerMap, transfer.From, transfer.Value, transfer.Hash, transfer.TimeStamp)
		}
	}

	// Convert map to slice
	payers := make([]Payer, 0, len(payerMap))
	for _, payer := range payerMap {
		payers = append(payers, *payer)
	}

	return payers, nil
}

// adds a transaction to the payer map
func (pa *PayerAnalyzer) processPayer(payerMap map[string]*Payer, 
	payerAddr, valueStr, hash, timestampStr string) {
		
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
	}

	// Create transaction details
	txDetails := TransactionDetails{
		TxAmount:      amount,
		DateTime:      dateTime,
		TransactionID: hash,
	}

	// Add to payer map
	if p, exists := payerMap[payerAddr]; exists {
		p.Amount += amount
		p.Transactions = append(p.Transactions, txDetails)
	} else {
		payerMap[payerAddr] = &Payer{
			Address:      payerAddr,
			Amount:       amount,
			Transactions: []TransactionDetails{txDetails},
		}
	}
}