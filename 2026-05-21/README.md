# Day 4 — RAG 最小实践：Handbook Q&A 系统

> **日期：** 2026-05-21
> **章节：** RAG（检索增强生成）
> **核心：** 用关键词检索 + DeepSeek API 搭建轻量 RAG 问答系统
> **技术栈：** Python + ChromaDB + DeepSeek API
> **数据源：** AI × Web3 School Handbook（21 篇文档）

---

## 项目：Handbook Q&A 系统

**先检索，再回答。** 从 Handbook 抓取全部 21 篇文档，构建可检索的知识库，用关键词匹配替代向量检索，轻量级实现 RAG 全流程。

### 流水线

```
fetch_pages.py                 chunk_docs.py                build_vector_db.py
     ↓                              ↓                              ↓
抓取 21 篇 Handbook 页面      分割为 200-500 字 chunk        构建 ChromaDB 向量索引
(pages/html/, pages/text/)    (pages/chunks.json)            (chroma_db/)

                                                                     ↓
                                                               rag_qa.py
                                                                     ↓
                                                     交互式问答：检索 → LLM → 回答
```

### 各模块说明

| 文件 | 功能 |
|:---|:---|
| `fetch_pages.py` | 抓取 aiweb3.school 的 21 篇中文文档（AI 基础 11 篇 + Web3 基础 10 篇），提取正文保存为 HTML 和纯文本 |
| `chunk_docs.py` | 将纯文本按章节/段落分割为 200-500 字的 chunk，输出 JSON |
| `build_vector_db.py` | 将 chunk 存入 ChromaDB（持久化向量数据库） |
| `rag_qa.py` | 交互式问答终端：输入问题 → 关键词检索 → DeepSeek API 回答 → 显示来源 |

### RAG 实现要点

- **检索方式：** 关键词匹配替代向量检索（无需 embedding 模型，零依赖部署）
- **上下文组装：** 每个 chunk 标注 `[来源]` 和标题，LLM 回答时引用
- **回答约束：** 只基于文档内容回答，不确定时明确说明
- **输出格式：** 结构化 JSON（回答 + 来源 + 不确定性 + 时效性提示）

### 学习笔记

**RAG vs 纯 Prompt：**
- 纯 Prompt：知识依赖模型训练数据，无法获取最新/特定文档内容
- RAG：检索外部知识库再回答，可控制信息来源

**关键词检索 vs 向量检索：**
- 关键词：简单、零依赖，但无法理解语义（搜索"智能合约"搜不到"solidity"）
- 向量：语义理解好，但需要 embedding 模型和向量数据库
