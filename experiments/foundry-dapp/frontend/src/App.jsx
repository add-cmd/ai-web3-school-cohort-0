import { useState, useCallback, useEffect } from 'react'
import { createPublicClient, createWalletClient, custom, http, parseAbi } from 'viem'
import { sepolia } from 'viem/chains'

// ─── 合约配置 ────────────────────────────────────────────────
const CONTRACTS = {
  counter: {
    address: '0x6d8521408b803813a1A963f511C74fB96ea23bd2',
    name: 'Counter',
    abi: parseAbi([
      'function increment()',
      'function decrement()',
      'function setCount(uint256 newCount)',
      'function count() view returns (uint256)',
    ]),
    functions: [
      { name: 'increment', args: [], label: '➕ Increment', risk: 'low' },
      { name: 'decrement', args: [], label: '➖ Decrement', risk: 'low' },
      { name: 'setCount', args: [{ name: 'newCount', type: 'uint256' }], label: '🔢 Set Count', risk: 'low' },
    ]
  },
  token: {
    address: '0x62E3395eCFa2d18afB8F0cfbB1FA55948Dd03674',
    name: 'SimpleToken',
    abi: parseAbi([
      'function transfer(address to, uint256 amount) returns (bool)',
      'function approve(address spender, uint256 amount)',
      'function balanceOf(address) view returns (uint256)',
    ]),
    functions: [
      { name: 'transfer', args: [
        { name: 'to', type: 'address' },
        { name: 'amount', type: 'uint256' }
      ], label: '💸 Transfer', risk: 'medium' },
      { name: 'approve', args: [
        { name: 'spender', type: 'address' },
        { name: 'amount', type: 'uint256' }
      ], label: '🔑 Approve', risk: 'high' },
    ]
  }
}

const RPC_URL = 'https://ethereum-sepolia-rpc.publicnode.com'
const publicClient = createPublicClient({ chain: sepolia, transport: http(RPC_URL) })

// ─── 颜色 ─────────────────────────────────────────────────────
const RISK_COLORS = {
  low:     { bg: '#dcfce7', text: '#16a34a', label: '🟢 Low' },
  medium:  { bg: '#fef9c3', text: '#ca8a04', label: '🟡 Medium' },
  high:    { bg: '#fee2e2', text: '#dc2626', label: '🔴 High' },
  critical:{ bg: '#fecaca', text: '#991b1b', label: '🚨 Critical' },
}

// ─── 编码函数调用数据 ─────────────────────────────────────────
function encodeCallData(contract, fnName, args) {
  const fn = contract.functions.find(f => f.name === fnName)
  if (!fn) return '0x'
  // 前端简化：只传函数名和参数给后端，后端通过 selector 判断
  return fnName
}

// ─── 主组件 ──────────────────────────────────────────────────
function App() {
  const [account, setAccount] = useState(null)
  const [walletClient, setWalletClient] = useState(null)

  // 操作表单
  const [selectedContract, setSelectedContract] = useState('counter')
  const [selectedFn, setSelectedFn] = useState('increment')
  const [argValues, setArgValues] = useState({})
  const [userIntent, setUserIntent] = useState('')
  const [txHistory, setTxHistory] = useState([])

  // 风险分析结果
  const [analyzing, setAnalyzing] = useState(false)
  const [analysis, setAnalysis] = useState(null)
  const [error, setError] = useState(null)

  const contract = CONTRACTS[selectedContract]
  const currentFn = contract.functions.find(f => f.name === selectedFn)

  // ── 连接钱包 ──────────────────────────────────────────────
  const connectWallet = async () => {
    if (!window.ethereum) { setError('请安装 MetaMask'); return }
    try {
      const wc = createWalletClient({ chain: sepolia, transport: custom(window.ethereum) })
      const [addr] = await wc.requestAddresses()
      setAccount(addr)
      setWalletClient(wc)
    } catch (e) { setError('钱包连接失败: ' + e.message) }
  }

  // ── 第一步：风险分析 ───────────────────────────────────────
  const handleAnalyze = async () => {
    setAnalyzing(true)
    setAnalysis(null)
    setError(null)

    // 编码参数
    let data = selectedFn
    let argsStr = ''
    if (currentFn.args.length > 0) {
      argsStr = currentFn.args.map(a => `${a.name}=${argValues[a.name] || ''}`).join(', ')
    }

    try {
      const resp = await fetch('/api/analyze', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          to: contract.address,
          data: '0x',
          value: '0',
          function: selectedFn,
          args: argsStr,
          user_intent: userIntent || '(未填写)',
        })
      })
      const result = await resp.json()
      if (result.error) { setError(result.error); return }
      setAnalysis(result)
    } catch (e) { setError('分析失败: ' + e.message) }
    setAnalyzing(false)
  }

  // ── 第二步：确认后执行 ─────────────────────────────────────
  const handleExecute = async () => {
    if (!walletClient || !account) { setError('请先连接钱包'); return }
    setError(null)
    try {
      const hash = await walletClient.writeContract({
        address: contract.address,
        abi: contract.abi,
        functionName: selectedFn,
        args: currentFn.args.map(a => {
          const v = argValues[a.name]
          return a.type === 'uint256' ? BigInt(v || '0') : (v || '0x0000000000000000000000000000000000000000')
        }),
        account,
        gas: 100000n,
      })
      setTxHistory(prev => [{ hash, fn: selectedFn, contract: contract.name, status: 'pending' }, ...prev])
      const receipt = await publicClient.waitForTransactionReceipt({ hash })
      setTxHistory(prev => prev.map(t =>
        t.hash === hash ? { ...t, status: receipt.status === 'success' ? 'success' : 'failed' } : t
      ))
      setAnalysis(null)
    } catch (e) { setError(`交易失败: ${e.shortMessage || e.message}`) }
  }

  return (
    <div style={styles.container}>
      <header style={styles.header}>
        <h1 style={styles.title}>🛡 交易风险分析器</h1>
        <p style={styles.subtitle}>Prompt 章节最小实践 · 先分析，再签名</p>
        <div style={{ marginTop: 12 }}>
          {account ? (
            <span style={{ color: '#bbf7d0', fontWeight: 600 }}>
              ✅ {account.slice(0, 6)}...{account.slice(-4)}
            </span>
          ) : (
            <button style={styles.connectBtn} onClick={connectWallet}>连接 MetaMask</button>
          )}
        </div>
      </header>

      <main style={styles.main}>
        {/* ── 操作面板 ── */}
        <section style={styles.card}>
          <h3 style={styles.cardTitle}>📋 构建交易</h3>

          <label style={styles.label}>合约</label>
          <select style={styles.select} value={selectedContract}
            onChange={e => { setSelectedContract(e.target.value); setSelectedFn(CONTRACTS[e.target.value].functions[0].name); setAnalysis(null) }}>
            <option value="counter">Counter (计数器)</option>
            <option value="token">SimpleToken (代币)</option>
          </select>

          <label style={styles.label}>函数</label>
          <select style={styles.select} value={selectedFn}
            onChange={e => { setSelectedFn(e.target.value); setAnalysis(null) }}>
            {contract.functions.map(f => (
              <option key={f.name} value={f.name}>{f.label}</option>
            ))}
          </select>

          {currentFn.args.map(arg => (
            <div key={arg.name}>
              <label style={styles.label}>{arg.name} ({arg.type})</label>
              <input style={styles.input}
                placeholder={arg.type === 'address' ? '0x...' : '数量'}
                value={argValues[arg.name] || ''}
                onChange={e => { setArgValues({ ...argValues, [arg.name]: e.target.value }); setAnalysis(null) }} />
            </div>
          ))}

          <label style={styles.label}>用户意图</label>
          <input style={styles.input}
            placeholder="例：转账给我的朋友Alice"
            value={userIntent}
            onChange={e => { setUserIntent(e.target.value); setAnalysis(null) }} />

          <div style={{ marginTop: 12 }}>
            <button style={styles.analyzeBtn} onClick={handleAnalyze} disabled={analyzing}>
              {analyzing ? '⏳ 分析中...' : '🛡 风险分析'}
            </button>
          </div>
        </section>

        {/* ── 风险摘要卡片 ── */}
        {analysis && (
          <section style={{ ...styles.card, borderLeft: `6px solid ${RISK_COLORS[analysis.risk_level]?.text || '#6366f1'}` }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
              <h3 style={{ margin: 0 }}>🛡 风险摘要</h3>
              <span style={{
                padding: '4px 14px', borderRadius: 20, fontWeight: 700, fontSize: '0.85rem',
                background: RISK_COLORS[analysis.risk_level]?.bg || '#f1f5f9',
                color: RISK_COLORS[analysis.risk_level]?.text || '#475569',
              }}>
                {RISK_COLORS[analysis.risk_level]?.label || analysis.risk_level}
              </span>
            </div>

            <p style={{ fontSize: '0.95rem', color: '#1e293b', marginBottom: 16 }}>{analysis.summary}</p>

            {/* 意图匹配 */}
            <div style={{ ...styles.infoRow, background: analysis.intent_match ? '#f0fdf4' : '#fef2f2' }}>
              <span style={styles.infoLabel}>意图匹配</span>
              <span style={{ fontWeight: 600, color: analysis.intent_match ? '#16a34a' : '#dc2626' }}>
                {analysis.intent_match ? '✅ 匹配' : '❌ 不匹配'}
              </span>
            </div>
            {analysis.intent_note && (
              <p style={{ fontSize: '0.85rem', color: '#64748b', margin: '4px 0 12px' }}>📌 {analysis.intent_note}</p>
            )}

            {/* 资产变化 */}
            {analysis.asset_changes?.length > 0 && (
              <div style={{ marginBottom: 12 }}>
                <span style={{ fontWeight: 600, fontSize: '0.85rem', color: '#475569' }}>💰 资产变化</span>
                {analysis.asset_changes.map((a, i) => (
                  <div key={i} style={{ fontSize: '0.9rem', padding: '4px 0', color: a.direction === 'out' ? '#dc2626' : '#16a34a' }}>
                    {a.direction === 'out' ? '−' : '+'} {a.amount} {a.asset}
                  </div>
                ))}
              </div>
            )}

            {/* 权限变更 */}
            {analysis.permissions_changed?.length > 0 && (
              <div style={{ marginBottom: 12 }}>
                <span style={{ fontWeight: 600, fontSize: '0.85rem', color: '#dc2626' }}>⚠ 权限变更</span>
                {analysis.permissions_changed.map((p, i) => (
                  <div key={i} style={{ fontSize: '0.85rem', padding: '4px 0', color: '#dc2626' }}>
                    {p.detail} — {p.contract?.slice(0, 10)}...
                  </div>
                ))}
              </div>
            )}

            {/* 不确定性 */}
            {analysis.uncertainties?.length > 0 && (
              <div style={{ marginBottom: 12 }}>
                <span style={{ fontWeight: 600, fontSize: '0.85rem', color: '#ca8a04' }}>❓ 不确定性</span>
                {analysis.uncertainties.map((u, i) => (
                  <div key={i} style={{ fontSize: '0.85rem', padding: '2px 0', color: '#64748b' }}>• {u}</div>
                ))}
              </div>
            )}

            {/* 检查清单 */}
            {analysis.recommended_user_checks?.length > 0 && (
              <div style={{ marginBottom: 12 }}>
                <span style={{ fontWeight: 600, fontSize: '0.85rem', color: '#6366f1' }}>🔐 检查清单</span>
                {analysis.recommended_user_checks.map((c, i) => (
                  <div key={i} style={{ fontSize: '0.85rem', padding: '2px 0', color: '#475569' }}>☐ {c}</div>
                ))}
              </div>
            )}

            {/* 确认 / 取消 */}
            <div style={{ display: 'flex', gap: 12, marginTop: 16 }}>
              <button style={{ ...styles.btnExec, opacity: analysis.requires_human_approval && !analysis.intent_match ? 0.5 : 1 }}
                onClick={handleExecute}
                disabled={analysis.requires_human_approval && !analysis.intent_match}>
                {analysis.requires_human_approval ? '⚠ 高风险，谨慎确认' : '✅ 安全，确认发送'}
              </button>
              <button style={{ ...styles.btnCancel }} onClick={() => setAnalysis(null)}>取消</button>
            </div>
          </section>
        )}

        {/* ── 错误 ── */}
        {error && (
          <section style={{ ...styles.card, borderColor: '#ef4444' }}>
            <p style={{ color: '#ef4444' }}>❌ {error}</p>
          </section>
        )}

        {/* ── 交易历史 ── */}
        {txHistory.length > 0 && (
          <section style={styles.card}>
            <h3 style={styles.cardTitle}>📜 交易历史</h3>
            {txHistory.map((tx, i) => (
              <div key={i} style={{ padding: '8px 0', borderBottom: i < txHistory.length - 1 ? '1px solid #f1f5f9' : 'none', fontSize: '0.85rem' }}>
                <code>{tx.hash.slice(0, 16)}...</code>
                <span style={{ margin: '0 8px', color: '#64748b' }}>{tx.contract}.{tx.fn}()</span>
                <span style={{
                  padding: '2px 8px', borderRadius: 10, fontSize: '0.75rem',
                  background: tx.status === 'success' ? '#dcfce7' : tx.status === 'failed' ? '#fee2e2' : '#fef9c3',
                  color: tx.status === 'success' ? '#16a34a' : tx.status === 'failed' ? '#dc2626' : '#ca8a04',
                }}>
                  {tx.status === 'success' ? '✅' : tx.status === 'failed' ? '❌' : '⏳'}
                </span>
              </div>
            ))}
          </section>
        )}

        {/* ── 说明 ── */}
        <section style={{ ...styles.card, background: '#fefce8' }}>
          <h3 style={styles.cardTitle}>📖 Prompt 最小实践 — 三组测试</h3>
          <ol style={{ color: '#475569', fontSize: '0.9rem', lineHeight: 2, paddingLeft: 20, margin: 0 }}>
            <li><strong>Counter.increment</strong> — 普通操作 → 期望 🟢 Low</li>
            <li><strong>Token.approve(恶意地址, ∞)</strong> — 无限授权 → 期望 🚨 Critical</li>
            <li><strong>Token.transfer(恶意地址) + 意图="转给朋友"</strong> — 地址不匹配 → 期望 🔴 High</li>
          </ol>
        </section>
      </main>

      <footer style={styles.footer}>
        <p>AI × Web3 School · Prompt 章节最小实践 · 先分析，再签名</p>
      </footer>
    </div>
  )
}

const styles = {
  container: { minHeight: '100vh', background: '#f8fafc', fontFamily: 'system-ui, -apple-system, sans-serif' },
  header: { background: 'linear-gradient(135deg, #6366f1, #8b5cf6)', color: 'white', padding: '30px 20px', textAlign: 'center' },
  title: { margin: 0, fontSize: '1.8rem' },
  subtitle: { margin: '8px 0 0', opacity: 0.9, fontSize: '0.95rem' },
  connectBtn: { padding: '8px 20px', background: 'rgba(255,255,255,0.2)', color: 'white', border: '1px solid rgba(255,255,255,0.4)', borderRadius: 8, cursor: 'pointer', fontWeight: 600 },
  main: { maxWidth: 700, margin: '0 auto', padding: '20px' },
  card: { background: 'white', borderRadius: 12, padding: 24, marginBottom: 16, border: '1px solid #e2e8f0', boxShadow: '0 1px 3px rgba(0,0,0,0.1)' },
  cardTitle: { margin: '0 0 12px', fontSize: '1.1rem', color: '#1e293b' },
  label: { display: 'block', marginBottom: 4, marginTop: 12, fontWeight: 600, color: '#475569', fontSize: '0.85rem' },
  input: { width: '100%', padding: '10px 14px', border: '1px solid #cbd5e1', borderRadius: 8, fontSize: '0.9rem', fontFamily: 'monospace', boxSizing: 'border-box' },
  select: { width: '100%', padding: '10px 14px', border: '1px solid #cbd5e1', borderRadius: 8, fontSize: '0.9rem', background: 'white', boxSizing: 'border-box' },
  analyzeBtn: { padding: '12px 24px', background: '#6366f1', color: 'white', border: 'none', borderRadius: 8, cursor: 'pointer', fontWeight: 700, fontSize: '1rem', width: '100%' },
  btnExec: { flex: 1, padding: '12px', color: 'white', border: 'none', borderRadius: 8, cursor: 'pointer', fontWeight: 700, fontSize: '0.9rem', background: '#22c55e' },
  btnCancel: { padding: '12px 20px', background: '#e2e8f0', color: '#475569', border: 'none', borderRadius: 8, cursor: 'pointer', fontWeight: 600 },
  infoRow: { display: 'flex', justifyContent: 'space-between', padding: '8px 12px', borderRadius: 8, marginBottom: 8 },
  infoLabel: { color: '#64748b', fontSize: '0.85rem' },
  footer: { textAlign: 'center', padding: 20, color: '#94a3b8', fontSize: '0.85rem' },
}

export default App
