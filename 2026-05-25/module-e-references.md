# 参考资料清单 — AgentPact MVP

> 方向：Wallet / Permission / Safe Execution
> 日期：2026-05-25
> 用途：Week 2 交付物 #5 — ≥5 条参考资料，含每条的价值判断

---

## 1. ERC-4337：Account Abstraction Standard

| 维度 | 内容 |
|:----|:-----|
| **链接** | https://eips.ethereum.org/EIPS/eip-4337 |
| **类型** | 以太坊核心标准（EIP） |
| **价值判断** | ⭐⭐⭐⭐⭐ **必读/核心参考** |

AgentPact MVP v0 先用测试 EOA 做 session key 简化版，但 Week 3 升级到真正的 smart account 时，4337 是必经之路。它定义了 UserOperation → EntryPoint → Bundler → Paymaster 的完整流程，是 Agent 的签名/权限/执行层的基础设施。

核心关注点：session key 注册方式、UserOperation 结构、paymaster 代付 gas 机制。

> 互补文档：https://docs.erc4337.io/（开发者文档，比 EIP 原文实操）

---

## 2. Safe Docs：多签 & Guard 模块

| 维度 | 内容 |
|:----|:-----|
| **链接** | https://docs.safe.global/ |
| **类型** | 产品文档 |
| **价值判断** | ⭐⭐⭐⭐⭐ **必读/核心参考** |

Safe 是生产环境下最多使用的智能钱包框架。它的 **Guard** 机制——一笔交易执行前通过 guard 合约强制检查——是 AgentPact Policy Engine 的设计原型。

核心关注点：Guard 的 before/after hook 模式、module 架构、如何用 guard 实现"白名单 + 额度 + 人工确认"三层校验。

> 关联：AgentPact 的 Policy Engine 是 off-chain guard，后续可迁移为 Safe Guard 合约。

---

## 3. OWASP Top 10 for LLM Applications

| 维度 | 内容 |
|:----|:-----|
| **链接** | https://owasp.org/www-project-top-10-for-llm-applications/ |
| **类型** | 安全分类标准 |
| **价值判断** | ⭐⭐⭐⭐ **高价值/威胁建模参考** |

AgentPact 的 10 类风险（R1-R10）中的 Prompt Injection（R1）、Tool Abuse（R2）、LLM 幻觉导致错误授权（R6），都对应 OWASP LLM Top 10 中的特定条目。

核心关注点：LLM01 (Prompt Injection)、LLM06 (Sensitive Information Disclosure)、LLM08 (Excessive Agency)。用于交叉验证自己的风险表是否覆盖全面。

> 注意：OWASP 偏向通用 LLM 应用安全，Web3 特有的（Session Key 泄漏、Replay Attack、合约层强制）需要自己补充。

---

## 4. Cobo CAW (Composable Account Wallet)

| 维度 | 内容 |
|:----|:-----|
| **链接** | https://www.cobo.com/products/cobowallet |
| **类型** | 产品/API 文档 |
| **价值判断** | ⭐⭐⭐⭐ **高价值/竞品参考** |

Cobo CAW 是一个面向 AI Agent 的生产级钱包产品。它的 **Session Key + Pact API** 模式和 AgentPact MVP 的核心假设高度重合——用户授 Pact → agent 在边界内自主执行。Week 3 集成路线中列为优先级 #2。

核心关注点：Cobo 的 Pact JSON schema 设计、权限维度（合约/函数/额度/时间）、escalation 模式、revocation 的紧急冻结机制。用于验证 AgentPact 的设计是否过于简单或遗漏了实际产品中的关键要素。

> 注意：CAW 是托管的（Cobo 作为 custodian），AgentPact 的核心理念是去信任化的 self-custody。两个方案的信任模型不同，但 UX 参考价值大。

---

## 5. ERC-7702（EIP-7702）：Set EOA Account Code

| 维度 | 内容 |
|:----|:-----|
| **链接** | https://eips.ethereum.org/EIPS/eip-7702 |
| **类型** | 以太坊标准（即将上线） |
| **价值判断** | ⭐⭐⭐⭐ **高价值/前瞻参考** |

ERC-7702 让 EOA 可以临时获得合约代码能力，无需部署完整 smart account。这对 AgentPact 的意义在于：用户不需要先部署一个 4337 smart account，而是直接在 EOA 上临时授权一个 session key 模式。

可能改变 MVP 的技术路径：如果 v0 用测试 EOA 做 session key，ERC-7702 可以 bridge 到"真正的 EOA + 合约级权限控制"的中间态。

> 状态：已 Draft，预计 2025 年 Pectra 升级后可用。Week 2-3 阶段关注进展即可，不必依赖。

---

## 6. Ethereum Attestation Service (EAS)

| 维度 | 内容 |
|:----|:-----|
| **链接** | https://attest.org/ |
| **类型** | 协议文档 |
| **价值判断** | ⭐⭐⭐ **参考/可选（Week 3 扩展用）** |

EAS 提供链上可验证的 attestation 数据结构。AgentPact 的 audit log 目前设计为本地 JSONL + 可选 EAS attestation（第 7 节中 on_chain_attestation.enabled = false）。

Week 3 扩展优先级 #4 涉及将关键决策（Pact 创建、高风险 ALLOW、DENY 原因）通过 EAS 上链，让第三方可验证"这个 agent 确实在 Pact 边界内执行了这些交易"。

核心关注点：Schema 注册、on-chain/off-chain attestation 模式、attestation 的 revocation。

> 最小实践：Week 2 先理解 EAS 工作原理，不需要实际集成。

---

## 7. ERC-7579：Modular Smart Account

| 维度 | 内容 |
|:----|:-----|
| **链接** | https://eips.ethereum.org/EIPS/eip-7579 |
| **类型** | 以太坊标准（Draft） |
| **价值判断** | ⭐⭐⭐ **参考/可选（Week 3 扩展用）** |

ERC-7579 定义了模块化智能账户标准，让 smart account 可以按需安装/卸载功能模块（比如 session key 模块、guard 模块、recovery 模块）。AgentPact 的 policy engine 如果以后要链上化，挂成 7579 的一个模块是不错的方向。

核心关注点：模块接口规范、模块之间的权限隔离、模块的升级/撤销方式。

> 当前优先级低——先跑通 MVP v0，再考虑模块化架构。

---

## 8. OpenZeppelin Defender

| 维度 | 内容 |
|:----|:-----|
| **链接** | https://defender.openzeppelin.com/ |
| **类型** | 运维工具 |
| **价值判断** | ⭐⭐⭐ **参考/可选（生产化参考）** |

Defender 提供了交易模拟（Simulate）、自动执行（Autotask）、监控（Sentinel）、Relayer 签名四大能力，和 AgentPact 的 policy + agent loop + audit log 架构有对应关系。

核心关注点：Relayer 的权限控制方式、Sentinel 的异常监控规则配置。可用于验证自己的 policy engine 设计是否遗漏了运维侧的安全考虑。

> Week 2 阶段不需要用 Defender，但理解它的架构可以对齐行业参考。

---

## 9. LangChain Security Policy

| 维度 | 内容 |
|:----|:-----|
| **链接** | https://python.langchain.com/docs/security/ |
| **类型** | 安全最佳实践文档 |
| **价值判断** | ⭐⭐⭐ **参考/可选** |

LangChain 的安全文档覆盖了 tool calling 的安全边界、外部输入的处理策略、如何避免过度代理（excessive agency）。AgentPact 的 Tool Permission Isolation（R2 缓解）可以参考这里的模式。

核心关注点：Tool 输入验证、tool 返回值可信度分级、human-in-the-loop 的触发模式。

> 注意：LangChain 是 Python 生态，AgentPact 用 Go 实现，所以参考设计思路而非代码。

---

## 10. OpenAI Function Calling / Tool Calling Docs

| 维度 | 内容 |
|:----|:-----|
| **链接** | https://platform.openai.com/docs/guides/function-calling |
| **类型** | API 文档 |
| **价值判断** | ⭐⭐⭐ **参考/基础理解用** |

理解 LLM 如何发出 tool call、结构化参数如何约束、模型如何选择工具。AgentPact 中 LLM 承担 Pact Drafter 和 Agent 规划层两个角色，都依赖正确的 tool calling 行为。

核心关注点：Parallel function calling、structured outputs（JSON mode）、tool_choice 控制。用于指导 Pact JSON schema 的结构化输出设计。

> 注意：AgentPact 用 DeepSeek API，但 tool calling 协议与 OpenAI 兼容。

---

## 清单总览

| # | 资料 | 价值 | 时机 | Week 2 动作 |
|:-:|:----|:----:|:----:|:-----------|
| 1 | ERC-4337 | ⭐⭐⭐⭐⭐ | Week 3 集成 | 通读架构图、理解 UserOperation 流程 |
| 2 | Safe Docs (Guard) | ⭐⭐⭐⭐⭐ | 设计参考 | 读 Guard 文档，对齐 Policy Engine 设计 |
| 3 | OWASP LLM Top 10 | ⭐⭐⭐⭐ | 安全验证 | 用 OWASP 框架交叉检查 R1-R10 |
| 4 | Cobo CAW | ⭐⭐⭐⭐ | 竞品分析 | 读 Pact API 文档，对比自己的 schema |
| 5 | ERC-7702 | ⭐⭐⭐⭐ | 前瞻参考 | 了解基本概念，记入 architectural decision log |
| 6 | EAS | ⭐⭐⭐ | Week 3 扩展 | 理解 schema + attestation 模式 |
| 7 | ERC-7579 | ⭐⭐⭐ | Week 3 扩展 | 了解模块化架构概念 |
| 8 | OpenZeppelin Defender | ⭐⭐⭐ | 生产化参考 | 浏览架构，不需要深入 |
| 9 | LangChain Security | ⭐⭐⭐ | 设计参考 | 读 excessive agency + tool validation 部分 |
| 10 | OpenAI Tool Calling | ⭐⭐⭐ | 基础参考 | 确认 structured outputs 用法 |
