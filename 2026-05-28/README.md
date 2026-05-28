# 每日学习记录 — 2026-05-28

## 今日课程
**Day 8：Week 3 实战 — Foundry AA 合约 Demo（Account Abstraction）**

### 今日路径

#### 核心任务
- [x] ✅ 搭建 Foundry 项目（`foundry.toml` + `lib/forge-std`）
- [x] ✅ 编写 `SimpleAccount.sol` — 最小化智能合约钱包
  - 构造函数绑定 owner 地址
  - `receive()` 接收 ETH
  - `execute()` — owner 可发起任意调用（极简授权模型）
- [x] ✅ 编写 `AccountFactory.sol` — CREATE2 确定性部署
  - `getAddress(owner, salt)` — 提前预测地址（零 Gas）
  - `createAccount(owner, salt)` — 首次交易时部署，已部署则复用
- [x] ✅ 编写 `DeployAccount.s.sol` — Foundry 部署脚本
  - 从 `.env` 读取私钥
  - 部署 `AccountFactory` + 测试用 `SimpleAccount`
- [x] ✅ 配置 `.env`（Sepolia RPC + 开发钱包私钥 + Etherscan API Key）

---

## 学习笔记

### 项目位置
`/home/unthurn/Foundry/aa-demo/`

本地已通过 `forge build` 编译通过，暂未部署上链。

### 为什么是今天做这个

从 5/25-5/26 的 AgentPact 设计文档进入合约层实现阶段。AgentPact MVP 需要：

| AgentPact 组件 | 今天实现的部分 |
|:---------------|:-------------|
| Session Key 作为 Agent 身份 | （下一步） |
| **Smart Account（合约钱包）** | **SimpleAccount + AccountFactory** ✅ |
| Policy Engine 校验层 | （下一步） |
| Pact JSON schema 链上存储 | （下一步） |
| Revocation 机制 | （下一步） |

### SimpleAccount 架构

```
User（owner EOA）←→ SimpleAccount（合约钱包）
                          ↓
                    execute(dest, value, func)
                          ↓
                    call：转发到目标合约
```

关键点：当前版本的 `execute()` 只在 `msg.sender == owner` 时放行。下一步需要改为支持 **Session Key** 签名验证，实现 Agent 在 Pact 边界内自主执行。

### AccountFactory（CREATE2）逻辑

```
getAddress(owner, salt)
    ↓
keccak256(0xff || address(this) || salt || keccak256(bytecode))
    ↓
提前算出合约地址（零 Gas）
    ↓
createAccount(owner, salt)
    ↓
if code.length > 0: 复用已部署地址
else: new SimpleAccount{salt}(owner) → 部署
```

### 学到的知识点

1. **CREATE2 与普通 CREATE 的区别**
   - CREATE：地址 = sender + nonce，不可预测
   - CREATE2：地址 = sender + salt + bytecode hash，**可预测**后部署
   - 意义：用户可以在账户没有 Gas 费时**提前让其他人转钱**到这个地址，自己真正使用时再部署

2. **Foundry 编译流程**
   - `forge init` 创建项目骨架
   - `forge build` 编译 Solidity
   - `forge script` 运行部署脚本

3. **Foundry 部署脚本模式**
   - `vm.envUint("PRIVATE_KEY")` — 从 `.env` 读取私钥
   - `vm.startBroadcast(key)` / `vm.stopBroadcast()` — 标记上链操作
   - `vm.addr(key)` — 从私钥派生出地址

### 待办（下一步）

- [ ] 合约部署到 Sepolia 测试网
- [ ] 添加 Session Key 验证（替代直接 owner check）
- [ ] 设计并实现 Pact struct 链上存储
- [ ] Policy Engine 校验逻辑
- [ ] Agent Loop 端到端集成

---

## 打卡草稿

```
📖 Day 8 | AI × Web3 School
主题：Week 3 实战 — Foundry AA Demo（Account Abstraction）

【合约】SimpleAccount — 最小化智能合约钱包（owner + execute）
【工厂】AccountFactory — CREATE2 确定性部署与地址预测
【脚本】DeployAccount.s.sol — Sepolia 部署脚本（forge script）
【位置】Foundry/aa-demo（本地编译通过）
【衔接】5/25-5/26 的 AgentPact 设计 → 合约层实现第一步

#AIxWeb3School #WalletPermission #AgentPact #Day8 #Foundry #AccountAbstraction
```

## 提交记录

- [x] ✅ 本地 Foundry 项目搭建（`forge init` + `forge build`）
- [x] ✅ SimpleAccount 合约（owner + execute + receive）
- [x] ✅ AccountFactory 合约（CREATE2 地址预测 + 部署）
- [x] ✅ DeployAccount 部署脚本（Sepolia 配置）
- [x] ✅ 今日 daily note

## 明日计划（5/29）

- [ ] 将合约部署到 Sepolia 测试网（`forge script --rpc-url $RPC_URL`)
- [ ] 研究 ERC-4337 的 UserOperation 结构，理解 Session Key 如何工作
- [ ] 考虑用 Solidity 实现 Pact struct + Policy Engine 雏形
