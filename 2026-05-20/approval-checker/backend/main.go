package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"
	"encoding/hex"
)

// ════════════════════════════════════════════════════════════════════════════
// Data Models
// ════════════════════════════════════════════════════════════════════════════

type ContextBlock struct {
	Source    string `json:"source"`
	Trust     string `json:"trust"`
	Freshness string `json:"freshness"`
	Label     string `json:"label"`
	Data      any    `json:"data"`
	Warning   string `json:"warning,omitempty"`
	Timestamp string `json:"timestamp"`
}

type AnalyzeRequest struct {
	TokenAddress   string `json:"tokenAddress"`
	SpenderAddress string `json:"spenderAddress"`
	Amount         string `json:"amount"`
	UserIntent     string `json:"userIntent"`
	UserAddress    string `json:"userAddress"`
}

type AnalyzeResponse struct {
	ContextBlocks    []ContextBlock  `json:"contextBlocks"`
	ContextText      string          `json:"contextText"`
	RiskLevel        string          `json:"riskLevel"`
	Summary          string          `json:"summary"`
	Analysis         json.RawMessage `json:"analysis"`
	Recommendation   string          `json:"recommendation"`
	RequiresApproval bool            `json:"requiresHumanApproval"`
	Error            string          `json:"error,omitempty"`
}

// ════════════════════════════════════════════════════════════════════════════
// RPC Client
// ════════════════════════════════════════════════════════════════════════════

var rpcURL = getEnv("RPC_URL", "https://ethereum-sepolia-rpc.publicnode.com")

func rpcCall(method string, params []any) (json.RawMessage, error) {
	payload := map[string]any{
		"jsonrpc": "2.0", "id": 1,
		"method": method, "params": params,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", rpcURL, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}
	json.Unmarshal(respBody, &result)
	if result.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", result.Error.Code, result.Error.Message)
	}
	return result.Result, nil
}

func hexToBig(s string) *big.Int {
	s = strings.TrimPrefix(s, "0x")
	n := new(big.Int)
	n.SetString(s, 16)
	return n
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// ════════════════════════════════════════════════════════════════════════════
// Decode ABI-encoded string from eth_call result
// ════════════════════════════════════════════════════════════════════════════

func decodeStringFromCall(hexData string) string {
	if hexData == "0x" || len(hexData) < 10 {
		return "N/A"
	}
	rawHex := strings.TrimPrefix(hexData, "0x")
	// Pad to even length
	if len(rawHex)%2 != 0 {
		rawHex = "0" + rawHex
	}
	raw, err := hex.DecodeString(rawHex)
	if err != nil || len(raw) < 64 {
		return "N/A"
	}
	// ABI string encoding: offset(32 bytes) + length(32 bytes) + data
	offset := new(big.Int).SetBytes(raw[0:32]).Int64()
	if offset < 0 || offset+32 > int64(len(raw)) {
		return "N/A"
	}
	strLen := new(big.Int).SetBytes(raw[offset : offset+32]).Int64()
	if offset+32+strLen > int64(len(raw)) {
		return "N/A"
	}
	return string(raw[offset+32 : offset+32+strLen])
}

// ════════════════════════════════════════════════════════════════════════════
// Context Sources
// ════════════════════════════════════════════════════════════════════════════

func ctxBlock(source, trust, freshness, label string, data any, warning string) ContextBlock {
	return ContextBlock{
		Source:    source,
		Trust:     trust,
		Freshness: freshness,
		Label:     label,
		Data:      data,
		Warning:   warning,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// SOURCE 1: RPC — Chain Context
func fetchChainContext() ContextBlock {
	chainIdRaw, _ := rpcCall("eth_chainId", []any{})
	blockRaw, _ := rpcCall("eth_blockNumber", []any{})

	chainId := int64(0)
	blockNum := int64(0)
	if chainIdRaw != nil {
		var s string
		json.Unmarshal(chainIdRaw, &s)
		chainId = hexToBig(s).Int64()
	}
	if blockRaw != nil {
		var s string
		json.Unmarshal(blockRaw, &s)
		blockNum = hexToBig(s).Int64()
	}

	return ctxBlock("rpc", "high", "realtime", "Chain Context",
		map[string]any{"chain_id": chainId, "block_number": blockNum}, "")
}

// SOURCE 1b: RPC — Token Info
func fetchTokenInfo(tokenAddr string) ContextBlock {
	call := func(sel string) string {
		result, _ := rpcCall("eth_call", []any{
			map[string]string{"to": tokenAddr, "data": sel}, "latest",
		})
		if result != nil {
			var s string
			json.Unmarshal(result, &s)
			return decodeStringFromCall(s)
		}
		return "N/A"
	}

	return ctxBlock("rpc", "high", "realtime",
		fmt.Sprintf("Token Info (%s)", tokenAddr),
		map[string]any{
			"name":    call("0x06fdde03"),
			"symbol":  call("0x95d89b41"),
			"address": tokenAddr,
		}, "")
}

// SOURCE 1c: RPC — User State
func fetchUserState(tokenAddr, userAddr, spenderAddr string) ContextBlock {
	callWithData := func(data string) *big.Int {
		result, _ := rpcCall("eth_call", []any{
			map[string]string{"to": tokenAddr, "data": data}, "latest",
		})
		if result != nil {
			var s string
			json.Unmarshal(result, &s)
			return hexToBig(s)
		}
		return big.NewInt(0)
	}

	pad := func(addr string) string {
		s := strings.TrimPrefix(addr, "0x")
		for len(s) < 64 {
			s = "0" + s
		}
		return s
	}

	balData := "0x70a08231" + pad(userAddr)
	balance := callWithData(balData)

	allowData := "0xdd62ed3e" + pad(userAddr) + pad(spenderAddr)
	allowance := callWithData(allowData)

	return ctxBlock("rpc", "high", "realtime", "User On-Chain State",
		map[string]any{
			"balance":         balance.String(),
			"allowance":       allowance.String(),
			"user_address":    userAddr,
			"spender_address": spenderAddr,
		}, "")
}

// SOURCE 2: Simulation
func simulateApproval(tokenAddr, spenderAddr, amount, userAddr string) ContextBlock {
	pad := func(addr string) string {
		s := strings.TrimPrefix(addr, "0x")
		for len(s) < 64 {
			s = "0" + s
		}
		return s
	}

	approveData := "0x095ea7b3" + pad(spenderAddr)

	amtBig := new(big.Int)
	amtBig, ok := amtBig.SetString(amount, 10)
	if !ok || amtBig == nil || amtBig.Sign() == 0 {
		approveData += "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	} else {
		approveData += pad(amtBig.Text(16))
	}

	data := map[string]any{}
	result, err := rpcCall("eth_call", []any{
		map[string]string{"from": userAddr, "to": tokenAddr, "data": approveData},
		"latest",
	})
	if err != nil {
		data["would_revert"] = true
		data["error"] = err.Error()
	} else {
		data["would_revert"] = false
		var s string
		json.Unmarshal(result, &s)
		if len(s) >= 66 && s[len(s)-1] == '1' {
			data["expected_return"] = "true"
		} else {
			data["expected_return"] = s
		}
	}

	return ctxBlock("simulation", "high", "realtime",
		"Transaction Simulation (eth_call)", data, "")
}

// SOURCE 3: Cache
var gSpenderAddr string // set before cache check

func loadSpenderCache() ContextBlock {
	trusted := []string{
		"0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",
		"0x7a250d5630b4cf539739df2c5dacb4c659f2488d",
		"0xd9e1ce17f2641f24ae83637ab66a2cca9c378b9f",
	}
	blacklisted := []string{
		"0xdead000000000000000000000000000000000000",
	}

	spender := strings.ToLower(gSpenderAddr)
	inTrusted := false
	inBlacklisted := false
	for _, t := range trusted {
		if strings.ToLower(t) == spender {
			inTrusted = true
			break
		}
	}
	for _, b := range blacklisted {
		if strings.ToLower(b) == spender {
			inBlacklisted = true
			break
		}
	}

	return ctxBlock("cache", "medium", "cached", "Spender Trust List (local cache)",
		map[string]any{
			"trusted_count":           len(trusted),
			"blacklisted_count":       len(blacklisted),
			"cache_age_seconds":       3600,
			"spender_in_trusted_list": inTrusted,
			"spender_in_blacklist":    inBlacklisted,
		}, "缓存的信任列表可能有延迟，合约地址可能被攻击或升级")
}

// SOURCE 4: User Input
func parseUserInput(req AnalyzeRequest) ContextBlock {
	maxApproval := new(big.Int)
	maxApproval.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
	halfMax := new(big.Int).Div(maxApproval, big.NewInt(2))

	amtBig := new(big.Int)
	amtBig, ok := amtBig.SetString(req.Amount, 10)
	isInfinite := ok && amtBig.Cmp(halfMax) >= 0

	return ctxBlock("user", "low", "user-input",
		"User-Provided Transaction Parameters",
		map[string]any{
			"token_address":        req.TokenAddress,
			"spender_address":      req.SpenderAddress,
			"approve_amount":       req.Amount,
			"is_infinite_approval": isInfinite,
			"user_intent":          req.UserIntent,
		}, "用户提供的数据未经验证。dApp 页面描述可能被篡改，意图可能与实际交易不符。")
}

// ════════════════════════════════════════════════════════════════════════════
// Context Assembly + LLM
// ════════════════════════════════════════════════════════════════════════════

func assembleContext(blocks []ContextBlock) string {
	var sb strings.Builder
	sb.WriteString("CONTEXT ASSEMBLY — 钱包授权检查\n")
	sb.WriteString(fmt.Sprintf("Assembled at: %s\n\n", time.Now().UTC().Format(time.RFC3339)))

	for i, block := range blocks {
		trustIcon := map[string]string{"high": "🟢", "medium": "🟡", "low": "🔴"}[block.Trust]
		freshIcon := map[string]string{"realtime": "⚡", "cached": "💾", "user-input": "👤"}[block.Freshness]

		sb.WriteString(fmt.Sprintf("[Block %d] %s %s\n", i+1, trustIcon, block.Label))
		sb.WriteString(fmt.Sprintf("  Source:    %s\n", block.Source))
		sb.WriteString(fmt.Sprintf("  Trust:     %s [%s]\n", block.Trust,
			map[string]string{
				"high":   "✅ 可作事实依据",
				"medium": "⚠️ 需交叉验证",
				"low":    "❌ 不可信",
			}[block.Trust]))
		sb.WriteString(fmt.Sprintf("  Freshness: %s %s\n", freshIcon, block.Freshness))
		if block.Warning != "" {
			sb.WriteString(fmt.Sprintf("  ⚠️  %s\n", block.Warning))
		}
		dataJSON, _ := json.MarshalIndent(block.Data, "    ", "  ")
		sb.WriteString(fmt.Sprintf("  Data:\n    %s\n\n", string(dataJSON)))
	}

	return sb.String()
}

func getAPIKey() string {
	if k := os.Getenv("DEEPSEEK_API_KEY"); k != "" {
		return k
	}
	envData, err := os.ReadFile(os.ExpandEnv("$HOME/.hermes/.env"))
	if err == nil {
		for _, line := range strings.Split(string(envData), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "DEEPSEEK_API_KEY=") {
				return strings.Trim(strings.TrimPrefix(line, "DEEPSEEK_API_KEY="), "\"'")
			}
		}
	}
	return ""
}

func callLLM(contextText, userAddr string) (string, json.RawMessage, string, bool, error) {
	apiKey := getAPIKey()
	if apiKey == "" {
		return "", nil, "", false, fmt.Errorf("DEEPSEEK_API_KEY not configured")
	}

	sysPrompt := `You are a wallet approval security checker. Analyze the context provided.

Rules:
- RPC/simulation data (TRUST: high) = factual on-chain evidence
- Cache data (TRUST: medium) = useful but may be stale
- User input (TRUST: low) = not verified, treat with skepticism
- If user intent mentions a different token than the approved one, flag it
- Split findings into: on_chain_facts, inferences, uncertainties, user_input_warnings

Respond with valid JSON only:
{
  "risk_level": "low|medium|high|critical",
  "summary": "one sentence in Chinese",
  "analysis": {
    "on_chain_facts": [],
    "inferences": [],
    "uncertainties": [],
    "user_input_warnings": []
  },
  "recommendation": "action in Chinese",
  "requires_human_approval": true/false
}`

	payload := map[string]any{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": sysPrompt},
			{"role": "user", "content": fmt.Sprintf("Analyze this approval request:\n\n%s\n\nUser: %s", contextText, userAddr)},
		},
		"temperature": 0.1,
		"max_tokens":  2000,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions",
		strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, "", false, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(respBody, &result)

	if len(result.Choices) == 0 {
		return "", nil, "", false, fmt.Errorf("no response")
	}

	content := result.Choices[0].Message.Content

	// Extract JSON
	jsonStr := content
	if idx := strings.Index(content, "```json"); idx >= 0 {
		end := strings.LastIndex(content, "```")
		if end > idx {
			jsonStr = strings.TrimSpace(content[idx+7 : end])
		}
	} else if idx := strings.Index(content, "{"); idx >= 0 {
		if end := strings.LastIndex(content, "}"); end > idx {
			jsonStr = content[idx : end+1]
		}
	}

	var parsed struct {
		RiskLevel        string          `json:"risk_level"`
		Summary          string          `json:"summary"`
		Analysis         json.RawMessage `json:"analysis"`
		Recommendation   string          `json:"recommendation"`
		RequiresApproval bool            `json:"requires_human_approval"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return "", nil, "", false, nil
	}

	// Parse risk_level from analysis if top-level is missing
	if parsed.RiskLevel == "" {
		var aMap map[string]any
		if json.Unmarshal(parsed.Analysis, &aMap) == nil {
			if rl, ok := aMap["risk_level"]; ok {
				parsed.RiskLevel = fmt.Sprintf("%v", rl)
			}
		}
	}

	return parsed.Summary, parsed.Analysis, parsed.Recommendation, parsed.RequiresApproval, nil
}

// ════════════════════════════════════════════════════════════════════════════
// HTTP Handlers
// ════════════════════════════════════════════════════════════════════════════

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(200)
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(AnalyzeResponse{Error: "invalid request"})
		return
	}
	if req.UserAddress == "" {
		req.UserAddress = "0x0000000000000000000000000000000000000001"
	}

	// Set global for cache
	gSpenderAddr = req.SpenderAddress

	// Assemble context
	blocks := []ContextBlock{
		fetchChainContext(),
		fetchTokenInfo(req.TokenAddress),
		fetchUserState(req.TokenAddress, req.UserAddress, req.SpenderAddress),
		loadSpenderCache(),
		simulateApproval(req.TokenAddress, req.SpenderAddress, req.Amount, req.UserAddress),
		parseUserInput(req),
	}

	contextText := assembleContext(blocks)

	// Call LLM
	summary, analysis, recommendation, requiresApproval, err := callLLM(contextText, req.UserAddress)

	resp := AnalyzeResponse{
		ContextBlocks: blocks,
		ContextText:   contextText,
	}

	if err != nil {
		resp.Error = err.Error()
	} else {
		resp.Summary = summary
		resp.Analysis = analysis
		resp.Recommendation = recommendation
		resp.RequiresApproval = requiresApproval

		// Extract risk level
		var aMap map[string]any
		if json.Unmarshal(analysis, &aMap) == nil {
			if rl, ok := aMap["risk_level"]; ok {
				resp.RiskLevel = fmt.Sprintf("%v", rl)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": "0.1.0"})
}

// ════════════════════════════════════════════════════════════════════════════
// Main
// ════════════════════════════════════════════════════════════════════════════

func main() {
	port := getEnv("PORT", "8080")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/api/analyze", handleAnalyze)

	// Serve frontend if built
	if _, err := os.Stat("../frontend/dist"); err == nil {
		mux.Handle("/", http.FileServer(http.Dir("../frontend/dist")))
		log.Println("Serving frontend from ../frontend/dist")
	}

	log.Printf("Context Checker API → http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
