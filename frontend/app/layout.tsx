import type { Metadata } from 'next'
import { Space_Grotesk, IBM_Plex_Mono } from 'next/font/google'
import './globals.css'
import '@solana/wallet-adapter-react-ui/styles.css'
import { WalletContextProvider } from '@/providers/WalletProvider'

const spaceGrotesk = Space_Grotesk({
  subsets:  ['latin'],
  weight:   ['300', '400', '500', '600', '700'],
  variable: '--font-sans',
  display:  'swap',
})

const ibmPlexMono = IBM_Plex_Mono({
  subsets:  ['latin'],
  weight:   ['400', '500', '600', '700'],
  variable: '--font-mono',
  display:  'swap',
})

export const metadata: Metadata = {
  title:       'SwapRouter',
  description: 'Best rate DEX aggregator',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${spaceGrotesk.variable} ${ibmPlexMono.variable}`}>
      <body className="min-h-screen">
        <WalletContextProvider>{children}</WalletContextProvider>
      </body>
    </html>
  )
}
