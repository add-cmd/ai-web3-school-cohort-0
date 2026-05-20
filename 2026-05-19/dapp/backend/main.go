package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// ─── 配置 ─────────────────────────────────────────────────────

var (
	rpcURL       = getEnv("RPC_URL", "https://ethereum-sepolia-rpc.publicnode.com")
	serverPort   = getEnv("PORT", "8080")
	counterAddr  = os.Getenv("CONTRACT_ADDRESS")
	tokenAddr    = os.Getenv("TOKEN_ADDRESS")
	deepseekKey  = ""
)

// ─── 模型 ─────────────────────────────────────────────────────

type AnalyzeRequest struct {
	To          string `json:"to"`          // 目标合约
	Data        string `json:"data"`        // 函数调用数据 (0x...)
	Value       string `json:"value"`       // ETH 金额
	UserIntent  string `json:"user_intent"` // 用户意图（如"转账给我朋友"）
	Function    string `json:"function"`    // 函数名（方便前端传）
	Args        string `json:"args"`        // 参数（方便前端传）
}

type RiskAnalysis struct {
	Summary              string   `json:"summary"`
	AssetChanges         []AssetChange  `json:"asset_changes"`
	PermissionsChanged   []PermissionChange `json:"permissions_changed"`
	RiskLevel            string   `json:"risk_level"`
	RequiresHumanApproval bool   `json:"requires_human_approval"`
	Uncertainties        []string `json:"uncertainties"`
	RecommendedChecks    []string `json:"recommended_user_checks"`
	IntentMatch          bool     `json:"intent_match"`
	IntentNote           string   `json:"intent_note,omitempty"`
	Error                string   `json:"error,omitempty"`
}

type AssetChange struct {
	Asset     string `json:"asset"`
	Amount    string `json:"amount"`
	Direction string `json:"direction"` // in / out
}

type PermissionChange struct {
	Contract   string `json:"contract"`
	Permission string `json:"permission"`
	Detail     string `json:"detail"`
}

type ContractInfo struct {
	Address string `json:"address"`
	Network string `json:"network"`
	Count   uint64 `json:"count"`
	Owner   string `json:"owner"`
}

// ─── 工具 ─────────────────────────────────────────────────────

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func init() {
	home, _ := os.UserHomeDir()
	data, err := os.ReadFile(home + "/.hermes/.env")
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "DEEPSEEK_API_KEY=") {
				deepseekKey = strings.Trim(strings.TrimPrefix(line, "DEEPSEEK_API_KEY="), "\"'")
				break
			}
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// ─── 深度改造：风险分析 Prompt ─────────────────────────────

func analyzeWithDeepSeek(req AnalyzeRequest) *RiskAnalysis {
	if deepseekKey == "" {
		return &RiskAnalysis{Error: "DeepSeek API key not configured"}
	}

	prompt := `你是一个Web3交易风险分析专家。你的任务是分析用户即将签名的交易，评估风险。

## 角色定位
你是用户的最后一道防线。用户可能看不懂交易内容，你的职责是用人能读懂的语言解释风险。

## 任务目标
分析交易数据，输出结构化的风险摘要JSON。

## 可用输入
- to: 目标合约地址
- data: 十六进制函数调用数据，包含函数选择器（前4字节）和参数
- value: 发送的ETH数量
- user_intent: 用户描述的意图

## 已知合约信息
Counter (0x6d8521408b803813a1A963f511C74fB96ea23bd2):
- increment(): 增加计数器，选择器 0xd09de08a — 低风险
- decrement(): 减少计数器，选择器 0x2baeceb4 — 低风险
- setCount(uint256): 设置计数器值，选择器 0x7f58af4c — 低风险

SimpleToken (0x62E3395eCFa2d18afB8F0cfbB1FA55948Dd03674):
- transfer(address,uint256): 转账代币，选择器 0xa9059cbb — 中等风险
- approve(address,uint256): 授权额度，选择器 0x095ea7b3 — 高风险（特别是无限额度）
- transferFrom(address,address,uint256): 代扣转账，选择器 0x23b872dd — 高风险

## 风险分类规则
- low: 只读操作、标准函数调用、无资产转移
- medium: 有限额度的转账、有上下文的授权
- high: 大额转账、与意图不匹配的操作
- critical: 无限授权、transferFrom、未知合约、意图严重不匹配

## 禁止行为
- 不能忽略无限授权的风险
- 不能忽略地址与意图不匹配
- 不能假设用户已经检查过地址
- 不能在不确定时给出"安全"的判断

## 输出格式
必须严格按照以下JSON schema，不要添加额外字段：
{
  "summary": "一句话总结交易内容",
  "asset_changes": [
    {"asset": "资产名称或合约简称", "amount": "数量", "direction": "in/out"}
  ],
  "permissions_changed": [
    {"contract": "获得权限的地址", "permission": "授权类型", "detail": "具体额度"}
  ],
  "risk_level": "low | medium | high | critical",
  "requires_human_approval": true/false,
  "uncertainties": ["不确定因素列表"],
  "recommended_user_checks": ["用户应检查的项目"],
  "intent_match": true/false,
  "intent_note": "如果意图不匹配，解释为什么"
}`

	// 解析函数选择器
	selector := ""
	data := strings.TrimPrefix(req.Data, "0x")
	if len(data) >= 8 {
		selector = "0x" + data[:8]
	}

	txInfo := fmt.Sprintf(`## 交易数据
to: %s
data: %s
函数选择器: %s
value: %s ETH
用户意图: %s`,
		req.To, req.Data, selector, req.Value, req.UserIntent)

	payload := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{"role": "system", "content": prompt},
			{"role": "user", "content": txInfo},
		},
		"temperature": 0.3,
		"max_tokens":  1024,
	}
	body, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+deepseekKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return &RiskAnalysis{Error: fmt.Sprintf("DeepSeek call failed: %v", err)}
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
		return &RiskAnalysis{Error: "DeepSeek returned no choices"}
	}

	content := result.Choices[0].Message.Content
	// 提取 JSON（处理模型可能的 markdown 包裹）
	if idx := strings.Index(content, "{"); idx != -1 {
		content = content[idx:]
	}
	if idx := strings.LastIndex(content, "}"); idx != -1 {
		content = content[:idx+1]
	}

	var analysis RiskAnalysis
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		return &RiskAnalysis{Error: fmt.Sprintf("JSON parse error: %v\nRaw: %s", err, content[:min(200, len(content))])}
	}
	return &analysis
}

// ─── API 路由 ─────────────────────────────────────────────────

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		writeJSON(w, 405, RiskAnalysis{Error: "仅支持 POST"})
		return
	}

	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, 400, RiskAnalysis{Error: "请求体解析失败: " + err.Error()})
		return
	}

	if req.To == "" {
		writeJSON(w, 400, RiskAnalysis{Error: "缺少 to 字段"})
		return
	}

	log.Printf("🔍 分析请求: to=%s func=%s value=%s intent=%s",
		truncate(req.To, 10), req.Function, req.Value, req.UserIntent)

	analysis := analyzeWithDeepSeek(req)
	writeJSON(w, 200, analysis)
}

func handleExplain(w http.ResponseWriter, r *http.Request) {
	txHash := r.URL.Query().Get("tx")
	if txHash == "" || !strings.HasPrefix(txHash, "0x") {
		writeJSON(w, 400, map[string]string{"error": "请提供交易哈希"})
		return
	}

	txRaw, err := rpcCall("eth_getTransactionByHash", []interface{}{txHash})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": fmt.Sprintf("获取交易失败: %v", err)})
		return
	}
	receiptRaw, err := rpcCall("eth_getTransactionReceipt", []interface{}{txHash})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": fmt.Sprintf("获取回执失败: %v", err)})
		return
	}

	writeJSON(w, 200, map[string]interface{}{
		"tx":      txRaw,
		"receipt": receiptRaw,
	})
}

func rpcCall(method string, params []interface{}) (json.RawMessage, error) {
	payload := map[string]interface{}{
		"jsonrpc": "2.0", "method": method, "params": params, "id": 1,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(rpcURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	var result struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	json.Unmarshal(respBody, &result)
	if result.Error != nil {
		return nil, fmt.Errorf("RPC: %s", result.Error.Message)
	}
	return result.Result, nil
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]interface{}{
		"status":       "ok",
		"counter_addr": counterAddr,
		"token_addr":   tokenAddr,
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

// ─── 主函数 ───────────────────────────────────────────────────

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/analyze", handleAnalyze)
	mux.HandleFunc("/api/explain", handleExplain)
	mux.HandleFunc("/api/status", handleStatus)
	mux.HandleFunc("/api/health", handleHealth)

	handler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(204)
				return
			}
			h.ServeHTTP(w, r)
		})
	}

	log.Printf("📄 Counter: %s", counterAddr)
	log.Printf("📄 Token:   %s", tokenAddr)
	log.Printf("🤖 DeepSeek: %s", map[bool]string{true: "已配置", false: "未配置"}[deepseekKey != ""])
	log.Printf("🚀 后端启动在 :%s", serverPort)
	log.Println("   POST /api/analyze  — 交易风险分析（新功能）")
	log.Println("   GET  /api/explain  — 交易解释")
	log.Fatal(http.ListenAndServe(":"+serverPort, handler(mux)))
}
