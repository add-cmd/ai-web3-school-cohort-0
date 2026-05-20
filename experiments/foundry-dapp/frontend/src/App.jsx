import { useState, useEffect } from 'react'
import { createPublicClient, http, parseAbi } from 'viem'
import { sepolia } from 'viem/chains'

const CONTRACT_ABI = parseAbi([
  'function count() view returns (uint256)',
  'function owner() view returns (address)',
  'function increment()',
  'function decrement()',
  'function setCount(uint256 newCount)',
  'event CountChanged(uint256 newCount, address triggeredBy)',
])

function App() {
  const [contractAddr, setContractAddr] = useState(import.meta.env.VITE_CONTRACT_ADDRESS || '')
  const [count, setCount] = useState(null)
  const [owner, setOwner] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  const client = createPublicClient({
    chain: sepolia,
    transport: http('https://ethereum-sepolia-rpc.publicnode.com'),
  })

  const fetchData = async () => {
    if (!contractAddr) return
    setLoading(true)
    setError(null)
    try {
      const [countVal, ownerVal] = await Promise.all([
        client.readContract({ address: contractAddr, abi: CONTRACT_ABI, functionName: 'count' }),
        client.readContract({ address: contractAddr, abi: CONTRACT_ABI, functionName: 'owner' }),
      ])
      setCount(Number(countVal))
      setOwner(ownerVal)
    } catch (err) {
      setError(err.message)
    }
    setLoading(false)
  }

  useEffect(() => {
    if (contractAddr) fetchData()
  }, [contractAddr])

  return (
    <div style={styles.container}>
      <header style={styles.header}>
        <h1 style={styles.title}>🔢 Sepolia DApp</h1>
        <p style={styles.subtitle}>Foundry + Go + React</p>
      </header>

      <main style={styles.main}>
        {/* 合约地址输入 */}
        <section style={styles.card}>
          <label style={styles.label}>合约地址</label>
          <div style={styles.row}>
            <input
              style={styles.input}
              value={contractAddr}
              onChange={(e) => setContractAddr(e.target.value)}
              placeholder="0x..."
            />
            <button style={styles.button} onClick={fetchData}>
              {loading ? '加载中...' : '查询'}
            </button>
          </div>
        </section>

        {/* 合约状态 */}
        {count !== null && (
          <section style={styles.card}>
            <div style={styles.counterDisplay}>
              <span style={styles.counterLabel}>当前计数</span>
              <span style={styles.counterValue}>{count}</span>
            </div>
            <div style={styles.infoRow}>
              <span style={styles.infoLabel}>Owner</span>
              <code style={styles.code}>{owner}</code>
            </div>
            <div style={styles.infoRow}>
              <span style={styles.infoLabel}>网络</span>
              <span style={styles.infoValue}>Sepolia</span>
            </div>
          </section>
        )}

        {/* 错误提示 */}
        {error && (
          <section style={{ ...styles.card, borderColor: '#ef4444' }}>
            <p style={{ color: '#ef4444' }}>❌ {error}</p>
          </section>
        )}

        {/* 后端状态 */}
        <section style={styles.card}>
          <h3 style={styles.cardTitle}>📡 Go 后端 API</h3>
          <p style={styles.help}>
            后端运行在 <code>localhost:8080</code>，提供 REST API：
          </p>
          <ul style={styles.list}>
            <li><code>/api/status</code> — 后端状态</li>
            <li><code>/api/contract</code> — 合约数据（需设置 CONTRACT_ADDRESS）</li>
            <li><code>/api/health</code> — 健康检查</li>
          </ul>
        </section>

        {/* 操作提示 */}
        <section style={{ ...styles.card, background: '#fefce8' }}>
          <h3 style={styles.cardTitle}>📋 接下来的步骤</h3>
          <ol style={styles.list}>
            <li>在 Sepolia 测试网获取 ETH（<a href="https://sepoliafaucet.com" target="_blank">水龙头</a>）</li>
            <li>部署合约: <code>forge script script/DeployCounter.s.sol --rpc-url sepolia --broadcast</code></li>
            <li>设置环境变量: <code>export CONTRACT_ADDRESS=0x...</code></li>
            <li>启动后端: <code>cd backend && ./server</code></li>
            <li>启动前端: <code>cd frontend && npm run dev</code></li>
          </ol>
        </section>
      </main>

      <footer style={styles.footer}>
        <p>AI × Web3 School · Counter DApp</p>
      </footer>
    </div>
  )
}

const styles = {
  container: {
    minHeight: '100vh',
    background: '#f8fafc',
    fontFamily: 'system-ui, -apple-system, sans-serif',
  },
  header: {
    background: 'linear-gradient(135deg, #6366f1, #8b5cf6)',
    color: 'white',
    padding: '40px 20px',
    textAlign: 'center',
  },
  title: { margin: 0, fontSize: '2rem' },
  subtitle: { margin: '8px 0 0', opacity: 0.9, fontSize: '1rem' },
  main: { maxWidth: 700, margin: '0 auto', padding: '20px' },
  card: {
    background: 'white',
    borderRadius: 12,
    padding: 24,
    marginBottom: 16,
    border: '1px solid #e2e8f0',
    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
  },
  cardTitle: { margin: '0 0 12px', fontSize: '1.1rem', color: '#1e293b' },
  label: { display: 'block', marginBottom: 8, fontWeight: 600, color: '#475569' },
  row: { display: 'flex', gap: 8 },
  input: {
    flex: 1,
    padding: '10px 14px',
    border: '1px solid #cbd5e1',
    borderRadius: 8,
    fontSize: '0.9rem',
    fontFamily: 'monospace',
  },
  button: {
    padding: '10px 20px',
    background: '#6366f1',
    color: 'white',
    border: 'none',
    borderRadius: 8,
    cursor: 'pointer',
    fontWeight: 600,
  },
  counterDisplay: {
    textAlign: 'center',
    padding: '20px 0',
  },
  counterLabel: {
    display: 'block',
    fontSize: '0.9rem',
    color: '#64748b',
    marginBottom: 8,
  },
  counterValue: {
    fontSize: '3.5rem',
    fontWeight: 700,
    color: '#6366f1',
  },
  infoRow: {
    display: 'flex',
    justifyContent: 'space-between',
    padding: '8px 0',
    borderTop: '1px solid #f1f5f9',
  },
  infoLabel: { color: '#64748b', fontSize: '0.9rem' },
  infoValue: { color: '#1e293b', fontWeight: 500 },
  code: {
    fontSize: '0.8rem',
    background: '#f1f5f9',
    padding: '2px 6px',
    borderRadius: 4,
    maxWidth: 300,
    overflow: 'hidden',
    textOverflow: 'ellipsis',
  },
  help: { color: '#64748b', fontSize: '0.9rem', lineHeight: 1.6 },
  list: { color: '#475569', fontSize: '0.9rem', lineHeight: 2, paddingLeft: 20 },
  footer: {
    textAlign: 'center',
    padding: 20,
    color: '#94a3b8',
    fontSize: '0.85rem',
  },
}

export default App
