package eth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"time"
)

type Client struct {
	http   *http.Client
	rpcURL string
}

func NewInfuraClient(apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("INFURA_API_KEY is required")
	}
	return &Client{
		http: &http.Client{Timeout: 10 * time.Second},
		rpcURL: "https://mainnet.infura.io/v3/" + apiKey,
	}, nil
}

type rpcReq struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type rpcResp struct {
	Result string `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) GetBalanceETH(ctx context.Context, address string, block string) (string, error) {
	if block == "" {
		block = "latest"
	}

	body, _ := json.Marshal(rpcReq{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_getBalance",
		Params:  []interface{}{address, block},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.rpcURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out rpcResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", errors.New(out.Error.Message)
	}

	wei := new(big.Int)
	if _, ok := wei.SetString(out.Result[2:], 16); !ok { // strip 0x
		return "", errors.New("invalid hex balance")
	}

	// ETH = wei / 1e18 as a decimal string (no float)
	r := new(big.Rat).SetInt(wei)
	r.Quo(r, new(big.Rat).SetInt(big.NewInt(1_000_000_000_000_000_000)))
	s := r.FloatString(18)
	s = trimTrailingZeros(s)
	return s, nil
}

func trimTrailingZeros(s string) string {
	if len(s) == 0 {
		return s
	}
	// if no dot, return
	dot := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			dot = i
			break
		}
	}
	if dot == -1 {
		return s
	}
	// trim zeros
	i := len(s) - 1
	for i > dot && s[i] == '0' {
		i--
	}
	// trim dot if needed
	if i == dot {
		i--
	}
	return s[:i+1]
}
