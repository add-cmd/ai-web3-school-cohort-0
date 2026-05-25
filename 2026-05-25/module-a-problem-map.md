# Week 2 模块 A 交付物 — 问题地图与方向选择

> 日期：2026-05-25
> 主方向（最终）：**Module D — Wallet / Permission / Safe Execution**
> 具体落点：**AgentPact MVP**（5/20 advisor 演进为 executor with bounded autonomy）

---

## 一、AI × Web3 问题地图

> 选 6 个方向，每个标清 **AI 承担的具体能力**（来自：理解 / 生成 / 规划 / 工具调用 / 自动化 / 监控 / 总结 / 协作）与 **Web3 提供的具体机制**（来自：支付 / 身份 / 权限 / 开放状态 / 可验证记录 / 结算 / 抗审查 / 协作机制）。

### 总览表

| # | 方向 | AI 作用 | Web3 机制 | 典型场景 |
|---|---|---|---|---|
| ① | **Payment / Commerce** | 理解服务规格、规划购买、自动化执行、监控交付 | 支付、结算、可验证记录（receipt/escrow） | Agent 自动购买 API / 数据 / 算力 |
| ② | **Identity / Capability** | 理解能力声明、协作（agent-to-agent） | 身份（DID/ENS）、可验证记录（EAS）、开放状态（registry） | Agent profile + capability manifest |
| ③ | **Wallet / Permission** ⭐主方向 | 理解用户意图、生成 policy 草案、规划执行、监控异常 | 权限（AA/session key）、可验证记录（audit）、结算 | 用户授 Pact → agent 在边界内自主执行 |
| ④ | **Privacy / Security** | 监控攻击、生成 threat scenarios、总结风险 | 抗审查、可验证执行（TEE）、开放状态（不依赖单一供应商） | Prompt injection 防御、TEE 推理 |
| ⑤ | **Dev Tooling** | 理解代码/文档、生成测试/脚本、自动化 workflow | 开放状态（公链文档/合约）、可验证记录（链上部署） | docs-to-agent、交易解释、合约助手 |
| ⑥ | **Governance** | 理解提案、总结讨论、生成行动项、协作分发 | 协作机制（DAO/quorum）、可验证记录（投票/贡献）、抗审查 | 提案 copilot、贡献追踪、预算 checklist |

### 每个方向的具体拆解

#### ① Payment / Commerce / Settlement
- **AI 做什么**：从服务描述里提取价格/SLA/交付标准；判断当前预算下是否购买；监控交付结果是否符合验收条件
- **Web3 提供什么**：USDC 链上结算、x402 paywall 协议、escrow 合约托管、链上 receipt 不可篡改
- **不能少 AI**：每次 agent 之间报价/验收都靠人写规则不可扩展
- **不能少 Web3**：Web2 支付没有"机器对机器 + 无需信任 + 可验证收据"的组合

#### ② Identity / Reputation / Capability
- **AI 做什么**：把"会做什么"翻译成结构化 capability manifest；agent 间用自然语言协商任务
- **Web3 提供什么**：ENS 命名、EAS 链上证明、ERC-8004 trust registry、可验证历史任务记录
- **不能少 AI**：capability 描述太复杂，纯 schema 不够，需要语义理解
- **不能少 Web3**：reputation 必须公开、不可篡改，否则就是中心化平台 review

#### ③ Wallet / Permission / Safe Execution（主方向，详见第二节）

#### ④ Privacy / Security / Sovereignty
- **AI 做什么**：模拟攻击者生成 prompt injection 样本；监控异常工具调用模式；用 LLM 解释审计日志
- **Web3 提供什么**：TEE 可验证执行、开源模型本地部署、不依赖单一供应商的 sovereignty
- **不能少 AI**：攻击模式持续进化，规则引擎追不上
- **不能少 Web3**：AI 安全的极端是"模型在哪里跑、谁能看到提示词"，必须有可验证执行环境

#### ⑤ Dev Tooling / Agent Workflow
- **AI 做什么**：读合约/文档生成解释、写测试、生成部署脚本、解释交易
- **Web3 提供什么**：公链状态本来就开放、合约 ABI 公开、git 历史可验证
- **不能少 AI**：开发者文档/合约阅读门槛高，靠 AI 摘要才能规模化
- **不能少 Web3**：但 Web3 部分对 AI 不是"必须"——Web2 dev tooling 也成立。**这是个弱交叉方向**

#### ⑥ Governance / Coordination / Public Goods
- **AI 做什么**：总结提案、转会议为行动项、追踪贡献、生成预算 checklist
- **Web3 提供什么**：Snapshot 投票、链上贡献证明、DAO treasury、Gitcoin 二次方资助
- **不能少 AI**：DAO 信息密度过高（千条讨论无人读完）
- **不能少 Web3**：治理的合法性需要公开+不可篡改记录，纯 AI 总结无法承担

---

## 二、两个方向的"为什么不是纯 AI / 纯 Web3 问题"

> 按官方要求选 2 个方向论证。我选 **Module D（主方向候选）** 和 **Module G（备选，已有 5/22 prototype）**。

### Module D — Wallet / Permission / Safe Execution

#### 为什么不是纯 AI 问题？

LLM 自己**不能**：
- **不能签名上链**：模型是无状态推理器，私钥的物理位置在钱包/HSM/MPC 节点
- **不能强制预算**：模型可以"承诺"不超额，但承诺没有强制力。链上 policy 合约才能在执行层拒绝
- **不能留证**：模型输出不可审计、不可追溯，"agent 说它没乱花钱"无法验证
- **不能解决信任**：用户为什么相信一个跑在 OpenAI/Anthropic 后端的 agent？必须有不依赖供应商的边界

→ 没有 Web3 提供的**权限合约 + 可验证记录 + 结算层**，所谓"agent 自动执行"就是"信中心化平台老板"。

#### 为什么不是纯 Web3 问题？

Multi-sig、ERC-4337、Safe Guards 已经存在 3-5 年，但：
- **门槛太高**：写 policy 合约需要 Solidity，普通用户表达不了"我只允许它在 Uniswap V3 上 swap USDC，每天不超过 $500"
- **静态规则**：传统 policy 是 if-else，agent 经济需要根据上下文动态决策（市场状况、协议风险、对手方信誉）
- **没有意图层**：用户说的是"帮我管这周的支付"，不是"approve 0xabc...spender, amount=10**18, deadline=..."
- **没有解释力**：policy 拒绝了某笔 tx 时，纯链上系统只能说"reverted"，无法解释为什么

→ 没有 **LLM 提供的意图理解 + 自然语言 Pact 草案 + 动态判断 + 可解释拒绝**，权限系统就停在工程师专属的小众市场。

**结论**：**Module D 是真正的交叉问题 —— LLM 把权限系统从"工程师专属"降到"普通用户可用"，Web3 把"代理执行"从"信平台"升到"信合约"。** 两个能力同时不可替代。

---

### Module G — Governance / Coordination

#### 为什么不是纯 AI 问题？

LLM 可以总结提案、生成行动项，但：
- **没有合法性**：AI 输出的"社区共识"没人认。治理需要 Snapshot 投票、DAO quorum、链上执行才形成约束
- **没有不可篡改**：会议纪要 AI 写完发 Discord，可以被改、被删。链上 attestation 才能留证
- **没有强制预算**：AI 建议"批 $50k 给这个团队"，需要 DAO treasury + multisig 才能执行
- **没有抗审查**：纯 AI workflow 在中心化平台，治理性敏感内容容易被删

→ 没有 Web3 的**治理机制 + 可验证记录 + 链上 treasury**，AI 总结只是 Notion 升级版。

#### 为什么不是纯 Web3 问题？

DAO 治理工具（Snapshot/Tally/Governor）已经存在多年，但：
- **信息瓶颈**：一个活跃 DAO 每周几百条讨论 + 几十个提案，没人读得完，导致少数人决策
- **行动项转化**：会议→执行的链条 90% 靠人工，AI 是唯一能规模化的解法
- **贡献评估**：传统 DAO 用打分卡片做贡献评估，主观且不可扩展；LLM 可以基于代码/讨论/输出做客观摘要
- **新人门槛**：DAO 内部黑话太多，新人加入需要 LLM 提供"提案历史/关键决策"摘要才能跟上

→ 没有 **LLM 提供的总结/翻译/摘要能力**，DAO 规模一大就治理瘫痪。

**结论**：Module G 也是真交叉，但当前 DAO 生态活跃度不及 2022 峰值，市场拉力比 Module D 弱。

---

## 三、Week 2 主方向选择

### 主方向：**Module D — Wallet / Permission / Safe Execution**

### 具体落点：**AgentPact MVP**

把 5/20 的 advisor（事前风险评估）演进为 executor with bounded autonomy（用户授 Pact，agent 在边界内自主执行）。

### 选择理由（4 条）

1. **真交叉**：LLM 把权限设计从工程师降到用户；Web3 把代理执行从"信平台"升到"信合约"。两个能力同时不可替代（见第二节论证）。
2. **承接最强**：直接复用 5/20 的 SimpleToken + TokenShop 合约、Context Engineering 引擎、Go backend、Sepolia 部署经验，新增工作量最小化。
3. **市场拉力**：ERC-4337 已落地，ERC-7702 2025 上线，Cobo / Safe / Coinbase / MetaMask 都在抢这层。Hackathon 有现成 sponsor 赛道（Cobo CAW）。
4. **可扩展性**：Pact + policy engine 是底层抽象，Week 3 可以横向扩展到 Module E（DeFi 应用）或 Module B（agent commerce 的 budget control）。

### 反例：考虑过但落选的方向

| 方向 | 落选原因 | 处理 |
|---|---|---|
| Module B（Payment） | 协议层 + 商业型，单人 Week 2 范围出不了货；先做 Pact 再做 commerce 更稳 | Backlog |
| Module C（Identity） | 基建型，变现慢，单点突破难 | Backlog |
| Module E（Agent DeFi） | 需要 DeFi 实战，无现成基础；可作为 Week 3 在 D 之上的应用层 | Week 3 候选 |
| Module F（Privacy/Security） | 偏研究/审计路径，4 周难出可演示成果 | Backlog（部分内容会自然嵌入 D） |
| Module G（Governance） | 真交叉但市场拉力弱、变现路径长；已有 5/22 prototype 可放 backlog | Backlog（保留 5/22 成果） |

### 后续模块如何围绕 D 展开

| Week 2 剩余模块 | 在 D 主线下怎么做 |
|---|---|
| 模块 B（Payment） | 跳过具体 task，但学习 budget control / Pact 在 agent commerce 中的角色（CAW 案例） |
| 模块 C（Identity） | 跳过具体 task，但理解 agent profile 如何承接 Pact（执行者身份谁来证明） |
| 模块 D（核心） | 深挖：流程图 + Pact JSON schema + Policy engine 设计 + 反例 + 风险 |
| 模块 E（Agent DeFi） | 跳过具体 task，作为 D 主线的"未来应用层"放 Week 3 backlog |
| 模块 F（Security） | 跳过具体 task，但把 prompt injection / tool abuse 写进 D 的 threat model |
| 模块 G（Governance） | 跳过具体 task，但理解"边界外动作的人工/治理确认"机制 |

---

## 下一步交付物（Week 2 剩余）

- [ ] 项目初步 proposal（#4，目标用户/场景/最小功能/赛道）
- [ ] 参考资料清单 ≥5 条 + 每条判断价值（#5）
- [ ] 方向 backlog 详化（#7）
- [ ] 5/25 daily-note
