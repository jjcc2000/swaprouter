'use client'

import { Quote } from '@/lib/api'

const TOKEN_SYMBOLS: Record<string, string> = {
  So11111111111111111111111111111111111111112:        'SOL',
  EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v: 'USDC',
}

function sym(address: string) {
  return TOKEN_SYMBOLS[address] ?? address.slice(0, 6) + '…'
}

function DataRow({ label, value, accent }: { label: string; value: string; accent?: boolean }) {
  return (
    <div
      className="flex justify-between items-baseline py-2.5 border-b"
      style={{ borderColor: 'var(--border)' }}
    >
      <span className="font-mono text-[10px] uppercase tracking-[0.1em] text-[var(--text-muted)]">
        {label}
      </span>
      <span className={`font-mono text-xs ${accent ? 'text-[var(--accent)]' : 'text-[var(--text)]'}`}>
        {value}
      </span>
    </div>
  )
}

interface Props {
  quote:     Quote
  amount:    string
  slippage:  number
  onConfirm: () => void
  onCancel:  () => void
}

export default function ConfirmModal({ quote, amount, slippage, onConfirm, onCancel }: Props) {
  const rate    = (quote.amountOut / quote.amountIn).toFixed(6)
  const fromSym = sym(quote.fromToken)
  const toSym   = sym(quote.toToken)

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center px-4"
      style={{ background: 'rgba(10,12,16,0.88)' }}
    >
      <div
        className="relative w-full max-w-sm rounded-2xl p-7 flex flex-col gap-5 screws"
        style={{ background: 'var(--panel)', boxShadow: 'var(--shadow-floating)' }}
      >
        {/* Header */}
        <div className="flex items-center justify-between">
          <span className="font-mono text-[10px] uppercase tracking-[0.12em] text-[var(--text-muted)]">
            // Confirm Swap
          </span>
          <div className="w-2 h-2 rounded-full bg-yellow-400 animate-led-yellow"
            style={{ boxShadow: '0 0 6px 2px rgba(250,204,21,0.8)' }} />
        </div>

        {/* Amount blocks */}
        <div className="flex flex-col gap-2">
          <div
            className="rounded-xl p-4"
            style={{ background: 'var(--recessed)', boxShadow: 'var(--shadow-recessed)' }}
          >
            <div className="font-mono text-[10px] uppercase tracking-[0.1em] text-[var(--text-muted)] mb-2">
              You sell
            </div>
            <div className="font-sans font-bold text-2xl text-[var(--text)] tracking-tight">
              {amount} {fromSym}
            </div>
          </div>

          <div className="flex justify-center">
            <div
              className="w-8 h-8 rounded-full flex items-center justify-center text-[var(--text-muted)]"
              style={{ background: 'var(--chassis)', boxShadow: 'var(--shadow-card)' }}
            >
              <span className="text-sm leading-none">↓</span>
            </div>
          </div>

          <div
            className="rounded-xl p-4"
            style={{ background: 'var(--recessed)', boxShadow: 'var(--shadow-recessed)' }}
          >
            <div className="font-mono text-[10px] uppercase tracking-[0.1em] text-[var(--text-muted)] mb-2">
              You receive
            </div>
            <div className="font-sans font-bold text-2xl text-[var(--text)] tracking-tight">
              {quote.amountOut.toFixed(6)} {toSym}
            </div>
          </div>
        </div>

        {/* Detail rows */}
        <div>
          <DataRow label="Route"    value={quote.protocol}                         accent />
          <DataRow label="Rate"     value={`1 ${fromSym} = ${rate} ${toSym}`} />
          <DataRow label="Slippage" value={`${(slippage / 100).toFixed(1)}%`} />
        </div>

        {/* Buttons */}
        <div className="flex gap-3 pt-1">
          <button
            onClick={onCancel}
            className="flex-1 py-3.5 rounded-xl font-mono text-xs uppercase tracking-[0.1em] text-[var(--text-muted)] hover:text-[var(--text)] shadow-card active:shadow-pressed active:translate-y-[1px] transition-all duration-150"
            style={{ background: 'var(--chassis)' }}
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            className="flex-1 py-3.5 rounded-xl bg-[var(--accent)] text-white font-mono text-xs uppercase tracking-[0.1em] font-semibold shadow-accent hover:brightness-110 active:shadow-accent-pressed active:translate-y-[1px] transition-all duration-150"
          >
            Confirm
          </button>
        </div>
      </div>
    </div>
  )
}
