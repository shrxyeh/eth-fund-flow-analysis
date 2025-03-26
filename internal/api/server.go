package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shrxyeh/ethereum-fund-flow/internal/analyzer"
	"github.com/shrxyeh/ethereum-fund-flow/internal/config"
	"github.com/shrxyeh/ethereum-fund-flow/internal/etherscan"
	"github.com/shrxyeh/ethereum-fund-flow/pkg/logger"
)

// Server is the API server
type Server struct {
	config       *config.Config
	logger       logger.Logger
	router       *Router
	defaultAddr  string
	analysisMode string
}

// NewServer creates a new server
func NewServer(config *config.Config, logger logger.Logger) *Server {
	// Create Etherscan client
	etherscanClient := etherscan.NewClient(config.EtherscanAPIKey)

	// Create analyzers
	beneficiaryAnalyzer := analyzer.NewBeneficiaryAnalyzer(etherscanClient)
	payerAnalyzer := analyzer.NewPayerAnalyzer(etherscanClient)

	// Create router
	router := NewRouter(beneficiaryAnalyzer, payerAnalyzer, logger)

	return &Server{
		config:       config,
		logger:       logger,
		router:       router,
		defaultAddr:  "",
		analysisMode: "both",
	}
}

// SetDefaultAddress sets the default Ethereum address for analysis
func (s *Server) SetDefaultAddress(address string) {
	s.defaultAddr = address
}

// SetAnalysisMode sets the analysis mode (beneficiary, payer, or both)
func (s *Server) SetAnalysisMode(mode string) {
	s.analysisMode = mode
}

// Start starts the server
func (s *Server) Start() error {
	// Pass default address and mode to the router
	s.router.SetDefaultAddress(s.defaultAddr)
	s.router.SetAnalysisMode(s.analysisMode)
	
	// Setup the router
	r := s.router.Setup()
	
	addr := fmt.Sprintf(":%s", s.config.Port)
	s.logger.Infof("Starting server with address: %s and mode: %s", s.defaultAddr, s.analysisMode)
	return http.ListenAndServe(addr, r)
}