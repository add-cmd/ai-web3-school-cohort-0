# Day 5 — Agent 最小实践：DAO 提案研究 Agent

> **日期：** 2026-05-22
> **章节：** Agent（智能体）
> **核心：** 用 Function Calling + Tool Use 实现自主研究 Agent
> **技术栈：** Go + DeepSeek API
> **语言：** Go（标准库，无外部依赖）

---

## 项目：DAO 提案研究 Agent

**让 LLM 自己决定用哪个工具、查什么资料。** 输入一份 DAO 治理提案，Agent 自主调用工具进行多轮研究，输出投票前检查清单。

### Agent 架构

```
用户输入 "研究这份 DAO 提案"
     ↓
┌─ Agent Loop (最多 8 轮) ──────────────────────────────────┐
│                                                           │
│  LLM (deepseek-chat)          ←→  Tool Set               │
│    ├─ 阅读提案内容             ├─ read_proposal()         │
│    ├─ 搜索 Handbook 知识        ├─ search_handbook()       │
│    ├─ 分析风险                  └─ (关键词匹配检索)         │
│    └─ 生成检查清单                                         │
│                                                           │
└───────────────────────────────────────────────────────────┘
     ↓
 输出: 投票前检查清单 (结构化 JSON)
```

### 文件结构

| 文件 | 说明 |
|:---|:---|
| `main.go` | Agent 主程序（347 行 Go），含 Tool 定义、LLM API 调用、Agent Loop |
| `go.mod` | Go Module 配置（Go 1.26，零外部依赖） |
| `chunks.json` | Handbook 知识库 chunk 数据（供 `search_handbook` 工具检索） |
| `proposal_governance_ai.txt` | 待研究的 DAO 治理提案全文（~7KB） |

### Agent 实现要点

- **Function Calling：** Go 原生实现，模拟 DeepSeek API 的 tool 调用协议
- **多轮思考：** Agent 自主决定调用工具的次序和次数，每次工具结果回送给 LLM 继续推理
- **终止条件：** LLM 不再调用工具时，输出最终检查清单
- **工具 1 `read_proposal`：** 读取提案文件，零参数
- **工具 2 `search_handbook`：** 关键词检索 Handbook 知识库，返回 top_k 个相关段落

### 示例输出结构

```json
{
  "proposal_summary": "一句话总结提案",
  "key_points": ["要点1", "要点2"],
  "relevant_knowledge": ["从 Handbook 中找到的相关知识"],
  "risks": ["识别出的风险"],
  "checklist": [
    {"item": "检查项", "status": "待确认|已确认|不适用", "note": "说明"}
  ],
  "recommendation": "建议投票方向",
  "uncertainties": ["仍然不确定的事项"]
}
```

### 学习笔记

**Agent vs RAG：**
- RAG（Day 4）：一次检索 + 一次回答，被动工具
- Agent（Day 5）：自主多轮调用工具，LLM 自己决定"下一步做什么"

**Tool Use 设计模式：**
- Tool 是 LLM 感知外部世界的接口
- LLM 输出 function_call → 程序执行 → 结果注入下一轮 → 循环直到 LLM 完成
- 关键设计：每轮消息序列保留完整上下文（system + user + assistant_toolcalls + tool_results + ...）
