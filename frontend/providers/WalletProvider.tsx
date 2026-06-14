'use client'

import { FC, ReactNode, useMemo } from 'react'
import { ConnectionProvider, WalletProvider } from '@solana/wallet-adapter-react'
import { WalletAdapterNetwork } from '@solana/wallet-adapter-base'
import { PhantomWalletAdapter } from '@solana/wallet-adapter-phantom'
import { WalletModalProvider } from '@solana/wallet-adapter-react-ui'
import { clusterApiUrl } from '@solana/web3.js'

// Type cast needed: Solana wallet adapters were built against React 17 types
const CP  = ConnectionProvider  as React.ComponentType<any>
const WP  = WalletProvider      as React.ComponentType<any>
const WMP = WalletModalProvider as React.ComponentType<any>

export const WalletContextProvider: FC<{ children: ReactNode }> = ({ children }) => {
  const network  = WalletAdapterNetwork.Mainnet
  const endpoint = useMemo(() => clusterApiUrl(network), [network])
  const wallets  = useMemo(() => [new PhantomWalletAdapter()], [network])

  return (
    <CP endpoint={endpoint}>
      <WP wallets={wallets} autoConnect>
        <WMP>{children}</WMP>
      </WP>
    </CP>
  )
}
