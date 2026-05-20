import React, { useState, useEffect } from 'react'
import { createWalletClient, custom, getAddress, erc20Abi } from 'viem'
import { sepolia } from 'viem/chains'

// ── Styles ───────────────────────────────────────────────────────────────────

const styles = {
  container: { maxWidth: '800px', margin: '0 auto', padding: '20px', fontFamily: '-apple-system, BlinkMacSystemFont, sans-serif', background: '#0d1117', color: '#c9d1d9', minHeight: '100vh' },
  header: { borderBottom: '1px solid #30363d', paddingBottom: '16px', marginBottom: '24px' },
  headerTitle: { fontSize: '24px', fontWeight: 700, color: '#58a6ff', margin: 0 },
  headerSub: { fontSize: '14px', color: '#8b949e', marginTop: '4px' },
  card: { background: '#161b22', border: '1px solid #30363d', borderRadius: '8px', padding: '20px', marginBottom: '16px' },
  label: { display: 'block', fontSize: '13px', color: '#8b949e', marginBottom: '6px', fontWeight: 600 },
  input: { width: '100%', padding: '10px 12px', background: '#0d1117', border: '1px solid #30363d', borderRadius: '6px', color: '#c9d1d9', fontSize: '14px', marginBottom: '12px', boxSizing: 'border-box' },
  textarea: { width: '100%', padding: '10px 12px', background: '#0d1117', border: '1px solid #30363d', borderRadius: '6px', color: '#c9d1d9', fontSize: '14px', marginBottom: '12px', boxSizing: 'border-box', minHeight: '60px', resize: 'vertical' },
  btn: { padding: '10px 20px', border: 'none', borderRadius: '6px', fontSize: '14px', fontWeight: 600, cursor: 'pointer', transition: 'all 0.2s' },
  btnPrimary: { background: '#238636', color: '#fff' },
  btnDanger: { background: '#da3633', color: '#fff' },
  btnWallet: { background: '#1f6feb', color: '#fff' },
  btnDisabled: { background: '#21262d', color: '#484f58', cursor: 'not-allowed' },
  tag: { display: 'inline-block', padding: '2px 8px', borderRadius: '12px', fontSize: '11px', fontWeight: 600, marginRight: '6px' },
  tagHigh: { background: '#1a3a1a', color: '#3fb950' },
  tagMedium: { background: '#3d2e00', color: '#d29922' },
  tagLow: { background: '#3d1515', color: '#f85149' },
  riskBadge: { display: 'inline-block', padding: '4px 14px', borderRadius: '20px', fontSize: '14px', fontWeight: 700 },
}

const riskColors = {
  low: { bg: '#1a3a1a', color: '#3fb950' },
  medium: { bg: '#3d2e00', color: '#d29922' },
  high: { bg: '#3d1515', color: '#f85149' },
  critical: { bg: '#5a0f0f', color: '#ff7b72' },
}

const trustConfig = {
  high: { icon: '🟢', label: '高可信 — 链上事实', style: styles.tagHigh },
  medium: { icon: '🟡', label: '中可信 — 须交叉验证', style: styles.tagMedium },
  low: { icon: '🔴', label: '低可信 — 仅作参考', style: styles.tagLow },
}

const freshIcons = { realtime: '⚡', cached: '💾', 'user-input': '👤' }

// ── App ────────────────────────────────────────────────────────────────────

export default function App() {
  const [account, setAccount] = useState(null)
  const [walletClient, setWalletClient] = useState(null)
  const [loading, setLoading] = useState(false)
  const [executing, setExecuting] = useState(false)

  // Form — Sepolia 测试网预填值
  const [tokenAddr, setTokenAddr] = useState('0x62E3395eCFa2d18afB8F0cfbB1FA55948Dd03674')
  const [spenderAddr, setSpenderAddr] = useState('0x1f9840a85d5af5bf1d1762f925bdaddc4201f984')
  const [amount, setAmount] = useState('1000000')
  const [intent, setIntent] = useState('我要在 Uniswap 上 swap STK 代币')

  // Result
  const [result, setResult] = useState(null)
  const [error, setError] = useState(null)

  // Connect Wallet
  const connectWallet = async () => {
    if (typeof window.ethereum === 'undefined') {
      setError('请安装 MetaMask')
      return
    }
    try {
      const client = createWalletClient({
        chain: sepolia,
        transport: custom(window.ethereum),
      })
      const [address] = await client.requestAddresses()
      setAccount(address)
      setWalletClient(client)
      setError(null)
    } catch (e) {
      setError('钱包连接失败: ' + e.message)
    }
  }

  // Analyze
  const handleAnalyze = async () => {
    setLoading(true)
    setError(null)
    setResult(null)
    try {
      const resp = await fetch('/api/analyze', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          tokenAddress: tokenAddr,
          spenderAddress: spenderAddr,
          amount,
          userIntent: intent,
          userAddress: account || '0x0000000000000000000000000000000000000001',
        }),
      })
      const data = await resp.json()
      if (data.error) {
        setError(data.error)
      } else {
        setResult(data)
      }
    } catch (e) {
      setError('分析请求失败: ' + e.message)
    }
    setLoading(false)
  }

  // Execute Approve
  const handleApprove = async () => {
    if (!walletClient || !account) {
      setError('请先连接钱包')
      return
    }
    setExecuting(true)
    setError(null)
    try {
      const hash = await walletClient.writeContract({
        address: getAddress(tokenAddr),
        abi: erc20Abi,
        functionName: 'approve',
        args: [getAddress(spenderAddr), BigInt(amount)],
        account,
        chain: sepolia,
      })
      setError(`✅ 交易已发送! Hash: ${hash}`)
    } catch (e) {
      setError('交易执行失败: ' + e.message)
    }
    setExecuting(false)
  }

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <h1 style={styles.headerTitle}>🔍 钱包授权检查 Agent</h1>
        <p style={styles.headerSub}>
          Context Engineering 最小实践 — 先装配上下文，再判断风险
        </p>
      </div>

      {/* Wallet */}
      <div style={styles.card}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{...styles.label, margin: 0}}>钱包</span>
          {account ? (
            <div>
              <span style={{ fontSize: '13px', color: '#3fb950' }}>✅ {account.slice(0,6)}...{account.slice(-4)}</span>
            </div>
          ) : (
            <button style={{...styles.btn, ...styles.btnWallet}} onClick={connectWallet}>
              连接 MetaMask
            </button>
          )}
        </div>
      </div>

      {/* Input Form */}
      <div style={styles.card}>
        <h3 style={{ margin: '0 0 16px', color: '#58a6ff', fontSize: '16px' }}>
          📝 交易参数 {account ? <span style={{fontSize:'12px', color:'#8b949e', fontWeight:400}}>— 填写后点分析</span> : ''}
        </h3>

        <label style={styles.label}>Token 合约地址</label>
        <input style={styles.input} value={tokenAddr} onChange={e => setTokenAddr(e.target.value)} placeholder="0x..." />

        <label style={styles.label}>Spender 地址 (接收授权的合约)</label>
        <input style={styles.input} value={spenderAddr} onChange={e => setSpenderAddr(e.target.value)} placeholder="0x..." />

        <label style={styles.label}>授权金额</label>
        <input style={styles.input} value={amount} onChange={e => setAmount(e.target.value)} placeholder="1000000" />

        <label style={styles.label}>你的意图 (dApp 页面告诉你的)</label>
        <textarea style={styles.textarea} value={intent} onChange={e => setIntent(e.target.value)} placeholder="我想..." />

        <button
          style={{...styles.btn, ...(loading ? styles.btnDisabled : styles.btnPrimary)}}
          onClick={handleAnalyze}
          disabled={loading}
        >
          {loading ? '⏳ 分析中...' : '🔍 分析风险'}
        </button>
      </div>

      {/* Result */}
      {error && (
        <div style={{...styles.card, borderColor: '#da3633'}}>
          <p style={{ color: '#ff7b72', margin: 0, fontSize: '14px' }}>{error}</p>
        </div>
      )}

      {result && (
        <>
          {/* Risk Level */}
          <div style={styles.card}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px', marginBottom: '12px' }}>
              <span style={{...styles.riskBadge, background: riskColors[result.riskLevel]?.bg || '#21262d', color: riskColors[result.riskLevel]?.color || '#c9d1d9', textTransform: 'uppercase'}}>
                {result.riskLevel || 'UNKNOWN'}
              </span>
              <span style={{ fontSize: '16px', fontWeight: 600 }}>{result.summary}</span>
            </div>

            {result.recommendation && (
              <div style={{ background: '#0d1117', padding: '12px', borderRadius: '6px', fontSize: '14px', lineHeight: 1.6, color: '#8b949e' }}>
                💡 <strong style={{color: '#c9d1d9'}}>建议：</strong>{result.recommendation}
              </div>
            )}
          </div>

          {/* Context Blocks */}
          <div style={styles.card}>
            <h3 style={{ margin: '0 0 16px', color: '#58a6ff', fontSize: '16px' }}>
              📦 上下文装配 (Context Assembly)
            </h3>
            {result.contextBlocks?.map((block, i) => {
              const tc = trustConfig[block.trust] || { icon: '❓', label: block.trust, style: { background: '#21262d', color: '#8b949e' } }
              return (
                <div key={i} style={{ background: '#0d1117', border: '1px solid #21262d', borderRadius: '6px', padding: '14px', marginBottom: '10px' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px', flexWrap: 'wrap' }}>
                    <span style={{ fontWeight: 700, fontSize: '14px' }}>{tc.icon} Block {i+1}: {block.label}</span>
                    <span style={{...styles.tag, ...tc.style}}>{block.trust.toUpperCase()}</span>
                    <span style={{ fontSize: '12px', color: '#8b949e' }}>{freshIcons[block.freshness]} {block.freshness}</span>
                    <span style={{ fontSize: '11px', color: '#484f58' }}>source: {block.source}</span>
                  </div>
                  {block.warning && (
                    <div style={{ background: '#3d2e00', padding: '6px 10px', borderRadius: '4px', fontSize: '12px', color: '#d29922', marginBottom: '8px' }}>
                      ⚠️ {block.warning}
                    </div>
                  )}
                  <pre style={{ fontSize: '12px', lineHeight: 1.5, color: '#8b949e', margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                    {JSON.stringify(block.data, null, 2)}
                  </pre>
                </div>
              )
            })}
          </div>

          {/* Analysis */}
          {result.analysis && (
            <div style={styles.card}>
              <h3 style={{ margin: '0 0 16px', color: '#58a6ff', fontSize: '16px' }}>
                🤖 LLM 分析
              </h3>
              <pre style={{ fontSize: '13px', lineHeight: 1.7, color: '#c9d1d9', margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                {JSON.stringify(JSON.parse(typeof result.analysis === 'string' ? result.analysis : JSON.stringify(result.analysis)), null, 2)}
              </pre>
            </div>
          )}

          {/* Execute Button */}
          {account && (
            <div style={styles.card}>
              <h3 style={{ margin: '0 0 16px', color: '#58a6ff', fontSize: '16px' }}>
                🚀 执行操作
              </h3>
              <p style={{ fontSize: '13px', color: '#8b949e', marginBottom: '12px' }}>
                分析完成后，可以选择执行 approve 操作：
              </p>
              <button
                style={{...styles.btn, ...(executing ? styles.btnDisabled : (result.requiresHumanApproval !== false ? styles.btnDanger : styles.btnPrimary))}}
                onClick={handleApprove}
                disabled={executing}
              >
                {executing ? '⏳ 交易处理中...' : (result.requiresHumanApproval !== false ? '⚠️ 高风险 — 谨慎执行' : '✅ 执行 Approve')}
              </button>
            </div>
          )}
        </>
      )}

      {/* Footer */}
      <div style={{ textAlign: 'center', fontSize: '12px', color: '#484f58', padding: '20px' }}>
        Day 3 — Context 最小实践 · Context Engineering · {new Date().getFullYear()}
      </div>
    </div>
  )
}
