'use client'

import { useState, useEffect } from 'react'
import { useWallet, useConnection } from '@solana/wallet-adapter-react'
import { VersionedTransaction, PublicKey } from '@solana/web3.js'
import { getQuote, executeSwap, confirmTrade, Quote } from '@/lib/api'
import ConfirmModal from './ConfirmModal'

const TOKENS = [
  { symbol: 'SOL',  address: 'So11111111111111111111111111111111111111112' },
  { symbol: 'USDC', address: 'EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v' },
]

const SLIPPAGE_PRESETS = [10, 50, 100]

type Status = 'idle' | 'quoting' | 'quoted' | 'swapping' | 'pending' | 'confirmed' | 'failed'

/* ── Micro-components ─────────────────────────────────────────────────── */

function VentSlots() {
  return (
    <div className="flex gap-1">
      {[0, 1, 2].map(i => (
        <div key={i} className="h-5 w-px rounded-full bg-[var(--recessed)]"
          style={{ boxShadow: 'inset 1px 1px 2px #0c0f15' }} />
      ))}
    </div>
  )
}

function LED({ color = 'red' }: { color?: 'red' | 'green' | 'yellow' }) {
  const map = {
    red:    { cls: 'bg-[var(--accent)] animate-led-red',    s: '0 0 6px 2px rgba(255,71,87,0.8)'  },
    green:  { cls: 'bg-green-400 animate-led-green',        s: '0 0 6px 2px rgba(34,197,94,0.8)' },
    yellow: { cls: 'bg-yellow-400 animate-led-yellow',      s: '0 0 6px 2px rgba(250,204,21,0.8)'},
  }
  const c = map[color]
  return <div className={`w-2 h-2 rounded-full shrink-0 ${c.cls}`} style={{ boxShadow: c.s }} />
}

/* ── Component ────────────────────────────────────────────────────────── */

export default function SwapForm({ token }: { token: string }) {
  const { sendTransaction, publicKey } = useWallet()
  const { connection } = useConnection()

  const [fromIdx, setFromIdx]               = useState(0)
  const [toIdx, setToIdx]                   = useState(1)
  const [amount, setAmount]                 = useState('')
  const [quote, setQuote]                   = useState<Quote | null>(null)
  const [status, setStatus]                 = useState<Status>('idle')
  const [txHash, setTxHash]                 = useState('')
  const [error, setError]                   = useState('')
  const [slippage, setSlippage]             = useState(50)
  const [showSettings, setShowSettings]     = useState(false)
  const [customSlippage, setCustomSlippage] = useState('')
  const [secondsLeft, setSecondsLeft]       = useState<number | null>(null)
  const [showConfirm, setShowConfirm]       = useState(false)
  const [solBalance, setSolBalance]         = useState<number | null>(null)
  const [usdcBalance, setUsdcBalance]       = useState<number | null>(null)

  const from = TOKENS[fromIdx]
  const to   = TOKENS[toIdx]

  async function loadBalances() {
    if (!publicKey) return
    try {
      const sol = await connection.getBalance(publicKey)
      setSolBalance(sol / 1e9)
      const accounts = await connection.getParsedTokenAccountsByOwner(publicKey, {
        mint: new PublicKey('EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v'),
      })
      const usdc = accounts.value[0]?.account.data.parsed.info.tokenAmount.uiAmount ?? 0
      setUsdcBalance(usdc)
    } catch {}
  }

  useEffect(() => { loadBalances() }, [publicKey])

  useEffect(() => {
    if (!quote?.expiresAt) { setSecondsLeft(null); return }
    const tick = () => {
      const left = Math.floor((new Date(quote.expiresAt).getTime() - Date.now()) / 1000)
      setSecondsLeft(left <= 0 ? 0 : left)
    }
    tick()
    const id = setInterval(tick, 1000)
    return () => clearInterval(id)
  }, [quote?.expiresAt])

  const quoteExpired = secondsLeft !== null && secondsLeft <= 0

  async function handleGetQuote() {
    if (!amount || isNaN(Number(amount))) return
    setStatus('quoting'); setError(''); setQuote(null); setSecondsLeft(null)
    try {
      const q = await getQuote(token, { fromToken: from.address, toToken: to.address, amount: Number(amount), chain: 'solana' })
      if (q.code) throw new Error(q.message ?? q.code)
      setQuote(q); setStatus('quoted')
    } catch (e: any) { setError(e.message); setStatus('idle') }
  }

  async function handleSwap() {
    if (!quote) return
    setShowConfirm(false); setStatus('swapping'); setError('')
    try {
      const result = await executeSwap(token, quote.quoteId, slippage)
      if (result.code) throw new Error(result.message ?? result.code)
      const tx  = VersionedTransaction.deserialize(Buffer.from(result.unsignedTx, 'base64'))
      const sig = await sendTransaction(tx, connection)
      setTxHash(sig)
      await confirmTrade(token, result.tradeId, sig)
      setStatus('pending')
      loadBalances()
    } catch (e: any) { setError(e.message); setStatus('quoted') }
  }

  function flip() { setFromIdx(toIdx); setToIdx(fromIdx); setQuote(null); setStatus('idle'); setSecondsLeft(null) }
  function reset() { setStatus('idle'); setQuote(null); setAmount(''); setTxHash(''); setError(''); setSecondsLeft(null) }

  function balanceFor(idx: number) {
    if (TOKENS[idx].symbol === 'SOL')  return solBalance  !== null ? `${solBalance.toFixed(4)} SOL`   : null
    if (TOKENS[idx].symbol === 'USDC') return usdcBalance !== null ? `${usdcBalance.toFixed(2)} USDC` : null
    return null
  }

  function applySlippage(bps: number) { setSlippage(bps); setCustomSlippage(''); setQuote(null); setStatus('idle') }
  function applyCustom(val: string) {
    setCustomSlippage(val)
    const bps = Math.round(parseFloat(val) * 100)
    if (!isNaN(bps) && bps > 0) { setSlippage(bps); setQuote(null); setStatus('idle') }
  }

  /* LED color based on status */
  const ledColor: 'red' | 'green' | 'yellow' =
    status === 'confirmed' ? 'green'
    : status === 'pending' || status === 'swapping' ? 'yellow'
    : 'red'

  return (
    <>
      {showConfirm && quote && (
        <ConfirmModal
          quote={quote}
          amount={amount}
          slippage={slippage}
          onConfirm={handleSwap}
          onCancel={() => setShowConfirm(false)}
        />
      )}

      <div className="w-full max-w-md">
        <div
          className="relative rounded-2xl p-6 screws"
          style={{ background: 'var(--panel)', boxShadow: 'var(--shadow-card)' }}
        >

          {/* ── Card header ────────────────────────────────────────── */}
          <div className="flex items-center justify-between mb-6">
            <div className="flex items-center gap-3">
              <LED color={ledColor} />
              <span className="font-mono text-[11px] uppercase tracking-[0.1em] text-[var(--text-muted)]">
                // Swap
              </span>
            </div>
            <div className="flex items-center gap-4">
              <VentSlots />
              <button
                onClick={() => setShowSettings(s => !s)}
                title="Slippage settings"
                className={[
                  'font-mono text-sm leading-none transition-colors duration-200',
                  showSettings ? 'text-[var(--accent)]' : 'text-[var(--text-faint)] hover:text-[var(--text-muted)]',
                ].join(' ')}
              >
                ⚙
              </button>
            </div>
          </div>

          {/* ── Slippage panel ─────────────────────────────────────── */}
          {showSettings && (
            <div
              className="rounded-xl p-4 mb-5"
              style={{ background: 'var(--recessed)', boxShadow: 'var(--shadow-recessed)' }}
            >
              <div className="font-mono text-[10px] uppercase tracking-[0.1em] text-[var(--text-muted)] mb-3">
                Slippage tolerance
              </div>
              <div className="flex gap-2 items-center">
                {SLIPPAGE_PRESETS.map(bps => (
                  <button
                    key={bps}
                    onClick={() => applySlippage(bps)}
                    className={[
                      'px-3 py-1.5 rounded-lg font-mono text-xs uppercase tracking-[0.06em] transition-all duration-150',
                      slippage === bps && !customSlippage
                        ? 'bg-[var(--accent)] text-white shadow-accent'
                        : 'text-[var(--text-muted)] hover:text-[var(--text)]',
                    ].join(' ')}
                    style={
                      slippage === bps && !customSlippage
                        ? {}
                        : { background: 'var(--panel)', boxShadow: 'var(--shadow-card)' }
                    }
                  >
                    {bps / 100}%
                  </button>
                ))}
                <div className="relative flex-1">
                  <input
                    type="number"
                    value={customSlippage}
                    onChange={e => applyCustom(e.target.value)}
                    placeholder="Custom"
                    className="w-full bg-transparent rounded-lg font-mono text-xs text-[var(--text)] py-1.5 px-3 outline-none placeholder-[var(--text-faint)]"
                    style={{ boxShadow: 'var(--shadow-recessed)', background: 'var(--chassis)' }}
                  />
                  {customSlippage && (
                    <span className="absolute right-3 top-1.5 font-mono text-xs text-[var(--text-muted)]">%</span>
                  )}
                </div>
              </div>
            </div>
          )}

          {/* ── You pay ────────────────────────────────────────────── */}
          <div
            className="rounded-xl p-4 mb-2"
            style={{ background: 'var(--recessed)', boxShadow: 'var(--shadow-recessed)' }}
          >
            <div className="flex justify-between items-center mb-3">
              <span className="font-mono text-[10px] uppercase tracking-[0.1em] text-[var(--text-muted)]">
                Sell
              </span>
              <select
                value={fromIdx}
                onChange={e => { setFromIdx(Number(e.target.value)); setQuote(null); setStatus('idle') }}
                className="bg-transparent font-mono text-xs uppercase tracking-[0.06em] text-[var(--text-muted)] outline-none cursor-pointer hover:text-[var(--text)] transition-colors"
              >
                {TOKENS.map((t, i) => (
                  <option key={t.address} value={i} className="bg-[#1c2028]">{t.symbol}</option>
                ))}
              </select>
            </div>
            <input
              type="number"
              value={amount}
              onChange={e => { setAmount(e.target.value); setQuote(null); setStatus('idle') }}
              placeholder="0.0"
              className="w-full bg-transparent font-sans font-bold text-3xl text-[var(--text)] outline-none placeholder-[var(--text-faint)]/40 tracking-tight"
            />
            {balanceFor(fromIdx) !== null && (
              <div className="font-mono text-[10px] uppercase tracking-[0.08em] text-[var(--text-faint)] mt-2">
                Bal: {balanceFor(fromIdx)}
              </div>
            )}
          </div>

          {/* ── Flip button ────────────────────────────────────────── */}
          <div className="flex justify-center items-center my-2">
            <button
              onClick={flip}
              className="w-9 h-9 rounded-full flex items-center justify-center text-[var(--text-muted)] hover:text-[var(--accent)] transition-all duration-150 active:translate-y-[1px] active:shadow-pressed"
              style={{ background: 'var(--chassis)', boxShadow: 'var(--shadow-card)' }}
              aria-label="Flip tokens"
            >
              <span className="text-sm leading-none">⇅</span>
            </button>
          </div>

          {/* ── You receive ────────────────────────────────────────── */}
          <div
            className="rounded-xl p-4 mb-5"
            style={{ background: 'var(--recessed)', boxShadow: 'var(--shadow-recessed)' }}
          >
            <div className="flex justify-between items-center mb-3">
              <span className="font-mono text-[10px] uppercase tracking-[0.1em] text-[var(--text-muted)]">
                Buy
              </span>
              <select
                value={toIdx}
                onChange={e => { setToIdx(Number(e.target.value)); setQuote(null); setStatus('idle') }}
                className="bg-transparent font-mono text-xs uppercase tracking-[0.06em] text-[var(--text-muted)] outline-none cursor-pointer hover:text-[var(--text)] transition-colors"
              >
                {TOKENS.map((t, i) => (
                  <option key={t.address} value={i} className="bg-[#1c2028]">{t.symbol}</option>
                ))}
              </select>
            </div>
            <div className="font-sans font-bold text-3xl text-[var(--text)] tracking-tight">
              {quote
                ? quote.amountOut.toFixed(6)
                : <span className="text-[var(--text-faint)]/40">—</span>}
            </div>
            {balanceFor(toIdx) !== null && (
              <div className="font-mono text-[10px] uppercase tracking-[0.08em] text-[var(--text-faint)] mt-2">
                Bal: {balanceFor(toIdx)}
              </div>
            )}
          </div>

          {/* ── Quote info + timer ─────────────────────────────────── */}
          {quote && (
            <div className="flex justify-between items-center mb-5 px-0.5">
              <span className="font-mono text-[10px] uppercase tracking-[0.06em] text-[var(--text-muted)]">
                Via <span className="text-[var(--accent)]">{quote.protocol}</span>
                <span className="ml-3 text-[var(--text-faint)]">
                  1 {from.symbol} = {(quote.amountOut / quote.amountIn).toFixed(4)} {to.symbol}
                </span>
              </span>
              {secondsLeft !== null && (
                <span className={`font-mono text-[10px] uppercase tracking-[0.06em] ${secondsLeft <= 10 ? 'text-[var(--accent)]' : 'text-[var(--text-muted)]'}`}>
                  {quoteExpired ? 'Expired' : `${secondsLeft}s`}
                </span>
              )}
            </div>
          )}

          {/* ── Status messages ────────────────────────────────────── */}
          {error && (
            <div className="font-mono text-[10px] uppercase tracking-[0.06em] text-[var(--accent)] mb-4 px-0.5">
              ✕ {error}
            </div>
          )}
          {quoteExpired && status === 'quoted' && (
            <div className="font-mono text-[10px] uppercase tracking-[0.06em] text-[var(--accent)] mb-4 px-0.5">
              Quote expired — refresh to continue
            </div>
          )}
          {(status === 'pending' || status === 'confirmed') && (
            <div className="mb-4 px-0.5">
              <div className={`font-mono text-[10px] uppercase tracking-[0.06em] ${status === 'confirmed' ? 'text-green-400' : 'text-yellow-400'}`}>
                {status === 'pending' ? '⏳ Awaiting confirmation…' : '✓ Confirmed on-chain'}
              </div>
              {txHash && (
                <a
                  href={`https://solscan.io/tx/${txHash}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="font-mono text-[10px] text-[var(--text-faint)] hover:text-[var(--text-muted)] mt-1 block truncate transition-colors"
                >
                  {txHash.slice(0, 22)}…
                </a>
              )}
            </div>
          )}

          {/* ── Action buttons ─────────────────────────────────────── */}
          {(status === 'idle' || status === 'quoting') && (
            <button
              onClick={handleGetQuote}
              disabled={!amount || status === 'quoting'}
              className="w-full py-4 rounded-xl bg-[var(--accent)] text-white font-mono text-xs uppercase tracking-[0.1em] font-semibold shadow-accent hover:brightness-110 active:translate-y-[2px] active:shadow-accent-pressed disabled:opacity-40 disabled:cursor-not-allowed transition-all duration-150"
            >
              {status === 'quoting' ? 'Fetching quote…' : 'Get Quote ▶'}
            </button>
          )}
          {status === 'quoted' && !quoteExpired && (
            <button
              onClick={() => setShowConfirm(true)}
              className="w-full py-4 rounded-xl bg-[var(--accent)] text-white font-mono text-xs uppercase tracking-[0.1em] font-semibold shadow-accent hover:brightness-110 active:translate-y-[2px] active:shadow-accent-pressed transition-all duration-150"
            >
              Swap ▶
            </button>
          )}
          {status === 'quoted' && quoteExpired && (
            <button
              onClick={handleGetQuote}
              className="w-full py-4 rounded-xl font-mono text-xs uppercase tracking-[0.1em] text-[var(--text-muted)] hover:text-[var(--text)] shadow-card active:shadow-pressed active:translate-y-[2px] transition-all duration-150"
              style={{ background: 'var(--chassis)' }}
            >
              Refresh Quote
            </button>
          )}
          {status === 'swapping' && (
            <button
              disabled
              className="w-full py-4 rounded-xl font-mono text-xs uppercase tracking-[0.1em] text-[var(--text-faint)] cursor-not-allowed"
              style={{ background: 'var(--chassis)', boxShadow: 'var(--shadow-recessed)' }}
            >
              Confirm in Phantom…
            </button>
          )}
          {(status === 'pending' || status === 'confirmed' || status === 'failed') && (
            <button
              onClick={reset}
              className="w-full py-4 rounded-xl font-mono text-xs uppercase tracking-[0.1em] text-[var(--text-muted)] hover:text-[var(--text)] shadow-card active:shadow-pressed active:translate-y-[2px] transition-all duration-150"
              style={{ background: 'var(--chassis)' }}
            >
              New Swap
            </button>
          )}
        </div>
      </div>
    </>
  )
}
