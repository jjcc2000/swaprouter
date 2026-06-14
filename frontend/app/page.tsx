'use client'

import { useEffect, useState } from 'react'
import { useWallet } from '@solana/wallet-adapter-react'
import { WalletMultiButton } from '@solana/wallet-adapter-react-ui'
import SwapForm from '@/components/SwapForm'
import TradeHistory from '@/components/TradeHistory'
import { getNonce, login } from '@/lib/api'

/* ── Micro-components ─────────────────────────────────────────────────── */

function LED({ color = 'red', label }: { color?: 'red' | 'green' | 'yellow'; label?: string }) {
  const map = {
    red:    { dot: 'bg-[var(--accent)]',    anim: 'animate-led-red',    shadow: '0 0 6px 2px rgba(255,71,87,0.8)'   },
    green:  { dot: 'bg-green-400',          anim: 'animate-led-green',  shadow: '0 0 6px 2px rgba(34,197,94,0.8)'  },
    yellow: { dot: 'bg-yellow-400',         anim: 'animate-led-yellow', shadow: '0 0 6px 2px rgba(250,204,21,0.8)' },
  }
  const c = map[color]
  return (
    <div className="flex items-center gap-2 shrink-0">
      <div className={`w-2 h-2 rounded-full ${c.dot} ${c.anim}`} style={{ boxShadow: c.shadow }} />
      {label && (
        <span className="font-mono text-[10px] uppercase tracking-[0.1em] text-[var(--text-muted)]">
          {label}
        </span>
      )}
    </div>
  )
}

function VentSlots() {
  return (
    <div className="flex gap-1">
      {[0, 1, 2].map(i => (
        <div
          key={i}
          className="h-5 w-px rounded-full bg-[var(--recessed)]"
          style={{ boxShadow: 'inset 1px 1px 2px #0c0f15' }}
        />
      ))}
    </div>
  )
}

/* ── Landing screen ───────────────────────────────────────────────────── */
function Landing() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen px-4 gap-0">
      {/* Radial hotspot — top-left lighting source */}
      <div
        className="pointer-events-none fixed inset-0"
        style={{
          background:
            'radial-gradient(ellipse 70% 55% at 20% 20%, rgba(255,255,255,0.025) 0%, transparent 60%)',
        }}
      />

      <div
        className="relative w-full max-w-sm rounded-2xl p-10 flex flex-col items-center gap-7 screws"
        style={{ background: 'var(--panel)', boxShadow: 'var(--shadow-floating)' }}
      >
        {/* Status row */}
        <div className="flex items-center justify-between w-full">
          <LED color="green" label="System Ready" />
          <VentSlots />
        </div>

        {/* Logo */}
        <div className="flex flex-col items-center gap-1">
          <div className="font-mono text-[10px] uppercase tracking-[0.15em] text-[var(--text-faint)]">
            // v1.0.0
          </div>
          <h1 className="font-sans font-extrabold text-5xl text-[var(--text)] tracking-tight"
              style={{ textShadow: '0 2px 4px rgba(0,0,0,0.6), 0 1px 0 rgba(255,255,255,0.04)' }}>
            SwapRouter
          </h1>
          <div className="font-mono text-[10px] uppercase tracking-[0.15em] text-[var(--text-muted)] mt-1">
            Best rate across every DEX
          </div>
        </div>

        {/* Divider */}
        <div className="w-full h-px" style={{ background: 'linear-gradient(90deg, transparent, var(--border), transparent)' }} />

        <WalletMultiButton />

        {/* Bottom label */}
        <div className="font-mono text-[9px] uppercase tracking-[0.12em] text-[var(--text-faint)]">
          Non-custodial · Solana
        </div>
      </div>
    </div>
  )
}

/* ── Auth screen ──────────────────────────────────────────────────────── */
function AuthScreen({
  authing,
  authError,
  onRetry,
}: {
  authing: boolean
  authError: string
  onRetry: () => void
}) {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen px-4">
      <div
        className="relative w-full max-w-sm rounded-2xl p-8 flex flex-col gap-5 screws"
        style={{ background: 'var(--panel)', boxShadow: 'var(--shadow-floating)' }}
      >
        <div className="flex items-center justify-between">
          <span className="font-mono text-[10px] uppercase tracking-[0.12em] text-[var(--text-muted)]">
            // Auth
          </span>
          <LED color={authing ? 'yellow' : 'red'} />
        </div>

        <h2 className="font-sans font-bold text-xl text-[var(--text)]">Sign In</h2>

        <p className="font-mono text-xs text-[var(--text-muted)] uppercase tracking-[0.06em]">
          {authing ? 'Sign the message in Phantom…' : 'Wallet signature required'}
        </p>

        {authError && (
          <p className="font-mono text-xs text-[var(--accent)]">{authError}</p>
        )}

        {!authing && (
          <button
            onClick={onRetry}
            className="mt-2 w-full py-3.5 rounded-xl bg-[var(--accent)] text-white font-mono text-xs uppercase tracking-[0.1em] font-semibold shadow-accent hover:brightness-110 active:translate-y-[2px] active:shadow-accent-pressed transition-all duration-150"
          >
            Sign In
          </button>
        )}
      </div>
    </div>
  )
}

/* ── Main app ─────────────────────────────────────────────────────────── */
export default function Home() {
  const { publicKey, signMessage, connected, disconnect } = useWallet()
  const [token, setToken]         = useState<string | null>(null)
  const [tab, setTab]             = useState<'swap' | 'history'>('swap')
  const [authing, setAuthing]     = useState(false)
  const [authError, setAuthError] = useState('')

  useEffect(() => {
    if (!connected || !publicKey) { setToken(null); return }
    const stored = localStorage.getItem(`jwt_${publicKey.toBase58()}`)
    if (stored) setToken(stored)
    else authenticate()
  }, [connected, publicKey?.toBase58()])

  async function authenticate() {
    if (!publicKey || !signMessage || authing) return
    setAuthing(true)
    setAuthError('')
    try {
      const wallet      = publicKey.toBase58()
      const { message } = await getNonce(wallet)
      const signature   = await signMessage(new TextEncoder().encode(message))
      const sigHex      = Buffer.from(signature).toString('hex')
      const res         = await login(wallet, sigHex)
      if (!res.token) throw new Error(res.message ?? 'Login failed')
      localStorage.setItem(`jwt_${wallet}`, res.token)
      setToken(res.token)
    } catch (e: any) {
      setAuthError(e.message)
    } finally {
      setAuthing(false)
    }
  }

  function handleDisconnect() {
    if (publicKey) localStorage.removeItem(`jwt_${publicKey.toBase58()}`)
    disconnect()
    setToken(null)
  }

  if (!connected) return <Landing />
  if (authing || !token) {
    return <AuthScreen authing={authing} authError={authError} onRetry={authenticate} />
  }

  const wallet = publicKey!.toBase58()
  const short  = `${wallet.slice(0, 4)}···${wallet.slice(-4)}`

  return (
    <div className="min-h-screen flex flex-col">

      {/* ── Navbar ──────────────────────────────────────────────────── */}
      <nav
        className="flex items-center justify-between px-6 py-4 border-b"
        style={{ borderColor: 'var(--border)', background: 'var(--panel)', boxShadow: '0 2px 8px #0c0f15' }}
      >
        <div className="flex items-center gap-4">
          <LED color="green" />
          <span className="font-sans font-bold text-base text-[var(--text)] tracking-tight">
            SwapRouter
          </span>
        </div>

        <div className="flex items-center gap-5">
          <span className="font-mono text-[11px] uppercase tracking-[0.08em] text-[var(--text-muted)]">
            {short}
          </span>
          <VentSlots />
          <button
            onClick={handleDisconnect}
            className="font-mono text-[10px] uppercase tracking-[0.1em] text-[var(--text-faint)] hover:text-[var(--accent)] transition-colors duration-200"
          >
            Eject
          </button>
        </div>
      </nav>

      {/* ── Tab bar ──────────────────────────────────────────────────── */}
      <div className="flex items-center justify-center gap-3 pt-8 pb-6">
        {(['swap', 'history'] as const).map(t => {
          const active = tab === t
          return (
            <button
              key={t}
              onClick={() => setTab(t)}
              className={[
                'font-mono text-xs uppercase tracking-[0.1em] px-7 py-3 rounded-xl transition-all duration-200',
                active
                  ? 'text-[var(--accent)] shadow-card'
                  : 'text-[var(--text-muted)] hover:text-[var(--text)]',
              ].join(' ')}
              style={
                active
                  ? { background: 'var(--panel)', boxShadow: 'var(--shadow-card)' }
                  : { background: 'transparent' }
              }
            >
              {t === 'swap' ? 'Swap' : 'History'}
            </button>
          )
        })}
      </div>

      {/* ── Content ──────────────────────────────────────────────────── */}
      <div className="flex justify-center px-4 pb-16">
        {tab === 'swap'
          ? <SwapForm token={token} />
          : <TradeHistory token={token} />}
      </div>
    </div>
  )
}
