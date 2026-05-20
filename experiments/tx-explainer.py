#!/usr/bin/env python3
"""
交易解释器 (Transaction Explainer)
AI × Web3 School — LLM 章节最小实践

用法:
  python3 tx-explainer.py <交易哈希> [--rpc RPC_URL]

示例:
  python3 tx-explainer.py 0x2b8e9c5c8e7d6f5a4b3c2d1e0f9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1
  python3 tx-explainer.py --sepolia <交易哈希>

把模型生成、链上事实、来源边界、不确定性分开。
"""

import json, sys, os, urllib.request, urllib.error, argparse
from datetime import datetime

# ─── 配置 ───────────────────────────────────────────────────────

# 公共 RPC 节点（免费，无需 API key）
MAINNET_RPC = "https://ethereum-rpc.publicnode.com"
SEPOLIA_RPC = "https://ethereum-sepolia-rpc.publicnode.com"

# DeepSeek API (先从环境变量读，再从 Hermes .env 读取)
DEEPSEEK_API_KEY = os.environ.get("DEEPSEEK_API_KEY", "")
if not DEEPSEEK_API_KEY:
    # 从 Hermes 的 .env 读取
    env_path = os.path.expanduser("~/.hermes/.env")
    if os.path.exists(env_path):
        with open(env_path) as f:
            for line in f:
                line = line.strip()
                if line.startswith("DEEPSEEK_API_KEY="):
                    DEEPSEEK_API_KEY = line.split("=", 1)[1].strip().strip('"').strip("'")
                    break
DEEPSEEK_MODEL = "deepseek-chat"
DEEPSEEK_URL = "https://api.deepseek.com/v1/chat/completions"


# ─── 第一步：从 RPC 获取交易数据 ────────────────────────────────

def rpc_call(rpc_url: str, method: str, params: list) -> dict:
    """调用 JSON-RPC（走 curl，更稳定）"""
    import subprocess
    payload = json.dumps({
        "jsonrpc": "2.0",
        "method": method,
        "params": params,
        "id": 1
    })
    result = subprocess.run(
        ["curl", "-s", "-m", "10", rpc_url,
         "-X", "POST",
         "-H", "Content-Type: application/json",
         "-d", payload],
        capture_output=True, text=True, timeout=15
    )
    if result.returncode != 0:
        raise ValueError(f"curl 请求失败: {result.stderr}")
    data = json.loads(result.stdout)
    if "error" in data:
        raise ValueError(f"RPC 错误: {data['error']}")
    return data["result"]


def fetch_tx(rpc_url: str, tx_hash: str) -> dict:
    """获取交易详情"""
    result = rpc_call(rpc_url, "eth_getTransactionByHash", [tx_hash])
    if result is None:
        raise ValueError(f"交易不存在: {tx_hash}")
    return result


def fetch_tx_receipt(rpc_url: str, tx_hash: str) -> dict:
    """获取交易回执"""
    result = rpc_call(rpc_url, "eth_getTransactionReceipt", [tx_hash])
    return result if result else {}


def fetch_contract_bytecode(rpc_url: str, addr: str) -> str:
    """获取合约字节码"""
    return rpc_call(rpc_url, "eth_getCode", [addr, "latest"])


def hex_to_eth(hex_str: str, decimals: int = 18) -> str:
    """将 hex 金额转为 ETH 单位（保留6位小数）"""
    if not hex_str or hex_str == "0x":
        return "0"
    value = int(hex_str, 16)
    return f"{value / 10**decimals:.6f}"


def format_address(addr: str) -> str:
    """截短地址显示"""
    if not addr:
        return "N/A"
    return f"{addr[:6]}...{addr[-4:]}"


# ─── 第三步：结构化交易数据 ──────────────────────────────────────

def parse_transaction(tx: dict, receipt: dict) -> dict:
    """将 RPC 原始数据转成可读结构"""
    
    # 基本信息
    tx_hash = tx.get("hash", "")
    block = int(tx.get("blockNumber", "0x0"), 16)
    
    from_addr = tx.get("from", "")
    to_addr = tx.get("to", "")
    value_hex = tx.get("value", "0x0")
    value_eth = hex_to_eth(value_hex)
    
    gas_limit = int(tx.get("gas", "0x0"), 16)
    gas_price_wei = int(tx.get("gasPrice", "0x0"), 16)
    gas_price_gwei = gas_price_wei / 1e9
    
    # 已使用的 gas（从 receipt）
    gas_used = int(receipt.get("gasUsed", "0x0"), 16)
    
    # 交易费用
    tx_fee_wei = gas_used * gas_price_wei
    tx_fee_eth = tx_fee_wei / 1e18
    
    # input data（函数调用）
    input_data = tx.get("input", "0x")
    has_data = input_data and input_data != "0x"
    
    # 解析函数选择器（前4字节）
    selector = input_data[:10] if has_data and len(input_data) >= 10 else None
    
    # 状态
    status = "成功" if receipt.get("status") == "0x1" else "失败"
    
    # 事件日志
    logs = receipt.get("logs", [])
    parsed_logs = []
    for log in logs[:5]:  # 最多显示5条
        parsed_logs.append({
            "address": log.get("address", ""),
            "topics": [t for t in log.get("topics", [])],
            "data": log.get("data", "0x")[:100]  # 截短防止过长
        })
    
    return {
        "tx_hash": tx_hash,
        "block": block,
        "from": from_addr,
        "to": to_addr,
        "value_eth": value_eth,
        "gas_limit": gas_limit,
        "gas_used": gas_used,
        "gas_price_gwei": f"{gas_price_gwei:.2f}",
        "tx_fee_eth": tx_fee_eth,
        "status": status,
        "has_input_data": has_data,
        "function_selector": selector,
        "input_data_preview": input_data[:150] if has_data else None,
        "logs_count": len(logs),
        "logs": parsed_logs
    }


# ─── 第四步：调用 DeepSeek API 做解释 ────────────────────────────

def explain_transaction(tx_info: dict) -> str:
    """让 LLM 解释这笔交易"""
    
    system_prompt = """你是交易解释助手。根据提供的链上数据，生成交易解释。

你的输出必须严格遵守以下格式（JSON）：

{
  "summary": "一句话总结用户做了什么",
  "action_type": "transfer | contract_interaction | deploy | other",
  "assets_involved": [
    {"asset": "ETH / USDC / ...", "amount": "数量", "direction": "in | out"}
  ],
  "addresses": {
    "from": "发起方（格式化）",
    "to": "接收方 / 合约（格式化）",
    "note": "地址说明"
  },
  "on_chain_facts": ["这些是链上确认的事实"],
  "model_inferences": ["这些是我的推断，不是链上事实"],
  "uncertainties": ["我不确定的地方"],
  "user_checks": ["签名前用户应该检查这些"]
}"""

    user_prompt = f"""请解释以下以太坊交易：

## 交易基本信息
- 哈希: {tx_info['tx_hash']}
- 区块: {tx_info['block']}
- 发起方: {tx_info['from']}
- 接收方: {tx_info['to']}
- 转账金额: {tx_info['value_eth']} ETH
- Gas 上限: {tx_info['gas_limit']}
- Gas 已用: {tx_info['gas_used']}
- Gas 价格: {tx_info['gas_price_gwei']} Gwei
- 交易费: {tx_info['tx_fee_eth']:.8f} ETH
- 状态: {tx_info['status']}

## 函数调用
{"有调用数据" if tx_info['has_input_data'] else "无调用数据（纯转账）"}
{"函数选择器: " + tx_info['function_selector'] if tx_info['function_selector'] else ""}
{"Input data 预览: " + tx_info['input_data_preview'] if tx_info['input_data_preview'] else ""}

## 事件日志
共 {tx_info['logs_count']} 条日志
{"第一条日志地址: " + tx_info['logs'][0]['address'] if tx_info['logs'] else "无日志"}
{"第一条日志 topics: " + str(tx_info['logs'][0]['topics']) if tx_info['logs'] else ""}
"""

    if not DEEPSEEK_API_KEY:
        return "⚠ 未配置 DEEPSEEK_API_KEY，跳过 LLM 解释"

    payload = {
        "model": DEEPSEEK_MODEL,
        "messages": [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": user_prompt}
        ],
        "temperature": 0.3,
        "max_tokens": 1024
    }
    
    req = urllib.request.Request(
        DEEPSEEK_URL,
        data=json.dumps(payload).encode(),
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {DEEPSEEK_API_KEY}"
        }
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            result = json.loads(resp.read())
            content = result["choices"][0]["message"]["content"]
            return content
    except Exception as e:
        return f"❌ LLM 调用失败: {e}"


# ─── 主程序 ──────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="交易解释器 — AI × Web3 School")
    parser.add_argument("tx_hash", help="交易哈希 (0x...)")
    parser.add_argument("--rpc", help="自定义 RPC URL")
    parser.add_argument("--sepolia", action="store_true", help="使用 Sepolia 测试网")
    args = parser.parse_args()
    
    tx_hash = args.tx_hash.strip()
    if not tx_hash.startswith("0x"):
        tx_hash = "0x" + tx_hash
    
    # 选择网络
    if args.sepolia:
        rpc_url = SEPOLIA_RPC
        network = "Sepolia 测试网"
    elif args.rpc:
        rpc_url = args.rpc
        network = "自定义网络"
    else:
        rpc_url = MAINNET_RPC
        network = "以太坊主网"
    
    print(f"{'='*60}")
    print(f"🔍 交易解释器")
    print(f"   网络: {network}")
    print(f"   哈希: {tx_hash}")
    print(f"{'='*60}\n")
    
    # 第一步：获取交易数据
    print("📡 正在从链上获取交易数据...", end=" ", flush=True)
    try:
        tx = fetch_tx(rpc_url, tx_hash)
        receipt = fetch_tx_receipt(rpc_url, tx_hash)
        print("✅ 成功\n")
    except Exception as e:
        print(f"❌ 失败: {e}")
        sys.exit(1)
    
    # 第二步：解析
    print("🔧 解析交易数据...")
    tx_info = parse_transaction(tx, receipt)
    
    # 打印链上数据摘要
    print(f"\n{'─'*60}")
    print(f"📋 链上数据摘要（来源: RPC 节点）")
    print(f"{'─'*60}")
    print(f"  发起方:     {format_address(tx_info['from'])}")
    print(f"  接收方:     {format_address(tx_info['to'])}")
    print(f"  转账金额:   {tx_info['value_eth']} ETH")
    print(f"  Gas 价格:   {tx_info['gas_price_gwei']} Gwei")
    print(f"  交易费:     {tx_info['tx_fee_eth']:.8f} ETH")
    print(f"  状态:       {tx_info['status']}")
    print(f"  区块:       {tx_info['block']}")
    print(f"  日志:       {tx_info['logs_count']} 条")
    if tx_info['function_selector']:
        print(f"  函数选择器: {tx_info['function_selector']}")
    print(f"{'─'*60}\n")
    
    # 第三步：检查接收方是否是合约
    print("📄 检查接收方是否为合约...", end=" ", flush=True)
    try:
        bytecode = fetch_contract_bytecode(rpc_url, tx_info['to'])
        is_contract = bytecode is not None and bytecode != "0x"
        print(f"{'✅ 是合约' if is_contract else 'ℹ️ 是 EOA 地址'}")
    except:
        print("⚠ 无法检查")
        is_contract = False
    
    # 第四步：调用 LLM 解释
    print(f"\n{'─'*60}")
    print(f"🤖 LLM 解释（来源: DeepSeek / 概率模型推断）")
    print(f"{'─'*60}")
    explanation = explain_transaction(tx_info)
    
    if explanation.startswith("⚠") or explanation.startswith("❌"):
        print(f"  {explanation}")
    else:
        try:
            parsed = json.loads(explanation)
            print(f"\n  📝 摘要: {parsed.get('summary', 'N/A')}")
            print(f"  🔢 动作类型: {parsed.get('action_type', 'N/A')}")
            
            print(f"\n  💰 涉及资产:")
            for asset in parsed.get("assets_involved", []):
                print(f"    • {asset.get('amount', '?')} {asset.get('asset', '?')} ({asset.get('direction', '?')})")
            
            print(f"\n  ✅ 链上事实（可信）:")
            for f in parsed.get("on_chain_facts", []):
                print(f"    • {f}")
            
            print(f"\n  ⚠ 模型推断（可能不准确）:")
            for i in parsed.get("model_inferences", []):
                print(f"    • {i}")
            
            print(f"\n  ❓ 不确定之处:")
            for u in parsed.get("uncertainties", []):
                print(f"    • {u}")
            
            print(f"\n  🔐 签名前检查清单:")
            for c in parsed.get("user_checks", []):
                print(f"    ☐ {c}")
        except json.JSONDecodeError:
            print(f"\n  原始输出:\n{explanation}")
    
    print(f"\n{'='*60}")
    print(f"✅ 完成！注意区分链上事实 vs 模型推断")
    print(f"{'='*60}")


if __name__ == "__main__":
    main()
