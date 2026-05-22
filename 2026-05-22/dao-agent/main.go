package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ═══════════════════════════════════════════════════════════════
// Data Models
// ═══════════════════════════════════════════════════════════════

type Chunk struct {
	Source  string `json:"source"`
	Heading string `json:"heading"`
	Text    string `json:"text"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type Message struct {
	Role         string     `json:"role"`
	Content      string     `json:"content,omitempty"`
	ToolCalls    []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID   string     `json:"tool_call_id,omitempty"`
	Name         string     `json:"name,omitempty"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []ToolDef `json:"tools,omitempty"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
}

type ToolDef struct {
	Type     string   `json:"type"`
	Function FuncSpec `json:"function"`
}

type FuncSpec struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  ParamObj   `json:"parameters"`
}

type ParamObj struct {
	Type       string              `json:"type"`
	Properties map[string]ParamDef `json:"properties"`
	Required   []string            `json:"required"`
}

type ParamDef struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ═══════════════════════════════════════════════════════════════
// Tools
// ═══════════════════════════════════════════════════════════════

func readProposal() string {
	data, err := os.ReadFile("proposal_governance_ai.txt")
	if err != nil {
		return "读取提案失败: " + err.Error()
	}
	return string(data)
}

func searchHandbook(query string, topK int) string {
	data, err := os.ReadFile("chunks.json")
	if err != nil {
		return "读取知识库失败: " + err.Error()
	}
	var chunks []Chunk
	json.Unmarshal(data, &chunks)

	words := strings.Fields(strings.ToLower(query))
	var unique []string
	seen := map[string]bool{}
	for _, w := range words {
		w = strings.TrimSpace(w)
		if len(w) > 0 && !seen[w] {
			seen[w] = true
			unique = append(unique, w)
		}
	}

	type scored struct {
		c Chunk
		s int
	}
	var results []scored
	for _, c := range chunks {
		text := strings.ToLower(c.Text + " " + c.Heading + " " + c.Source)
		score := 0
		for _, w := range unique {
			if strings.Contains(text, w) {
				score++
			}
		}
		if score > 0 {
			results = append(results, scored{c, score})
		}
	}
	// 排序
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].s > results[i].s {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	if topK > len(results) {
		topK = len(results)
	}

	var out []string
	for i := 0; i < topK; i++ {
		r := results[i].c
		text := r.Text
		runes := []rune(text)
		if len(runes) > 300 {
			text = string(runes[:300]) + "..."
		}
		out = append(out, fmt.Sprintf("[来源 %d] %s — %s\n%s", i+1, r.Source, r.Heading, text))
	}
	return strings.Join(out, "\n\n")
}

// ═══════════════════════════════════════════════════════════════
// Tool definitions (for LLM function calling)
// ═══════════════════════════════════════════════════════════════

func getToolDefs() []ToolDef {
	return []ToolDef{
		{
			Type: "function",
			Function: FuncSpec{
				Name:        "read_proposal",
				Description: "读取 DAO 提案的完整内容。返回提案文本。调用此工具了解提案在讨论什么。",
				Parameters: ParamObj{
					Type:       "object",
					Properties: map[string]ParamDef{},
					Required:   []string{},
				},
			},
		},
		{
			Type: "function",
			Function: FuncSpec{
				Name:        "search_handbook",
				Description: "搜索 AI × Web3 School Handbook，找到与指定关键词相关的章节内容。参数 query 是搜索关键词（空格分隔），top_k 是返回结果数量。",
				Parameters: ParamObj{
					Type: "object",
					Properties: map[string]ParamDef{
						"query": {Type: "string", Description: "搜索关键词，用空格分隔（如：投票 提案 风险）"},
						"top_k": {Type: "string", Description: "返回结果数量，默认 3"},
					},
					Required: []string{"query"},
				},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════════
// LLM API
// ═══════════════════════════════════════════════════════════════

func getAPIKey() string {
	if k := os.Getenv("DEEPSEEK_API_KEY"); k != "" {
		return k
	}
	data, err := os.ReadFile(os.ExpandEnv("$HOME/.hermes/.env"))
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "DEEPSEEK_API_KEY=") {
				return strings.Trim(strings.TrimPrefix(line, "DEEPSEEK_API_KEY="), "\"'")
			}
		}
	}
	return ""
}

func callLLM(messages []Message, tools []ToolDef) (Message, error) {
	apiKey := getAPIKey()
	if apiKey == "" {
		return Message{}, fmt.Errorf("DEEPSEEK_API_KEY 未设置")
	}

	req := ChatRequest{
		Model:       "deepseek-chat",
		Messages:    messages,
		Tools:       tools,
		Temperature: 0.1,
		MaxTokens:   2000,
	}

	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions",
		bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return Message{}, fmt.Errorf("API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
	}
	json.Unmarshal(respBody, &result)

	if len(result.Choices) == 0 {
		return Message{}, fmt.Errorf("API 无返回")
	}

	msg := result.Choices[0].Message
	return Message{
		Role:      "assistant",
		Content:   msg.Content,
		ToolCalls: msg.ToolCalls,
	}, nil
}

// ═══════════════════════════════════════════════════════════════
// Execute tool calls
// ═══════════════════════════════════════════════════════════════

func executeTool(tc ToolCall) string {
	switch tc.Function.Name {
	case "read_proposal":
		return readProposal()
	case "search_handbook":
		var args struct {
			Query string `json:"query"`
			TopK string `json:"top_k"`
		}
		json.Unmarshal([]byte(tc.Function.Arguments), &args)
		topK := 3
		fmt.Sscanf(args.TopK, "%d", &topK)
		return searchHandbook(args.Query, topK)
	default:
		return "未知工具: " + tc.Function.Name
	}
}

// ═══════════════════════════════════════════════════════════════
// Agent Loop
// ═══════════════════════════════════════════════════════════════

func main() {
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║   DAO 提案研究 Agent                    ║")
	fmt.Println("║   智能体（Agent）最小实践 — Go 版本     ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	systemPrompt := `你是一个 DAO 提案研究 Agent。你的任务是研究一份提案，输出投票前检查清单。

你可以使用以下工具：
1. read_proposal — 读取提案完整内容
2. search_handbook — 搜索 Handbook，查找与提案相关的概念、风险、注意事项

流程：
1. 先调用 read_proposal 了解提案内容
2. 根据提案内容，调用 search_handbook 搜索相关章节（至少搜索 2-3 次不同关键词）
3. 综合所有信息，输出结构化的投票前检查清单

输出格式要求：
{
  "proposal_summary": "一句话总结提案",
  "key_points": ["要点1", "要点2"],
  "relevant_knowledge": ["从 Handbook 中找到的相关知识"],
  "risks": ["识别出的风险"],
  "checklist": [
    {"item": "检查项", "status": "待确认|已确认|不适用", "note": "说明"}
  ],
  "recommendation": "建议投票方向",
  "uncertainties": ["仍然不确定的事项", "需要人工判断的事项"]
}`

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: "请研究这份 DAO 提案，输出投票前检查清单。"},
	}

	tools := getToolDefs()
	maxRounds := 8

	for round := 1; round <= maxRounds; round++ {
		fmt.Printf("\n🔄 [第 %d 轮] 调用 LLM...\n", round)

		reply, err := callLLM(messages, tools)
		if err != nil {
			fmt.Printf("❌ 错误: %v\n", err)
			break
		}

		// LLM 没有调用工具 → 最终输出
		if len(reply.ToolCalls) == 0 {
			fmt.Println("✅ Agent 完成，输出检查清单:")
			fmt.Println(strings.Repeat("-", 50))
			fmt.Println(reply.Content)
			break
		}

		// LLM 调用了工具 → 执行
		messages = append(messages, reply)
		for _, tc := range reply.ToolCalls {
			fmt.Printf("   🛠 调用工具: %s(%s)\n", tc.Function.Name, tc.Function.Arguments)
			result := executeTool(tc)
			fmt.Printf("   ✅ 返回 %d 字符\n", len([]rune(result)))

			messages = append(messages, Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Name:       tc.Function.Name,
				Content:    result,
			})
		}
	}
}
