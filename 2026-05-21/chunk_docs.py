#!/usr/bin/env python3
"""将抓取的 21 篇文档按 h2 切 chunk，短小节自动合并"""

import os, json, re

TEXT_DIR = "pages/text"
OUTPUT = "pages/chunks.json"

chunks = []

for fname in sorted(os.listdir(TEXT_DIR)):
    if not fname.endswith(".txt"):
        continue
    name = fname.replace(".txt", "")
    filepath = os.path.join(TEXT_DIR, fname)

    with open(filepath, encoding="utf-8") as f:
        text = f.read()

    # 按 h2 拆分
    sections = re.split(r"\n## ", text)
    page_chunks = []

    for i, sec in enumerate(sections):
        if i == 0:
            # 第一篇 h2 之前的内容属于 h1（标题/摘要）
            lines = sec.strip().split("\n")
            title = lines[0].replace("# ", "").strip() if lines[0].startswith("# ") else name
            body = "\n".join(lines[1:]).strip()
            heading = title
        else:
            lines = sec.split("\n")
            heading = lines[0].strip()
            body = "\n".join(lines[1:]).strip()

        if not body:
            continue

        page_chunks.append({
            "source": name,
            "heading": heading,
            "text": f"## {heading}\n{body}"
        })

    # 智能合并：小于 100 字的 chunk 合并到上一个
    merged = []
    for c in page_chunks:
        if merged and len(c["text"]) < 100:
            # 合并到上一个
            merged[-1]["text"] += f"\n\n{c['text']}"
            merged[-1]["heading"] += f" + {c['heading']}"
        else:
            merged.append(c)
    page_chunks = merged

    chunks.extend(page_chunks)
    print(f"  {name}: {len(page_chunks)} chunks")

# 输出
with open(OUTPUT, "w", encoding="utf-8") as f:
    json.dump(chunks, f, ensure_ascii=False, indent=2)

total_chars = sum(len(c["text"]) for c in chunks)
print(f"\n✅ 共 {len(chunks)} 个 chunk，总字数 {total_chars}")
