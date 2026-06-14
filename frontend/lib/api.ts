const API = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080'

function headers(token?: string): Record<string, string> {
  const h: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) h['Authorization'] = `Bearer ${token}`
  return h
}

export async function getNonce(wallet: string): Promise<{ message: string }> {
  const res = await fetch(`${API}/auth/nonce?wallet=${wallet}`)
  return res.json()
}

export async function login(wallet: string, signature: string): Promise<{ token?: string; code?: string; message?: string }> {
  const res = await fetch(`${API}/auth/login`, {
    method: 'POST',
    headers: headers(),
    body: JSON.stringify({ wallet, signature }),
  })
  return res.json()
}

export interface Quote {
  quoteId: string
  protocol: string
  fromToken: string
  toToken: string
  amountIn: number
  amountOut: number
  chain: string
  expiresAt: string
  code?: string
  message?: string
}

export async function getQuote(
  token: string,
  params: { fromToken: string; toToken: string; amount: number; chain: string }
): Promise<Quote> {
  const url = new URL(`${API}/v1/quote`)
  url.searchParams.set('fromToken', params.fromToken)
  url.searchParams.set('toToken', params.toToken)
  url.searchParams.set('amount', String(params.amount))
  url.searchParams.set('chain', params.chain)
  const res = await fetch(url.toString(), { headers: headers(token) })
  return res.json()
}

export interface SwapResult {
  tradeId: string
  unsignedTx: string
  status: string
  chain: string
  protocol: string
  fromToken: string
  toToken: string
  amountIn: number
  amountOut: number
  code?: string
  message?: string
}

export async function executeSwap(token: string, quoteId: string, slippageBps = 50): Promise<SwapResult> {
  const res = await fetch(`${API}/v1/swap`, {
    method: 'POST',
    headers: headers(token),
    body: JSON.stringify({ quoteId, slippageBps }),
  })
  return res.json()
}

export async function confirmTrade(token: string, tradeId: string, txHash: string) {
  const res = await fetch(`${API}/v1/trades/confirm`, {
    method: 'PATCH',
    headers: headers(token),
    body: JSON.stringify({ tradeId, txHash }),
  })
  return res.json()
}

export interface Trade {
  id: string
  txHash: string
  wallet: string
  chain: string
  protocol: string
  fromToken: string
  toToken: string
  amountIn: number
  amountOut: number
  status: 'unsigned' | 'pending' | 'confirmed' | 'failed'
  createdAt: string
}

export async function getTrades(token: string): Promise<{ trades: Trade[] }> {
  const res = await fetch(`${API}/v1/trades`, { headers: headers(token) })
  return res.json()
}
