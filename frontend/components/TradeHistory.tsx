'use client'

import { useEffect, useState } from 'react'
import { getTrades, Trade } from '@/lib/api'

const TOKEN_SYMBOLS: Record<string, string> = {
  So11111111111111111111111111111111111111112:        'SOL',
  EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v: 'USDC',
}

function sym(address: string) {
  return TOKEN_SYMBOLS[address] ?? address.slice(0, 4) + '…'
}

function StatusLED({ status }: { status: string }) {
  const map: Record<string, { dot: string; label: string; shadow: string }> = {
    unsigned:  { dot: 'bg-[var(--text-faint)]', label: 'Unsigned', shadow: 'none' },
    pending:   { dot: 'bg-yellow-400 animate-led-yellow', label: 'Pending',   shadow: '0 0 5px 1px rgba(250,204,21,0.8)' },
    confirmed: { dot: 'bg-green-400  animate-led-green',  label: 'Confirmed', shadow: '0 0 5px 1px rgba(34,197,94,0.8)'  },
    failed:    { dot: 'bg-[var(--accent)] animate-led-red', label: 'Failed',  shadow: '0 0 5px 1px rgba(255,71,87,0.8)'  },
  }
  const c = map[status] ?? map.unsigned
  return (
    <div className="flex items-center gap-1.5">
      <div className={`w-2 h-2 rounded-full shrink-0 ${c.dot}`} style={{ boxShadow: c.shadow }} />
      <span className="font-mono text-[10px] uppercase tracking-[0.08em] text-[var(--text-muted)] whitespace-nowrap">
        {c.label}
      </span>
    </div>
  )
}

export default function TradeHistory({ token }: { token: string }) {
  const [trades, setTrades]   = useState<Trade[]>([])
  const [loading, setLoading] = useState(true)

  async function load() {
    const data = await getTrades(token)
    setTrades(data.trades ?? [])
    setLoading(false)
  }

  useEffect(() => {
    load()
    const id = setInterval(load, 10_000)
    return () => clearInterval(id)
  }, [token])

  if (loading) {
    return (
      <div className="font-mono text-[11px] uppercase tracking-[0.1em] text-[var(--text-muted)] py-16 text-center">
        Loading…
      </div>
    )
  }

  if (!trades.length) {
    return (
      <div
        className="relative rounded-2xl p-14 text-center w-full max-w-2xl screws"
        style={{ background: 'var(--panel)', boxShadow: 'var(--shadow-card)' }}
      >
        <div className="font-sans font-semibold text-lg text-[var(--text)] mb-2">No trades yet</div>
        <div className="font-mono text-[10px] uppercase tracking-[0.1em] text-[var(--text-muted)]">
          Execute a swap to see your history
        </div>
      </div>
    )
  }

  return (
    <div className="w-full max-w-2xl">
      <div
        className="relative rounded-2xl overflow-hidden screws"
        style={{ background: 'var(--panel)', boxShadow: 'var(--shadow-card)' }}
      >
        {/* Header */}
        <div
          className="px-6 py-3 flex items-center justify-between border-b"
          style={{ borderColor: 'var(--border)', background: 'var(--recessed)' }}
        >
          <span className="font-mono text-[10px] uppercase tracking-[0.12em] text-[var(--text-muted)]">
            // Trade Log
          </span>
          <span className="font-mono text-[10px] uppercase tracking-[0.1em] text-[var(--text-faint)]">
            {trades.length} records
          </span>
        </div>

        {/* Rows */}
        <div>
          {trades.map((t, i) => (
            <div
              key={t.id}
              className={[
                'px-6 py-4 flex items-start justify-between gap-4',
                i < trades.length - 1 ? 'border-b' : '',
              ].join(' ')}
              style={i < trades.length - 1 ? { borderColor: 'var(--border)' } : {}}
            >
              {/* Amount + meta */}
              <div className="min-w-0 flex-1">
                <div className="flex items-baseline gap-2 flex-wrap">
                  <span className="font-sans font-semibold text-base text-[var(--text)]">
                    {t.amountIn} {sym(t.fromToken)}
                  </span>
                  <span className="font-mono text-xs text-[var(--accent)]">→</span>
                  <span className="font-sans font-semibold text-base text-[var(--text)]">
                    {t.amountOut.toFixed(4)} {sym(t.toToken)}
                  </span>
                </div>

                <div className="flex items-center gap-3 mt-1.5 flex-wrap">
                  <span className="font-mono text-[10px] uppercase tracking-[0.08em] text-[var(--text-faint)]">
                    {t.protocol}
                  </span>
                  <span className="font-mono text-[10px] text-[var(--text-faint)]">
                    {new Date(t.createdAt).toLocaleString()}
                  </span>
                </div>

                {t.txHash && (
                  <a
                    href={`https://solscan.io/tx/${t.txHash}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="font-mono text-[10px] text-[var(--text-faint)] hover:text-[var(--text-muted)] mt-1 block truncate transition-colors"
                  >
                    {t.txHash.slice(0, 16)}…
                  </a>
                )}
              </div>

              {/* Status LED */}
              <div className="mt-1 shrink-0">
                <StatusLED status={t.status} />
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
