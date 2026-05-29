# Day — Agent 链上足迹实验：Agentic Ping

> **日期：** 2026-05-29
> **章节：** Module D — Wallet / Permission / Safe Execution
> **核心：** Agent 调用智能合约在链上留下 Ping 足迹
> **技术栈：** Solidity ^0.8.24 + Go (GoFrame v2 + go-ethereum) + Sepolia 测试网
> **语言：** Solidity + Go

---

## 项目：Agentic Ping

**让 AI Agent 在链上留下可验证的执行足迹。** 通过 Go 后端，Agent 触发 Sepolia 上的 `PingTarget` 合约 `ping()` 函数，递增链上计数器并发射事件——实现 Agent 行为的链上锚定。

### 合约：PingTarget

```solidity
contract PingTarget {
    uint256 public pingCount;

    event Pinged(address indexed agent, uint256 timestamp);

    function ping() external {
        pingCount++;
        emit Pinged(msg.sender, block.timestamp);
    }
}
```

- **部署地址 (Sepolia):** `0xb6941F837bf7A84988BcAF82C437CCC3926CCba7`
- **核心逻辑：** 谁（agent 地址）在什么时间，第几次 ping 了。
- **源文件：** `PingTarget.sol`

### 后端：agentic-ping-backend

Go 后端服务，提供 HTTP API 供前端/Agent 触发链上 ping。

| 路由 | 方法 | 功能 |
|------|------|------|
| `/api/ping` | POST | Agent 发起链上 Ping |

**架构计划（骨架阶段）：**

```
前端/Agent → POST /api/ping
                ↓
        Go 后端 (GoFrame)
                ↓
        编码 ping() callData (0x5c36b186)
                ↓
        Alchemy Bundler (ERC-4337)
          ├─ pm_sponsorUserOperation  (Gas 代付)
          └─ eth_sendUserOperation    (上链)
                ↓
        Sepolia: PingTarget.ping()
```

**当前状态：** 骨架阶段——handler 已注册、callData 编码已实现，TODO 包括：
- [ ] 填入 Alchemy API Key
- [ ] 实现 ERC-4337 UserOperation 构造与发送
- [ ] Gas Manager 代付集成

### 项目文件

```
2026-05-29/
├── README.md
└── agentic-ping/
    ├── PingTarget.sol          ← Sepolia 部署的合约源码
    └── backend/                ← Go 后端
        ├── main.go
        ├── go.mod
        └── go.sum
```
