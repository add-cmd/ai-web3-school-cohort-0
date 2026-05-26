# 每日学习记录 — 2026-05-25

## 今日课程
**Day 6：Week 2 方向选择 — Wallet / Permission / Safe Execution**

### 今日路径

#### 核心任务（方向选择与方案设计）
- [x] ✅ 6 方向全量分析：AI 作用 × Web3 机制 × 交叉性论证
- [x] ✅ 主方向确认：Module D — Wallet / Permission / Safe Execution
- [x] ✅ 落点确认：AgentPact MVP（从 5/20 advisor 演进为 executor with bounded autonomy）
- [x] ✅ 模块 A 交付物：问题地图（`module-a-problem-map.md`）
- [x] ✅ 模块 D 交付物：方向深挖设计文档（`module-d-deep-dive.md`）

#### 挑战路径（有余力做）
- [x] ✅ 参考资料清单 ≥5 条 + 每条价值判断（`module-e-references.md`）— **10 条**
- [x] ✅ 方向 backlog 详化（`module-g-backlog-detailed.md`）— **5 个方向**

---

## 学习笔记

### 🔵 核心认知：如何选择 Week 2 方向

Week 2 的要求不是"选一个方向随便看看"，而是：**论证为什么这个方向是真正的 AI × Web3 交叉问题**，即"不能少 AI，也不能少 Web3"。

选择 Module D（Wallet / Permission / Safe Execution）的决策链：

| 思考步骤 | 结论 |
|:---------|:------|
| 6 方向逐一做交叉性分析 | 每个方向标清 AI 承担的具体能力和 Web3 提供的具体机制 |
| Module D 论证 | LLM 把权限系统从"工程师专属"降到"普通用户可用"，Web3 把"代理执行"从"信平台"升到"信合约" |
| 落选方向反证 | 自然语言钱包（纯 UI 升级）不是 D、AI Trading Bot 不加 Pact 不是 D、硬编码白名单（AI 角色缺失）不是 D |

#### 决策树

```
用户输入意图 → LLM 生成 Pact 草案 → 用户审阅签名 → Pact 生效
                              ↓
              Agent 在 Pact 边界内自主规划执行
                              ↓
              每笔 tx 发起前 → Policy Engine 判定
                              ├─ ALLOW → Session Key 签名上链
                              ├─ DENY → 直接拒绝 + 记录
                              └─ ESCALATE → 等用户决策（超时=DENY）
                              ↓
                   Audit Log 记录完整决策 snapshot
```

#### 关键文件一览

| 文件 | 核心内容 | 行数 |
|:----|:---------|:----:|
| `module-a-problem-map.md` | 6 方向分析表 + 交叉性论证 + 落选理由 + Week 2 剩余任务 | 159 |
| `module-d-deep-dive.md` | 参与方/流程/Pact JSON schema/Policy Engine 伪代码/反例/风险表(R1-R10)/验证计划 | 523 |
| `module-e-references.md` | 10 条参考资料 + 每条价值判断 + Week 2 动作建议 | 184 |
| `module-g-backlog-detailed.md` | 5 个方向(B/C/F/G/E)的 backlog + 触发条件 + Week 2→4 路线图 | 211 |

---

### 🟢 Web3 线：AgentPact MVP 设计要点

#### Pact JSON Schema 核心设计

```
Pact = 用户授予 Agent 的一组可验证、可撤销、有边界的执行权限

├─ principal:   谁授权（用户 EOA）
├─ agent:       谁执行（session key）
├─ scope:       在哪里执行（合约白名单 × 函数白名单 × token × 对手方）
├─ budget:      花多少（per_call_max + total_max + runtime spent）
├─ limits:      频率（max_calls + rate_limit）
├─ time:        有效期（valid_from → valid_to）
├─ escalation:  什么情况等用户（金额阈值/新合约/预算告警/simulation 异常）
├─ revocation:  如何收回（用户签名/到期/紧急冻结地址）
├─ audit:       怎么记录（本地 JSONL + 可选 EAS attestation）
└─ hard_constraints: 绝对禁止（setApprovalForAll / transferOwnership / selfdestruct）
```

#### Policy Engine 四层决策

| 层 | 检查项 | 结果 |
|:--|:-------|:-----|
| Layer 1: Hard Constraints | 绝对禁止函数、gas 上限、黑名单地址 | DENY |
| Layer 2: Pact 边界 | 有效期、合约白名单、函数白名单、额度、频率 | DENY |
| Layer 3: Simulation | eth_call 模拟、检查是否 revert | DENY（如果 revert） |
| Layer 4: Escalation | 金额阈值、新合约、预算警戒、simulation 警告 | ESCALATE（等用户） |
| 全部通过 | — | ALLOW（自动签名） |

#### 反例（什么不是 AgentPact）

| 反例 | 为什么不是 D | 其实是哪个方向 |
|:----|:------------|:--------------|
| 自然语言钱包（NL Wallet） | 没有 policy / 每笔仍用户签 | Module C（Capability） |
| AI 提案 + Multi-sig 通过 | 没有自动执行 | Module G（Governance） |
| AI Trading Bot | 没有 Pact 边界 | Module E（DeFi） |
| 硬编码白名单 Bot | 没有"意图→边界→确认"环节 | 传统 bot |

#### 风险表摘要（TOP 5）

| # | 风险 | 缓解核心 |
|:-:|:----|:---------|
| R1 | Prompt Injection → 超界 Pact | 草案必须 user 签名才生效 + schema 校验 |
| R2 | Tool Abuse（在 Pact 内做坏事） | Policy Engine 强制校验 + 硬上限 |
| R3 | Session Key 私钥泄漏 | Pact TTL 短 + 总预算上限 + emergency freeze |
| R4 | Pact 写错（用户授权过宽） | UI 默认保守值 + dry-run 演练 |
| R6 | LLM 幻觉误判意图 | 签名前强制审阅 + diff 视图 |

---

## 打卡草稿

```
📖 Day 6 | AI × Web3 School
主题：Week 2 方向选择 — Wallet / Permission / Safe Execution

【决策】6 方向全量交叉分析后，选择了 Wallet / Permission / Safe Execution（Module D）
【落点】AgentPact MVP — 从 5/20 的 advisor 演进为 bounded executor
【设计】Pact JSON schema（9 大字段）+ Policy Engine 四层决策 + 10 类风险评估
【产出】4 份设计文档（问题地图 / 深挖设计 / 参考资料 / Backlog 详化），累计约 68KB

#AIxWeb3School #WalletPermission #AgentPact #Day6
```

## 提交记录

- [x] ✅ 方向选择 + 深挖设计文档（module-a, module-d）
- [x] ✅ 参考资料清单（10 条，含价值判断）
- [x] ✅ 方向 backlog 详化（5 个方向 + 触发条件 + 路线图）
- [ ] ❌ Week 1 打卡提交至 WCB（笔记已生成，待手动提交）

## 明日计划（5/26）

- [ ] 确认 daily note 内容 + commit 到 GitHub
- [ ] Week 1 打卡提交至 WCB 平台
- [ ] 开始 Week 3 准备：如果 Week 2 方向选择算完成，可以开始规划 MVP 实现
