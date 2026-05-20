#!/usr/bin/env python3
"""
Day 3 — Context 最小实践：钱包授权检查 Agent
=============================================
核心概念：Context Engineering — 为 LLM 装配可信上下文

每个上下文片段都带标签：
  [SOURCE: rpc|cache|user] [TRUST: high|medium|low] [FRESH: realtime|cached|user-input]

对比 Day 1（交易解释器 — 事后分析）和 Day 2（风险分析器 — Prompt 设计）：
  Day 3 的重点不是 Prompt 怎么写，而是 "放什么进上下文、从哪里来、可信度如何"。
"""

import os, sys, json, time, hashlib
from datetime import datetime

# ── Configuration (from environment, same pattern as tx-explainer.py) ──────────
RPC_URL = os.getenv("RPC_URL", "https://ethereum-sepolia-rpc.publicnode.com")
API_KEY = os.getenv("DEEPSEEK_API_KEY") or os.getenv("HERMES_DEEPSEEK_KEY")
MODEL = "deepseek-chat"

if not API_KEY:
    # Try reading from Hermes .env
    env_path = os.path.expanduser("~/.hermes/.env")
    if os.path.exists(env_path):
        with open(env_path) as f:
            for line in f:
                line = line.strip()
                if line.startswith("DEEPSEEK_API_KEY="):
                    API_KEY = line.split("=", 1)[1].strip().strip('"').strip("'")
                    break

if not API_KEY:
    print("❌ 需要设置 DEEPSEEK_API_KEY 环境变量")
    sys.exit(1)


# ══════════════════════════════════════════════════════════════════════════════
# Context Sources — 每个源都返回带元数据的数据块
# ══════════════════════════════════════════════════════════════════════════════

def ctx_block(source, trust, freshness, label, data, warning=None):
    """创建一个带元数据的上下文块"""
    return {
        "source": source,
        "trust": trust,
        "freshness": freshness,
        "label": label,
        "data": data,
        "warning": warning,
        "ts": datetime.utcnow().isoformat()
    }

def query_rpc(method, params):
    """通用 RPC 调用 — 使用 curl 避免 403"""
    import subprocess
    payload = json.dumps({
        "jsonrpc": "2.0", "id": 1,
        "method": method, "params": params
    })
    result = subprocess.run(
        ["curl", "-s", "-m", "15", RPC_URL,
         "-X", "POST",
         "-H", "Content-Type: application/json",
         "-d", payload],
        capture_output=True, text=True, timeout=20
    )
    return json.loads(result.stdout)["result"]

###############################################################################
# ⚡ SOURCE 1: RPC (实时链上数据) — TRUST: HIGH
###############################################################################

def fetch_chain_context():
    """chain id + 当前区块号"""
    chain_id = query_rpc("eth_chainId", [])
    block_num = query_rpc("eth_blockNumber", [])
    return ctx_block(
        source="rpc", trust="high", freshness="realtime",
        label="Chain Context",
        data={
            "chain_id": int(chain_id, 16),
            "block_number": int(block_num, 16),
        }
    )

def fetch_token_info(token_address):
    """代币基本信息（符号、名称、精度）— 通过 RPC 读合约"""
    # 构造 eth_call — name() / symbol() / decimals()
    # selector: name()=0x06fdde03, symbol()=0x95d89b41, decimals()=0x313ce567
    results = {}
    selectors = {
        "name": "0x06fdde03",
        "symbol": "0x95d89b41",
        "decimals": "0x313ce567",
    }
    addr = token_address.lower()
    for field, sel in selectors.items():
        try:
            data = query_rpc("eth_call", [{
                "to": addr, "data": sel
            }, "latest"])
            # Decode hex string (ABI-encoded string)
            if data and data != "0x":
                raw = bytes.fromhex(data[2:])
                offset = int.from_bytes(raw[0:32], 'big') if len(raw) >= 32 else 0
                length = int.from_bytes(raw[offset+0:offset+32], 'big') if len(raw) >= offset+32 else 0
                str_bytes = raw[offset+32:offset+32+length] if len(raw) >= offset+32+length else b""
                try:
                    results[field] = str_bytes.decode('utf-8', errors='replace')
                except:
                    results[field] = f"0x{raw.hex()[:40]}..."
            else:
                results[field] = "N/A"
        except Exception as e:
            results[field] = f"ERROR: {e}"

    # decimals is uint8
    try:
        data = query_rpc("eth_call", [{"to": addr, "data": "0x313ce567"}, "latest"])
        if data and data != "0x":
            results["decimals"] = int(data, 16)
    except:
        results["decimals"] = 18

    return ctx_block(
        source="rpc", trust="high", freshness="realtime",
        label=f"Token Info ({token_address})",
        data=results
    )

def fetch_user_state(token_address, user_address, spender_address):
    """用户的 balance 和 allowance"""
    addr_token = token_address.lower()
    addr_user = user_address.lower()
    addr_spender = spender_address.lower()

    # balanceOf(user)
    bal_sel = "0x70a08231" + addr_user[2:].zfill(64)
    try:
        bal_data = query_rpc("eth_call", [{"to": addr_token, "data": bal_sel}, "latest"])
        balance = int(bal_data, 16) if bal_data else 0
    except:
        balance = None

    # allowance(user, spender)
    allow_sel = "0xdd62ed3e" + addr_user[2:].zfill(64) + addr_spender[2:].zfill(64)
    try:
        allow_data = query_rpc("eth_call", [{"to": addr_token, "data": allow_sel}, "latest"])
        allowance = int(allow_data, 16) if allow_data else 0
    except:
        allowance = None

    return ctx_block(
        source="rpc", trust="high", freshness="realtime",
        label="User On-Chain State",
        data={
            "balance": balance,
            "allowance": allowance,
            "user_address": user_address,
            "spender_address": spender_address,
        }
    )

###############################################################################
# 🟡 SOURCE 2: Cache (已知可信列表) — TRUST: MEDIUM
###############################################################################

def load_spender_cache():
    """
    模拟一个本地缓存的可信/黑名单 spender 列表。
    真实场景中这可以来自配置文件、数据库、或链上注册表。
    """
    cache = {
        # 已知安全的 DeFi 协议 (Sepolia 测试用示例地址)
        "trusted": [
            "0x1f9840a85d5af5bf1d1762f925bdaddc4201f984",  # Uniswap (mainnet, 示例)
            "0x7a250d5630b4cf539739df2c5dacb4c659f2488d",  # Uniswap V2 Router
            "0xd9e1ce17f2641f24ae83637ab66a2cca9c378b9f",  # SushiSwap Router
        ],
        # 已知恶意/钓鱼地址 (示例)
        "blacklisted": [
            "0xdead000000000000000000000000000000000000",
        ]
    }

    now_ts = time.time()
    # 模拟缓存创建时间 — 假设上次更新是 1 小时前
    cache_ts = now_ts - 3600

    # 判断 spender 是否在列表中
    return ctx_block(
        source="cache", trust="medium", freshness="cached",
        label="Spender Trust List (local cache)",
        data={
            "trusted_count": len(cache["trusted"]),
            "blacklisted_count": len(cache["blacklisted"]),
            "cache_age_seconds": int(now_ts - cache_ts),
            "list": cache,
        }
    )

###############################################################################
# 🔴 SOURCE 3: User Input (用户提供的参数和意图) — TRUST: LOW
###############################################################################

def parse_user_input():
    """从命令行参数解析用户输入 — 模拟用户提供的交易参数和意图"""
    if len(sys.argv) < 4:
        print("用法: python3 context_approval_check.py <token_address> <spender_address> <amount> [user_intent]")
        print()
        print("示例:")
        print("  # 正常 approve")
        print(f"  python3 {sys.argv[0]} 0x... 0x... 1000000 \"我要在 Uniswap 上 swap USDC\"")
        print()
        print("  # 钓鱼尝试")
        print(f"  python3 {sys.argv[0]} 0x... 0xdead... 115792089237316195423570985008687907853269984665640564039457584007913129639935 \"领空投需要先授权\"")
        sys.exit(1)

    token = sys.argv[1]
    spender = sys.argv[2]
    amount = sys.argv[3]
    intent = sys.argv[4] if len(sys.argv) > 4 else "未提供"

    is_infinite = False
    MAX_APPROVAL = 2**256 - 1
    try:
        if int(amount) >= MAX_APPROVAL / 2:
            is_infinite = True
    except:
        pass

    return ctx_block(
        source="user", trust="low", freshness="user-input",
        label="User-Provided Transaction Parameters",
        data={
            "token_address": token,
            "spender_address": spender,
            "approve_amount": amount,
            "is_infinite_approval": is_infinite,
            "user_intent": intent,
        },
        warning="用户提供的数据未经验证。dApp 页面描述可能被篡改，意图可能与实际交易不符。"
    )

###############################################################################
# ⚡ SOURCE 4: Simulation (模拟执行) — TRUST: HIGH
###############################################################################

def simulate_approval(token_address, spender_address, amount, user_address):
    """用 eth_call 模拟 approve 执行，检查是否会 revert"""
    addr_token = token_address.lower()
    addr_spender = spender_address.lower()

    # approve(spender, amount)
    # selector: 0x095ea7b3
    amt_hex = hex(int(amount))[2:].zfill(64) if amount and amount != "infinite" else "f" * 64
    approve_data = "0x095ea7b3" + addr_spender[2:].zfill(64) + amt_hex

    sim_result = {}
    try:
        result = query_rpc("eth_call", [{
            "from": user_address,
            "to": addr_token,
            "data": approve_data,
        }, "latest"])
        sim_result["would_revert"] = False
        # decode bool return — 0x...{62}1 = true
        if result and len(result) >= 66:
            sim_result["expected_return"] = "true" if result[-1] == '1' else "false"
        else:
            sim_result["expected_return"] = result[:66] if result else "unknown"
    except Exception as e:
        sim_result["would_revert"] = True
        sim_result["error"] = str(e)

    return ctx_block(
        source="simulation", trust="high", freshness="realtime",
        label="Transaction Simulation (eth_call)",
        data=sim_result
    )


# ══════════════════════════════════════════════════════════════════════════════
# Context Assembly — 装配最终上下文
# ══════════════════════════════════════════════════════════════════════════════

def assemble_context(blocks):
    """
    将所有上下文块组装成给 LLM 的结构化输入。
    每个块保留完整的元数据 — LLM 可以据此判断信息可信度。
    """
    # 按可信度排序：high → medium → low
    trust_order = {"high": 0, "medium": 1, "low": 2}
    blocks.sort(key=lambda b: trust_order.get(b["trust"], 99))

    lines = []
    lines.append("=" * 60)
    lines.append("CONTEXT ASSEMBLY — 上下文装配 (Context Engineering Demo)")
    lines.append(f"Assembled at: {datetime.utcnow().isoformat()}")
    lines.append("=" * 60)
    lines.append("")

    for i, block in enumerate(blocks, 1):
        trust_color = {"high": "🟢", "medium": "🟡", "low": "🔴"}
        freshness_icon = {"realtime": "⚡", "cached": "💾", "user-input": "👤"}

        lines.append(f"[Block {i}] {trust_color.get(block['trust'], '❓')} {block['label']}")
        lines.append(f"  Source:    {block['source']}")
        lines.append(f"  Trust:     {block['trust'].upper()} {'✅ 可作事实依据' if block['trust'] == 'high' else '⚠️ 需交叉验证' if block['trust'] == 'medium' else '❌ 不可信，仅供上下文参考'}")
        lines.append(f"  Freshness: {freshness_icon.get(block['freshness'], '❓')} {block['freshness']}")
        if block.get("warning"):
            lines.append(f"  ⚠️  {block['warning']}")
        lines.append(f"  Data:")
        for key, val in block["data"].items():
            val_str = str(val)
            if len(val_str) > 120:
                val_str = val_str[:120] + "..."
            lines.append(f"    {key}: {val_str}")
        lines.append("")

    return "\n".join(lines)


# ══════════════════════════════════════════════════════════════════════════════
# LLM Call — 把装配好的上下文发给 LLM
# ══════════════════════════════════════════════════════════════════════════════

def call_llm(context_text, user_address):
    """发送装配好的上下文到 LLM 进行分析"""
    system_prompt = """你是钱包授权安全检查 Agent。你的职责是：

1. **理解上下文来源的可信度**：RPC 数据（high）> 缓存数据（medium）> 用户输入（low）
2. 明确标注每个结论的依据来源
3. 区分布告"链上事实"、"合理推断"和"不确定性"
4. 给出清晰的风险等级和操作建议

输出格式必须为 JSON：
{
  "risk_level": "low|medium|high|critical",
  "summary": "一句话总结",
  "analysis": {
    "on_chain_facts": ["从 RPC/simulation 确认的事实"],
    "inferences": ["基于缓存+链上数据的推断"],
    "uncertainties": ["不能确认的事项"],
    "user_input_warnings": ["用户提供的数据中值得注意的问题"]
  },
  "recommendation": "建议用户采取的操作",
  "requires_human_approval": true/false
}"""

    payload = {
        "model": MODEL,
        "messages": [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": f"""请分析以下钱包授权请求。

{context_text}

用户地址: {user_address}

请判断这个 approve 操作是否有风险。注意：
- 来源标注为 [TRUST: HIGH] 的数据可作事实依据
- 来源标注为 [TRUST: MEDIUM] 的数据需说明局限性
- 来源标注为 [TRUST: LOW] 的数据不可信，需提示用户验证
"""}
        ],
        "temperature": 0.1,
        "max_tokens": 2000,
    }

    import subprocess
    req = subprocess.run(
        ["curl", "-s", "-m", "60", "https://api.deepseek.com/v1/chat/completions",
         "-X", "POST",
         "-H", "Content-Type: application/json",
         "-H", f"Authorization: Bearer {API_KEY}",
         "-d", json.dumps(payload)],
        capture_output=True, text=True, timeout=65
    )
    result = json.loads(req.stdout)
    return result["choices"][0]["message"]["content"]


# ══════════════════════════════════════════════════════════════════════════════
# Main
# ══════════════════════════════════════════════════════════════════════════════

def main():
    print("╔══════════════════════════════════════════════╗")
    print("║  Day 3 — Context 最小实践                    ║")
    print("║  钱包授权检查 Agent · Context Engineering    ║")
    print("╚══════════════════════════════════════════════╝")
    print()

    # 1. 解析用户输入
    print("📥 [SOURCE: user] 解析用户输入的交易参数...")
    user_block = parse_user_input()
    token = user_block["data"]["token_address"]
    spender = user_block["data"]["spender_address"]
    amount = user_block["data"]["approve_amount"]
    print(f"   Token:   {token}")
    print(f"   Spender: {spender}")
    print(f"   Amount:  {amount}")
    print(f"   Intent:  {user_block['data']['user_intent']}")
    print(f"   Infinite:{user_block['data']['is_infinite_approval']}")
    print()

    # 获取用户地址（模拟，实际从钱包/参数获取）
    user_address = os.getenv("USER_ADDRESS", "0x0000000000000000000000000000000000000001")

    # 2. 从 RPC 拿实时数据
    blocks = []

    print("⚡ [SOURCE: rpc] 查询链上数据 (TRUST: HIGH)...")
    blocks.append(fetch_chain_context())
    print(f"   Chain ID: {blocks[-1]['data']['chain_id']}, Block: {blocks[-1]['data']['block_number']}")

    blocks.append(fetch_token_info(token))
    print(f"   Token: {blocks[-1]['data'].get('symbol', '?')} ({blocks[-1]['data'].get('name', '?')})")

    blocks.append(fetch_user_state(token, user_address, spender))
    bal = blocks[-1]["data"]["balance"]
    allow = blocks[-1]["data"]["allowance"]
    print(f"   Balance: {bal}, Allowance: {allow}")

    # 3. 从缓存拿可信列表
    print("💾 [SOURCE: cache] 查询本地缓存 (TRUST: MEDIUM)...")
    blocks.append(load_spender_cache())
    print(f"   Trusted: {blocks[-1]['data']['trusted_count']}, Blacklisted: {blocks[-1]['data']['blacklisted_count']}")

    # 4. 模拟执行
    print("⚡ [SOURCE: simulation] 模拟交易执行 (TRUST: HIGH)...")
    blocks.append(simulate_approval(token, spender, amount, user_address))
    print(f"   Would revert: {blocks[-1]['data']['would_revert']}")

    # 5. 用户输入放最后
    blocks.append(user_block)

    # 6. 装配上下文
    print("\n📦 装配上下文...\n")
    context = assemble_context(blocks)

    # 打印装配好的上下文
    print("\033[36m" + context + "\033[0m")
    print()

    # 7. 调 LLM
    print("🤖 发送到 LLM 分析...")
    print(f"   Model: {MODEL}")
    print()

    try:
        response = call_llm(context, user_address)
        print("═" * 60)
        print("📋 LLM 分析结果")
        print("═" * 60)
        print()

        # Try to pretty-print JSON
        try:
            parsed = json.loads(response)
            print(json.dumps(parsed, indent=2, ensure_ascii=False))
        except:
            print(response)

    except Exception as e:
        print(f"❌ LLM 调用失败: {e}")
        print("   但仍可查看上下文装配结果（核心产出）")


if __name__ == "__main__":
    main()
