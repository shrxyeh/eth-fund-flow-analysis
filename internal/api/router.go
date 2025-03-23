package api

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/shrxyeh/ethereum-fund-flow/internal/analyzer"
	"github.com/shrxyeh/ethereum-fund-flow/pkg/logger"
)

// Router is the HTTP router
type Router struct {
	handler        *Handler
	logger         logger.Logger
	defaultAddress string
	analysisMode   string
}

// NewRouter creates a new router
func NewRouter(beneficiaryAnalyzer *analyzer.BeneficiaryAnalyzer, payerAnalyzer *analyzer.PayerAnalyzer, logger logger.Logger) *Router {
	handler := NewHandler(beneficiaryAnalyzer, payerAnalyzer, logger)
	return &Router{
		handler:        handler,
		logger:         logger,
		defaultAddress: "",
		analysisMode:   "both",
	}
}

// SetDefaultAddress sets the default Ethereum address for examples
func (r *Router) SetDefaultAddress(address string) {
	r.defaultAddress = address
}

// SetAnalysisMode sets the analysis mode
func (r *Router) SetAnalysisMode(mode string) {
	r.analysisMode = mode
}

// Setup sets up the HTTP routes
func (r *Router) Setup() *mux.Router {
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/beneficiary", r.handler.HandleBeneficiary).Methods("GET")
	router.HandleFunc("/payer", r.handler.HandlePayer).Methods("GET")

	// Root handler
	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// Use the command-line address for examples if available
		exampleAddress := r.defaultAddress
		if exampleAddress == "" {
			exampleAddress = "0xYourEthereumAddressHere"
		}

		// Define the additional info for current address if it exists
		var addressInfo string
		if r.defaultAddress != "" {
			addressInfo = fmt.Sprintf("<p>Current address: <code>%s</code></p>", r.defaultAddress)
		} else {
			addressInfo = ""
		}

		html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Ethereum Fund Flow Analysis API</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            margin: 0;
            padding: 20px;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
        }
        h1 {
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
        }
        h2 {
            color: #2980b9;
        }
        pre {
            background-color: #f8f8f8;
            border: 1px solid #ddd;
            padding: 10px;
            border-radius: 4px;
            overflow-x: auto;
        }
        .endpoint {
            background-color: #e9f7fe;
            border-left: 3px solid #3498db;
            padding: 15px;
            margin: 15px 0;
        }
        .example {
            font-family: monospace;
            background-color: #f5f5f5;
            padding: 10px;
            border-radius: 4px;
            margin: 10px 0;
        }
        a {
            color: #3498db;
            text-decoration: none;
        }
        a:hover {
            text-decoration: underline;
        }
        .button {
            display: inline-block;
            padding: 10px 15px;
            background-color: #3498db;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin: 10px 0;
        }
        .info {
            background-color: #f8f9fa;
            border-left: 5px solid #3498db;
            padding: 15px;
            margin: 15px 0;
        }
    </style>
</head>
<body>
    <h1>Ethereum Fund Flow Analysis API</h1>
    <p>This API analyzes the flow of funds for Ethereum addresses to determine beneficiary addresses and payers.</p>
    
    <div class="info">
        <p>Current analysis mode: <strong>%s</strong></p>
        %s
    </div>
    
    <h2>Endpoints</h2>
    
    <div class="endpoint">
        <h3>Health Check</h3>
        <p>Check if the API is running:</p>
        <div class="example">
            <a href="/health" target="_blank">/health</a>
        </div>
    </div>
    
    <div class="endpoint">
        <h3>Beneficiary Analysis</h3>
        <p>Identifies where funds are flowing to from a given address:</p>
        <div class="example">
            /beneficiary?address=&lt;ethereum_address&gt;
        </div>
        <p>Example:</p>
        <div class="example">
            <a href="/beneficiary?address=%s" target="_blank">/beneficiary?address=%s</a>
        </div>
    </div>
    
    <div class="endpoint">
        <h3>Payer Analysis</h3>
        <p>Identifies where funds are coming from to a given address:</p>
        <div class="example">
            /payer?address=&lt;ethereum_address&gt;
        </div>
        <p>Example:</p>
        <div class="example">
            <a href="/payer?address=%s" target="_blank">/payer?address=%s</a>
        </div>
    </div>
    
    <h2>Sample Ethereum Addresses for Testing</h2>
    <ul>
        <li><code>0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2</code> (WETH Contract)</li>
        <li><code>0x7a250d5630b4cf539739df2c5dacb4c659f2488d</code> (Uniswap V2 Router)</li>
        <li><code>0xb8901acb165ed027e32754e0ffe830802919727f</code> (Sample address)</li>
    </ul>
    
    <h2>Command Line Usage</h2>
    <pre>
  ./bin/api -address=0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2 -mode=beneficiary
  ./bin/api -address=0x7a250d5630b4cf539739df2c5dacb4c659f2488d -mode=payer
  ./bin/api -help</pre>
</body>
</html>
`, r.analysisMode, addressInfo, exampleAddress, exampleAddress, exampleAddress, exampleAddress)

		tmpl, err := template.New("home").Parse(html)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		tmpl.Execute(w, nil)
	}).Methods("GET")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Analyze default route
	if r.defaultAddress != "" {
		router.HandleFunc("/analyze-default", func(w http.ResponseWriter, req *http.Request) {
			// Generate the HTML buttons based on analysis mode
			var beneficiaryButton, payerButton string
			
			if r.analysisMode == "beneficiary" || r.analysisMode == "both" {
				beneficiaryButton = fmt.Sprintf(`<a href="/beneficiary?address=%s" class="button">Analyze Beneficiaries</a>`, r.defaultAddress)
			}
			
			if r.analysisMode == "payer" || r.analysisMode == "both" {
				payerButton = fmt.Sprintf(`<a href="/payer?address=%s" class="button">Analyze Payers</a>`, r.defaultAddress)
			}

			html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Analyze Ethereum Address</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            line-height: 1.6;
            max-width: 800px;
            margin: 0 auto;
        }
        .button {
            display: inline-block;
            padding: 10px 15px;
            background-color: #3498db;
            color: white;
            text-decoration: none;
            border-radius: 4px;
            margin: 10px 5px;
        }
        .info {
            background-color: #f8f9fa;
            border-left: 5px solid #3498db;
            padding: 15px;
            margin: 15px 0;
        }
    </style>
</head>
<body>
    <h1>Analyze Ethereum Address</h1>
    
    <div class="info">
        <p>Ethereum address: <code>%s</code></p>
        <p>Analysis mode: <code>%s</code></p>
    </div>
    
    <p>Click below to analyze this address:</p>
    
    %s
    %s
    
    <p><a href="/">Back to Home</a></p>
</body>
</html>
`, r.defaultAddress, r.analysisMode, beneficiaryButton, payerButton)

			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(html))
		}).Methods("GET")
	}

	// Log requests
	router.Use(r.loggingMiddleware)

	return router
}

// loggingMiddleware logs HTTP requests
func (r *Router) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.logger.Infof("Request: %s %s", req.Method, req.URL.Path)
		next.ServeHTTP(w, req)
	})
}