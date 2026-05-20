# Day 1 — 交易解释器

> **日期：** 2026-05-18
> **章节：** LLM / Network / Wallet / Smart Contract
> **技术栈：** Python + DeepSeek API

---

## 项目：交易解释器

`tx-explainer.py` 是一个独立的 Python 脚本，通过 RPC 获取链上交易数据，调用 DeepSeek API 生成结构化解释。

### 核心设计

严格区分**链上事实**与**模型推断**，符合「LLM 是推理入口，不是最终验证」原则：

```
交易哈希
  → RPC 获取链上数据（交易详情、回执、合约字节码）
    → 组装 Prompt（链上事实 + ABI 上下文）
      → DeepSeek 结构化输出
        ├── ✅ on_chain_facts     链上可验证的事实
        ├── ⚠  model_inferences   模型推测的解释
        ├── ❓  uncertainties      模型无法确认的部分
        └── 🔐  user_checks       签名前检查清单
```

### 快速开始

```bash
# 确保 Hermes .env 中有 DEEPSEEK_API_KEY
python3 tx-explainer.py <交易哈希>
```

### 示例

```bash
python3 tx-explainer.py 0xf40fcde4e23a7a99f82f0a4375c43bca51549e07696f80681460535f708556dc
```

### 依赖

- `requests` — 调 DeepSeek API 和 RPC 节点
- 环境变量 `DEEPSEEK_API_KEY`（从 Hermes `.env` 读取）

---

## 学习笔记（摘录）

### LLM 核心认知

| 概念 | 一句话理解 |
|------|-----------|
| **Token** | 模型处理文本的最小单位，中文 1 字 ≈ 1-2 token |
| **Embedding** | 把文本转成向量，衡量语义相似度 |
| **Transformer** | 通过 Attention 关注上下文不同位置 |
| **Hallucination** | 模型生成看似合理但不真实的内容 |
| **Multimodal** | 能处理文本+图片/音频 |

### 最小实践收获

- LLM API 调用流程：组装 Prompt → 发送请求 → 解析结构化输出
- 链上数据必须通过 RPC 获取，不能依赖模型「回忆」
- ABI 作为 Few-shot 示例传递给模型，提高函数识别准确率
