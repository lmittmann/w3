import './style.css'
import 'nextra-theme-docs/style.css'
import { Inter } from 'next/font/google'

const inter = Inter({ subsets: ['latin'] })

export default function Nextra({ Component, pageProps }) {
    const getLayout = Component.getLayout || ((page) => page)
    return getLayout(
        <div className={inter.className}>
            <Component  {...pageProps} />
        </div>
    )
}
