# 📁 foundry-dapp — 项目总览

完整的全栈 dApp 项目：Foundry (Solidity) + Go (后端) + React (前端)
部署在 Sepolia 测试网。

## 项目结构

```
foundry-dapp/
├── contracts/              # Foundry 智能合约
│   ├── src/Counter.sol     # Counter 合约
│   ├── test/Counter.t.sol  # 测试文件 (7 tests)
│   ├── script/             # 部署脚本
│   └── foundry.toml        # Foundry 配置
├── backend/                # Go 后端
│   ├── main.go             # API 服务器
│   └── go.mod / go.sum
├── frontend/               # React 前端 (Vite)
│   ├── src/App.jsx         # 主页面
│   └── package.json
└── README.md
```

## 快速开始

```bash
# 1. 编译合约
cd contracts && forge build

# 2. 测试合约
forge test

# 3. 部署到 Sepolia（需要 PRIVATE_KEY 和 Sepolia ETH）
export PRIVATE_KEY=0x...
forge script script/DeployCounter.s.sol --rpc-url sepolia --broadcast

# 4. 设置合约地址
export CONTRACT_ADDRESS=0x<部署的合约地址>

# 5. 启动后端
cd backend && go run . &

# 6. 启动前端
cd frontend && npm run dev
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `PRIVATE_KEY` | 部署用的私钥 (0x...) | — |
| `CONTRACT_ADDRESS` | 已部署的合约地址 | — |
| `RPC_URL` | Sepolia RPC 节点 | https://ethereum-sepolia-rpc.publicnode.com |
| `PORT` | 后端端口 | 8080 |
| `VITE_CONTRACT_ADDRESS` | 前端合约地址 (frontend/.env) | — |
