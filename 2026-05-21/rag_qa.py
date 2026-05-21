#!/usr/bin/env python3
"""RAG 问答系统 — 用关键词检索 + DeepSeek API，无需下载模型"""

import json
import os
import sys
import urllib.request
import re

# ─── 读取 chunk 数据 ───────────────────────────────────────────────
with open("pages/chunks.json", encoding="utf-8") as f:
    chunks = json.load(f)

print(f"📖 加载 {len(chunks)} 个 chunk，准备检索")

# ─── 简易关键词检索（替代向量检索） ────────────────────────────────
def search_chunks(query, top_k=5):
    """把查询拆成关键词，按匹配数排序返回 top_k 个 chunk"""
    # 分词（中文按字/词，英文按空格）
    words = set(re.findall(r'[\w\u4e00-\u9fff]+', query.lower()))

    scored = []
    for i, chunk in enumerate(chunks):
        text_lower = chunk["text"].lower()
        matches = sum(1 for w in words if w in text_lower)
        if matches > 0:
            scored.append((matches, i, chunk))

    scored.sort(key=lambda x: -x[0])
    return [c for _, _, c in scored[:top_k]]

# ─── 调用 DeepSeek API ────────────────────────────────────────────
def get_api_key():
    if k := os.environ.get("DEEPSEEK_API_KEY"):
        return k
    env_path = os.path.expanduser("~/.hermes/.env")
    if os.path.exists(env_path):
        for line in open(env_path):
            line = line.strip()
            if line.startswith("DEEPSEEK_API_KEY="):
                return line.split("=", 1)[1].strip().strip("\"'")
    return None

def ask_llm(query, context_chunks):
    """把检索到的 chunk 作为上下文，发给 LLM 回答"""
    api_key = get_api_key()
    if not api_key:
        print("❌ 未找到 DEEPSEEK_API_KEY")
        return None

    # 组装上下文
    context = ""
    for i, c in enumerate(context_chunks):
        context += f"\n【来源 {i+1}】{c['source']} — {c['heading']}\n"
        context += c["text"][:600] + "\n"

    system_prompt = """你是 AI × Web3 School Handbook 的问答助手。

规则：
- 只基于提供的文档内容回答
- 如果文档中没有相关内容，明确说"文档中没有找到相关信息"
- 如果问题涉及版本或时效性，提示需要核对
- 回答格式：先给出答案，再列出引用来源

输出 JSON：
{
  "answer": "你的回答",
  "sources": ["来源1", "来源2"],
  "uncertainties": ["不确定的地方"],
  "needs_version_check": true/false
}"""

    payload = json.dumps({
        "model": "deepseek-chat",
        "messages": [
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": f"基于以下文档，回答问题：{query}\n\n文档内容：\n{context}"}
        ],
        "temperature": 0.1,
        "max_tokens": 1000,
    }).encode()

    req = urllib.request.Request(
        "https://api.deepseek.com/v1/chat/completions",
        data=payload,
        headers={
            "Content-Type": "application/json",
            "Authorization": f"Bearer {api_key}",
        }
    )
    resp = urllib.request.urlopen(req, timeout=60)
    result = json.loads(resp.read())
    content = result["choices"][0]["message"]["content"]

    # 提取 JSON
    json_str = content
    if "```json" in content:
        start = content.index("```json") + 7
        end = content.rindex("```")
        json_str = content[start:end].strip()
    elif "{" in content:
        start = content.index("{")
        end = content.rindex("}") + 1
        json_str = content[start:end]

    try:
        return json.loads(json_str)
    except:
        return {"answer": content, "sources": [], "uncertainties": [], "needs_version_check": False}

# ─── 交互式问答 ───────────────────────────────────────────────────
def main():
    print("=" * 50)
    print("📚 AI × Web3 School Handbook RAG")
    print("输入问题，或输入 exit 退出")
    print("=" * 50)

    while True:
        query = input("\n❓ ").strip()
        if not query or query.lower() in ("exit", "quit", "q"):
            break

        print("🔍 检索中...")
        results = search_chunks(query)
        print(f"   找到 {len(results)} 个相关段落")

        print("🤖 正在分析...")
        answer = ask_llm(query, results)

        if answer:
            print(f"\n📋 {answer.get('answer', '无回答')}")
            if answer.get("sources"):
                print(f"\n📎 来源: {', '.join(answer['sources'])}")
            if answer.get("uncertainties"):
                print(f"⚠️ 不确定: {'; '.join(answer['uncertainties'])}")
            if answer.get("needs_version_check"):
                print("🕐 提示：该信息可能有时效性，请核对版本")
        else:
            print("❌ 调用失败，检查 API Key")

    print("\n👋 再见！")

if __name__ == "__main__":
    main()
