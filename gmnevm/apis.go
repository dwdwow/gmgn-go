package gmneth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	BASE_URL       = "https://gmgn.ai/"
	EXACT_IN_PATH  = "defi/router/v1/tx/available_routes_exact_in"
	EXACT_OUT_PATH = "defi/router/v1/tx/available_routes_exact_out"
	SLIPPAGE_PATH  = "api/v1/recommend_slippage"
	GAS_PRICE_PATH = "defi/quotation/v1/chains"
)

type Resp[D any] struct {
	Msg  string `json:"msg"`  // Error message, e.g., "amountIn is required" - contains error details when the request fails
	Code int    `json:"code"` // 0 for success, -1 for error - status code indicating success (0) or failure (-1)
	Data D      `json:"data"`
}

func rest[T any](method string, path string, params any) (*T, error) {
	baseURL := fmt.Sprintf("%s%s", BASE_URL, path)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var req *http.Request
	var err error

	switch method {
	case http.MethodGet:
		// For GET requests, convert params to query parameters
		queryParams := url.Values{}

		if params != nil {
			// Convert params to JSON first
			jsonData, err := json.Marshal(params)
			if err != nil {
				return nil, fmt.Errorf("gmgn: failed to marshal params: %w", err)
			}

			// Convert JSON to map
			var paramMap map[string]interface{}
			if err := json.Unmarshal(jsonData, &paramMap); err != nil {
				return nil, fmt.Errorf("gmgn: failed to unmarshal params: %w", err)
			}

			// Add each parameter to query string
			for key, value := range paramMap {
				if value != nil {
					queryParams.Set(key, fmt.Sprintf("%v", value))
				}
			}
		}

		// Create full URL with query parameters
		fullURL := baseURL
		if len(queryParams) > 0 {
			fullURL = fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())
		}

		// Create GET request
		req, err = http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("gmgn: failed to create request: %w", err)
		}

	case http.MethodPost:
		// For POST requests, marshal params as JSON body
		var body []byte
		if params != nil {
			body, err = json.Marshal(params)
			if err != nil {
				return nil, fmt.Errorf("gmgn: failed to marshal params: %w", err)
			}
		}

		// Create POST request
		req, err = http.NewRequest("POST", baseURL, bytes.NewBuffer(body))
		if err != nil {
			return nil, fmt.Errorf("gmgn: failed to create request: %w", err)
		}

		// Set Content-Type for POST requests
		req.Header.Set("Content-Type", "application/json")
	}

	// Set common headers
	// req.Header.Set("Accept", "application/json")

	// Print URL for debugging
	fmt.Printf("Making request to: %s\n", req.URL.String())

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gmgn: failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gmgn: failed to read response body: %w", err)
	}

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gmgn: API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var result Resp[*T]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("gmgn: failed to unmarshal response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("gmgn: API request failed with code %d: %s", result.Code, result.Msg)
	}

	return result.Data, nil
}

type ExactInParams struct {
	TokenInChain    string `json:"token_in_chain"`    // Network code like eth/bsc/arb - specifies the blockchain network for the input token
	TokenOutChain   string `json:"token_out_chain"`   // Network code like eth/bsc/arb - specifies the blockchain network for the output token
	TokenInAddress  string `json:"token_in_address"`  // Input token contract address - the smart contract address of the token to be swapped from
	TokenOutAddress string `json:"token_out_address"` // Output token contract address - the smart contract address of the token to be swapped to
	FromAddress     string `json:"from_address"`      // Wallet address initiating the transaction - address that will execute the swap
	InAmount        string `json:"in_amount"`         // Input amount in smallest unit - the amount of input token to swap (in wei or smallest unit)
	Src             string `json:"src,omitempty"`     // Source, can be gmgn or swapx, default gmgn - specifies the data source for the swap
}

type ExactOutParams struct {
	TokenInChain    string `json:"token_in_chain"`    // Network code like eth/bsc/arb - specifies the blockchain network for the input token
	TokenOutChain   string `json:"token_out_chain"`   // Network code like eth/bsc/arb - specifies the blockchain network for the output token
	TokenInAddress  string `json:"token_in_address"`  // Input token contract address - the smart contract address of the token to be swapped from
	TokenOutAddress string `json:"token_out_address"` // Output token contract address - the smart contract address of the token to be swapped to
	// FromAddress     string `json:"from_address"`      // Wallet address initiating the transaction - address that will execute the swap
	OutAmount string `json:"out_amount"`    // Output amount in smallest unit - the desired amount of output token to receive (in wei or smallest unit)
	Src       string `json:"src,omitempty"` // Source, can be gmgn or swapx, default gmgn - specifies the data source for the swap
}

// ExactInOutData contains the basic information of cross-chain transactions
type ExactInOutData struct {
	Routes       []Route      `json:"routes"`       // List of routes - array of available swap routes with different DEX protocols
	Volatilities Volatilities `json:"volatilities"` // Price volatility information - contains price volatility data for input and output tokens
}

// Route represents a single route in the response
type Route struct {
	ChainID            int    `json:"chain_id"`                      // Source chain ID - blockchain network identifier (e.g., 1 for Ethereum mainnet)
	To                 string `json:"to"`                            // Contract address - destination smart contract address for the swap
	AmountIn           string `json:"amount_in"`                     // Source token amount - amount of input token to be swapped (in smallest unit)
	AmountIn2          string `json:"amount_in2,omitempty"`          // Additional amount field - secondary amount field for complex swaps
	AmountOut          string `json:"amount_out"`                    // Output token amount - expected amount of output token to be received
	InputTokenAddress  string `json:"input_token_address"`           // Input token contract address - smart contract address of the token being swapped from
	OutputTokenAddress string `json:"output_token_address"`          // Output token contract address - smart contract address of the token being swapped to
	Type               string `json:"type"`                          // Route type: v0, v2, v3, v2-v2, v2-v3, v3-v2 - DEX protocol version and routing type
	Path               any    `json:"path"`                          // Path as string or array of strings - token swap path through liquidity pools
	PathBytes          string `json:"path_bytes,omitempty"`          // Path bytes for V3 multi-hop - encoded path for Uniswap V3 multi-hop swaps
	PoolAddress        any    `json:"pool_address"`                  // Pool address as string or array - liquidity pool addresses used in the swap
	FactoryAddress     any    `json:"factory_address"`               // Factory address as string or array - DEX factory contract addresses
	Fee                any    `json:"fee"`                           // Fee as string or array (V3 uses this) - trading fees for the swap (V3 uses tiered fees)
	Steps              []Step `json:"steps"`                         // Steps list with id/type/tool fields - detailed steps of the swap process
	TokenInUsdPrice    string `json:"token_in_usd_price,omitempty"`  // Input token USD price - current USD price of the input token
	AmountInUsd        string `json:"amount_in_usd,omitempty"`       // Input amount in USD - USD value of the input amount
	TokenOutUsdPrice   string `json:"token_out_usd_price,omitempty"` // Output token USD price - current USD price of the output token
	AmountOutUsd       string `json:"amount_out_usd,omitempty"`      // Output amount in USD - USD value of the expected output amount
	Value              string `json:"value"`                         // Value field - transaction value or additional value parameter
	PriceImpact        string `json:"price_impact,omitempty"`        // Price impact - percentage price impact of the swap on the market
	GasLimit           string `json:"gas_limit,omitempty"`           // Gas limit - estimated gas limit required for the transaction
	FromAddress        string `json:"from_address,omitempty"`        // Wallet address initiating the transaction - address that will execute the swap
}

// Step represents a step in the route with id, type, and tool fields
type Step struct {
	ID   int    `json:"id"`   // Step ID - unique identifier for this step in the swap process
	Type string `json:"type"` // Step type, e.g., "swap" - type of operation performed in this step
	Tool string `json:"tool"` // Tool used, e.g., "uniswapv2" - specific DEX protocol or tool used for this step
}

// Volatilities represents price volatility information
type Volatilities struct {
	TokenIn  int  `json:"token_in"`  // 5-minute price volatility % for input token - price volatility percentage of the input token over 5 minutes
	TokenOut int  `json:"token_out"` // 5-minute price volatility % for output token - price volatility percentage of the output token over 5 minutes
	IsFomo   bool `json:"is_fomo"`   // FOMO indicator (true/false) - indicates if there's fear of missing out sentiment affecting the token prices
}

// ExactIn makes a request to the GMGN API to get available swap routes for exact input amount
func ExactIn(params ExactInParams) (*ExactInOutData, error) {
	return rest[ExactInOutData](http.MethodGet, EXACT_IN_PATH, params)
}

// ExactOut makes a request to the GMGN API to get available swap routes for exact output amount
func ExactOut(params ExactOutParams) (*ExactInOutData, error) {
	return rest[ExactInOutData](http.MethodGet, EXACT_OUT_PATH, params)
}

type SlippageParams struct {
	TokenAddress   string `json:"token_address"`    // Output token contract address - the smart contract address of the output token
	TokenInAddress string `json:"token_in_address"` // Input token contract address - optional, mainly for solving slippage accumulation issues when swapping between two altcoins
}

type SlippageData struct {
	RecommendSlippage string `json:"recommend_slippage"` // Recommended slippage value - e.g., "1" for 1%
	DisplaySlippage   string `json:"display_slippage"`   // Display slippage value - the slippage value to show to users
	HasTax            bool   `json:"has_tax"`            // Tax indicator - whether the token has tax mechanism
}

type GasPriceData struct {
	LastBlock      int64  `json:"last_block"`       // Last block number - the most recent block number
	Average        string `json:"average"`          // Average gas price - average gas price in wei
	High           string `json:"high"`             // High gas price - highest gas price in wei
	Low            string `json:"low"`              // Low gas price - lowest gas price in wei
	SuggestBaseFee string `json:"suggest_base_fee"` // Suggested base fee - recommended base fee for transactions
	EthUsdPrice    string `json:"eth_usd_price"`    // ETH USD price - current ETH price in USD
	HighPrioFee    string `json:"high_prio_fee"`    // High priority fee - fee for high priority transactions
	AveragePrioFee string `json:"average_prio_fee"` // Average priority fee - average priority fee for transactions
	LowPrioFee     string `json:"low_prio_fee"`     // Low priority fee - fee for low priority transactions
}

// Slippage makes a request to the GMGN API to get recommended slippage for a token
func Slippage(params SlippageParams) (*SlippageData, error) {
	// Build the slippage path with chain and token address
	return rest[SlippageData](http.MethodGet, SLIPPAGE_PATH, params)
}

// GasPrice makes a request to the GMGN API to get gas price information for a specific network
func GasPrice(network string) (*GasPriceData, error) {
	// Build the gas price path with network
	path := fmt.Sprintf("%s/%s/gas_price", GAS_PRICE_PATH, network)
	return rest[GasPriceData](http.MethodGet, path, nil)
}
