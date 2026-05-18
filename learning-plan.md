# 学习计划

> 基于 Handbook 结构 + 个人画像定制。
> 每日可投入：3小时+ | 方向：开发

## 学习路径概览

```
Phase 1: AI 基础 (预计 2-3 周)
     ↓
Phase 2: AI × Web3 Bridge (预计 2-3 周)
     ↓
Phase 3: 前沿探索 × Hackathon (持续)
```

---

## Phase 1：AI 基础

**目标**：建立 LLM、Prompt、Context、RAG、Agent 的共同语言

### Week 1 — 理解模型与提示
| 天 | 学习内容 | Handbook 章节 | 实践 |
|---|---|---|---|
| 1 | 大模型能做什么/不能做什么 | [LLM](https://aiweb3.school/zh/handbook/ai/llm/) | 用 API 调用一个模型 |
| 2 | 写好 Prompt 的三要素 | [Prompt](https://aiweb3.school/zh/handbook/ai/prompt/) | 练习写结构化 Prompt |
| 3 | 模型上下文窗口管理 | [Context](https://aiweb3.school/zh/handbook/ai/context/) | 实验 Token 管理 |
| 4-5 | 外部知识检索接入 | [RAG](https://aiweb3.school/zh/handbook/ai/rag/) | 搭建最小 RAG 系统 |
| 6-7 | 复习 & 打卡提交 | — | 写 Week 1 总结 |

### Week 2 — Agent 与工具调用
| 天 | 学习内容 | Handbook 章节 | 实践 |
|---|---|---|---|
| 1-2 | Agent 工作流基础 | [Agent](https://aiweb3.school/zh/handbook/ai/agent/) | 实现一个 Tool Calling Demo |
| 3 | 框架入门对比 | [Frameworks](https://aiweb3.school/zh/handbook/ai/frameworks/) | 选一个框架跑通 |
| 4 | MCP 协议理解 | [MCP](https://aiweb3.school/zh/handbook/ai/mcp/) | 连接本地 MCP Server |
| 5 | 评估与测试 | [Evaluation](https://aiweb3.school/zh/handbook/ai/evaluation/) | 给你的 Agent 写评估 |
| 6-7 | 复习 & 打卡提交 | — | 写 Week 2 总结 |

### Week 3 — AI 进阶与工具体验
| 天 | 学习内容 | Handbook 章节 | 实践 |
|---|---|---|---|
| 1 | 氛围编程 | [Vibe Coding](https://aiweb3.school/zh/handbook/ai/vibe-coding/) | 用 AI 辅助写一个小工具 |
| 2 | 微调入门 | [Fine-tuning](https://aiweb3.school/zh/handbook/ai/fine-tuning/) | 了解微调流程 |
| 3 | 推理服务 | [Inference](https://aiweb3.school/zh/handbook/ai/inference/) | 本地/云端跑一个模型 |
| 4-7 | Phase 1 复盘、补漏、提交 | — | 手写一份 AI 基础总结笔记 |

---

## Phase 2：AI × Web3 Bridge

**前提**：已有 Web3 基础，重点看 AI 侧的接入方式

### Week 4 — 链上上下文与工具调用
| 天 | 学习内容 | Handbook 章节 | 实践 |
|---|---|---|---|
| 1-2 | 链上状态进入 Agent 上下文 | [Chain-aware Context](https://aiweb3.school/zh/handbook/bridge/chain-aware-context/) | Agent 读取链上数据 |
| 3-4 | Web3 工具调用 | [Web3 Tool Use](https://aiweb3.school/zh/handbook/bridge/web3-tool-use/) | Agent 调用 RPC/合约 |
| 5-7 | Agent 工作流设计 | [Agent Workflow](https://aiweb3.school/zh/handbook/bridge/agent-workflow/) | 画一个 human-in-the-loop 流程 |

### Week 5 — 钱包、支付与身份
| 天 | 学习内容 | Handbook 章节 | 实践 |
|---|---|---|---|
| 1-2 | Agent 钱包权限 | [Agent Wallet](https://aiweb3.school/zh/handbook/bridge/agent-wallet/) | 配置 Session Key/Policy |
| 3 | 机器支付 | [Machine Payment](https://aiweb3.school/zh/handbook/bridge/machine-payment/) | 模拟小额支付 |
| 4-5 | Agent 身份与信誉 | [Agent Identity](https://aiweb3.school/zh/handbook/bridge/agent-identity/) + [Trust](https://aiweb3.school/zh/handbook/bridge/agent-trust-and-reputation/) | 设计身份方案 |
| 6-7 | 可验证 AI | [Verifiable AI](https://aiweb3.school/zh/handbook/bridge/verifiable-ai/) | 理解验证流程 |

### Week 6 — 安全、隐私、治理与复盘
| 天 | 学习内容 | Handbook 章节 | 实践 |
|---|---|---|---|
| 1-2 | AI 安全 | [AI Security](https://aiweb3.school/zh/handbook/bridge/ai-security/) | 写 Prompt Injection 测试 |
| 3 | AI 隐私 | [AI Privacy](https://aiweb3.school/zh/handbook/bridge/ai-privacy/) | 了解隐私边界 |
| 4 | 治理 AI | [Governance AI](https://aiweb3.school/zh/handbook/bridge/governance-ai/) | 设计治理辅助工具 |
| 5-7 | Phase 2 复盘 & 提交 | — | Bridge 总结笔记 |

---

## Phase 3：前沿探索 × 项目

**前提**：Phase 1-2 完成，选择一条赛道做原型

### 可选方向

| 方向 | Handbook 章节 | 适合谁 |
|---|---|---|
| 智能体商业 | [Agentic Commerce](https://aiweb3.school/zh/handbook/tracks/agentic-commerce/) | 想做支付/服务场景 |
| 钱包与权限 | [Wallet / Permission](https://aiweb3.school/zh/handbook/tracks/wallet-permission/) | 想做授权/安全方案 |
| AI 安全 | [AI Security](https://aiweb3.school/zh/handbook/tracks/ai-security/) | 想做安全/审计工具 |
| 治理 | [Governance](https://aiweb3.school/zh/handbook/tracks/governance/) | 想做 DAO/公共工具 |
| 开发工具 | [Dev Tooling](https://aiweb3.school/zh/handbook/tracks/dev-tooling/) | 想做开发工作流工具 |
| 开放赛道 | [Open Track](https://aiweb3.school/zh/handbook/tracks/open-track/) | 其他交叉方向 |

---

## 每周节奏建议

| 时间 | 内容 |
|---|---|
| 周一至周五 | 每天 3h：理论学习 + 动手实践 |
| 周六 | 项目/实验时间（Hackathon 或 experiments/） |
| 周日 | 复盘、写打卡、整理 Handbook feedback |

## 打卡提醒

- [ ] 早上：查看 WCB Learning 页面 → 规划今日任务
- [ ] 晚上：写 daily note → 生成打卡草稿 → 提交
