package api

import (
	"encoding/json"
	"net/http"

	"github.com/shrxyeh/ethereum-fund-flow/internal/analyzer"
	"github.com/shrxyeh/ethereum-fund-flow/pkg/logger"
)

// Handler handles API requests
type Handler struct {
	beneficiaryAnalyzer *analyzer.BeneficiaryAnalyzer
	payerAnalyzer       *analyzer.PayerAnalyzer
	logger              logger.Logger
}

// NewHandler creates a new API handler
func NewHandler(beneficiaryAnalyzer *analyzer.BeneficiaryAnalyzer, payerAnalyzer *analyzer.PayerAnalyzer, logger logger.Logger) *Handler {
	return &Handler{
		beneficiaryAnalyzer: beneficiaryAnalyzer,
		payerAnalyzer:       payerAnalyzer,
		logger:              logger,
	}
}

// BeneficiaryData represents a single beneficiary entry in the response
type BeneficiaryData struct {
	BeneficiaryAddress string                  `json:"beneficiary_address"`
	Amount             float64                 `json:"amount"`
	Transactions       []TransactionDetails    `json:"transactions"`
}

// PayerData represents a single payer entry in the response
type PayerData struct {
	PayerAddress     string               `json:"payer_address"`
	Amount           float64              `json:"amount"`
	Transactions     []TransactionDetails `json:"transactions"`
}

// TransactionDetails represents transaction details in the response
type TransactionDetails struct {
	TxAmount      float64 `json:"tx_amount"`
	DateTime      string  `json:"date_time"`
	TransactionID string  `json:"transaction_id"`
}

// BeneficiaryResponse represents the response format for the beneficiary endpoint
type BeneficiaryResponse struct {
	Message string          `json:"message"`
	Data    []BeneficiaryData `json:"data"`
}

// PayerResponse represents the response format for the payer endpoint
type PayerResponse struct {
	Message string     `json:"message"`
	Data    []PayerData `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

// HandleBeneficiary handles the /beneficiary endpoint
func (h *Handler) HandleBeneficiary(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		h.respondWithError(w, http.StatusBadRequest, "address parameter is required")
		return
	}

	h.logger.Infof("Analyzing beneficiaries for address: %s", address)

	beneficiaries, err := h.beneficiaryAnalyzer.AnalyzeBeneficiary(address)
	if err != nil {
		h.logger.Errorf("Error analyzing beneficiary: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert the analyzer.Beneficiary objects to BeneficiaryData objects
	responseData := make([]BeneficiaryData, len(beneficiaries))
	for i, b := range beneficiaries {
		txDetails := make([]TransactionDetails, len(b.Transactions))
		for j, tx := range b.Transactions {
			txDetails[j] = TransactionDetails{
				TxAmount:      tx.TxAmount,
				DateTime:      tx.DateTime,
				TransactionID: tx.TransactionID,
			}
		}
		
		responseData[i] = BeneficiaryData{
			BeneficiaryAddress: b.Address,
			Amount:             b.Amount,
			Transactions:       txDetails,
		}
	}

	h.respondWithJSON(w, http.StatusOK, BeneficiaryResponse{
		Message: "success",
		Data:    responseData,
	})
}

// HandlePayer handles the /payer endpoint
func (h *Handler) HandlePayer(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		h.respondWithError(w, http.StatusBadRequest, "address parameter is required")
		return
	}

	h.logger.Infof("Analyzing payers for address: %s", address)

	payers, err := h.payerAnalyzer.AnalyzePayer(address)
	if err != nil {
		h.logger.Errorf("Error analyzing payer: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert the analyzer.Payer objects to PayerData objects
	responseData := make([]PayerData, len(payers))
	for i, p := range payers {
		txDetails := make([]TransactionDetails, len(p.Transactions))
		for j, tx := range p.Transactions {
			txDetails[j] = TransactionDetails{
				TxAmount:      tx.TxAmount,
				DateTime:      tx.DateTime,
				TransactionID: tx.TransactionID,
			}
		}
		
		responseData[i] = PayerData{
			PayerAddress:     p.Address,
			Amount:           p.Amount,
			Transactions:     txDetails,
		}
	}

	h.respondWithJSON(w, http.StatusOK, PayerResponse{
		Message: "success",
		Data:    responseData,
	})
}

// respondWithJSON writes a JSON response
func (h *Handler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		h.logger.Errorf("Error marshaling response: %v", err)
		h.respondWithError(w, http.StatusInternalServerError, "Error creating response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// respondWithError writes an error response
func (h *Handler) respondWithError(w http.ResponseWriter, code int, message string) {
	h.respondWithJSON(w, code, ErrorResponse{
		Message: "error",
		Error:   message,
	})
}