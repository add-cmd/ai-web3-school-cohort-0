#!/usr/bin/env python3
"""测试风险分析 API"""
import json, urllib.request

API = "http://localhost:8080/api/analyze"

tests = [
    {
        "name": "测试1: Counter.increment (低风险)",
        "payload": {
            "to": "0x6d8521408b803813a1A963f511C74fB96ea23bd2",
            "data": "0xd09de08a",
            "value": "0",
            "function": "increment",
            "user_intent": "测试计数功能"
        }
    },
    {
        "name": "测试2: Token.approve 无限授权 (高风险)",
        "payload": {
            "to": "0x62E3395eCFa2d18afB8F0cfbB1FA55948Dd03674",
            "data": "0x095ea7b3000000000000000000000000dead000000000000000000000000000000000000ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
            "value": "0",
            "function": "approve",
            "user_intent": "给一个合约授权"
        }
    },
    {
        "name": "测试3: Token.transfer (意图不匹配)",
        "payload": {
            "to": "0x62E3395eCFa2d18afB8F0cfbB1FA55948Dd03674",
            "data": "0xa9059cbb000000000000000000000000dead0000000000000000000000000000000000000000000000000000000000000064",
            "value": "0",
            "function": "transfer",
            "args": "to=0xdead..., amount=100",
            "user_intent": "转账给我的朋友Alice"
        }
    }
]

for t in tests:
    print(f"\n{'='*60}")
    print(f"  {t['name']}")
    print(f"{'='*60}")
    req = urllib.request.Request(
        API,
        data=json.dumps(t["payload"]).encode(),
        headers={"Content-Type": "application/json"}
    )
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            result = json.loads(resp.read())
        print(f"  risk_level:        {result.get('risk_level', 'N/A')}")
        print(f"  requires_approval: {result.get('requires_human_approval', 'N/A')}")
        print(f"  intent_match:      {result.get('intent_match', 'N/A')}")
        print(f"  summary:           {result.get('summary', 'N/A')}")
        if result.get('intent_note'):
            print(f"  intent_note:       {result['intent_note']}")
        if result.get('uncertainties'):
            for u in result['uncertainties']:
                print(f"  ❓ {u}")
        if result.get('recommended_user_checks'):
            for c in result['recommended_user_checks']:
                print(f"  🔐 {c}")
    except Exception as e:
        print(f"  ❌ 失败: {e}")
