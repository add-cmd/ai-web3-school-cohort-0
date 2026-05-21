#!/usr/bin/env python3
"""抓取 AI x Web3 School Handbook 的 21 篇文档页面，提取正文"""

import requests
import os
from bs4 import BeautifulSoup

BASE = "https://aiweb3.school"

PAGES = [
    # AI 基础
    ("zh/handbook/ai/llm/", "LLM"),
    ("zh/handbook/ai/prompt/", "Prompt"),
    ("zh/handbook/ai/context/", "Context"),
    ("zh/handbook/ai/rag/", "RAG"),
    ("zh/handbook/ai/agent/", "Agent"),
    ("zh/handbook/ai/frameworks/", "Frameworks"),
    ("zh/handbook/ai/vibe-coding/", "Vibe Coding"),
    ("zh/handbook/ai/mcp/", "MCP"),
    ("zh/handbook/ai/evaluation/", "Evaluation"),
    ("zh/handbook/ai/fine-tuning/", "Fine-tuning"),
    ("zh/handbook/ai/inference/", "Inference"),
    # Web3 基础
    ("zh/handbook/web3/cryptography/", "Cryptography"),
    ("zh/handbook/web3/wallet/", "Wallet"),
    ("zh/handbook/web3/smart-contract/", "Smart Contract"),
    ("zh/handbook/web3/dev-stack/", "Dev Stack"),
    ("zh/handbook/web3/network/", "Network"),
    ("zh/handbook/web3/account-abstraction/", "Account Abstraction"),
    ("zh/handbook/web3/defi/", "DeFi"),
    ("zh/handbook/web3/oracle/", "Oracle"),
    ("zh/handbook/web3/indexing/", "Indexing"),
    ("zh/handbook/web3/security/", "Security"),
]

os.makedirs("pages/html", exist_ok=True)
os.makedirs("pages/text", exist_ok=True)

for path, name in PAGES:
    url = BASE + "/" + path
    html_path = f"pages/html/{name}.html"
    text_path = f"pages/text/{name}.txt"

    if os.path.exists(html_path):
        print(f"⏭  {name} — 已存在，跳过")
        continue

    print(f"⬇  {name} — 下载中...")
    try:
        resp = requests.get(url, timeout=15)
        resp.encoding = "utf-8"

        # 保存原始 HTML
        with open(html_path, "w", encoding="utf-8") as f:
            f.write(resp.text)

        # 提取正文（保留标题层级）
        soup = BeautifulSoup(resp.text, "html.parser")
        main = soup.find("main")

        lines = []
        for el in main.find_all(["h1", "h2", "h3", "p", "li", "pre"]):
            tag = el.name
            text = el.get_text(strip=True)
            if not text:
                continue
            if tag == "h1":
                lines.append(f"\n# {text}\n")
            elif tag == "h2":
                lines.append(f"\n## {text}\n")
            elif tag == "h3":
                lines.append(f"\n### {text}\n")
            elif tag == "pre":
                lines.append(f"```\n{text}\n```")
            elif tag == "li":
                lines.append(f"- {text}")
            else:
                lines.append(text)

        with open(text_path, "w", encoding="utf-8") as f:
            f.write("\n".join(lines))

        print(f"   ✅ {name}: {len(lines)} 行")

    except Exception as e:
        print(f"   ❌ {name}: {e}")

print(f"\n🎉 完成！共 {len(PAGES)} 篇")
