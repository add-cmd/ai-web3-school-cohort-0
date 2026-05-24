#!/usr/bin/env python3
"""读取 chunks.json，生成 embedding 并存入 ChromaDB"""

import json
import chromadb

# 1. 读取 chunk 数据
with open("pages/chunks.json", encoding="utf-8") as f:
    chunks = json.load(f)

print(f"📖 读取 {len(chunks)} 个 chunk")

# 2. 初始化 ChromaDB（用自带默认 embedding 模型）
client = chromadb.PersistentClient(path="./chroma_db")
collection = client.get_or_create_collection(name="handbook")

# 3. 准备数据
documents = []
metadatas = []
ids = []

for i, chunk in enumerate(chunks):
    documents.append(chunk["text"])
    metadatas.append({
        "source": chunk["source"],
        "heading": chunk["heading"]
    })
    ids.append(f"{chunk['source']}_{i}")

# 4. 批量存入（一次最多 100 条，分两批）
BATCH_SIZE = 100
for start in range(0, len(documents), BATCH_SIZE):
    end = min(start + BATCH_SIZE, len(documents))
    collection.add(
        documents=documents[start:end],
        metadatas=metadatas[start:end],
        ids=ids[start:end]
    )
    print(f"  ✅ 存入 {start}~{end}")

print(f"\n🎉 共存入 {collection.count()} 条 embedding 到 chroma_db/")
