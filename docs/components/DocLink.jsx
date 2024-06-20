import { Link } from 'nextra-theme-docs'
import { Code } from 'nextra/components'

const pkgNameToPath = {
    'w3': 'github.com/lmittmann/w3',
    'module': 'github.com/lmittmann/w3/module',
    'debug': 'github.com/lmittmann/w3/module/debug',
    'eth': 'github.com/lmittmann/w3/module/eth',
    'txpool': 'github.com/lmittmann/w3/module/txpool',
    'web3': 'github.com/lmittmann/w3/module/web3',
    'w3types': 'github.com/lmittmann/w3/w3types',
    'w3vm': 'github.com/lmittmann/w3/w3vm',
}

export const DocLink = ({ title }) => {
    let [pkg, comp] = title.split('.', 2)
    let url = `https://pkg.go.dev/${pkgNameToPath[pkg]}#${comp}`
    return (
        <Link href={url}>
            <Code>{title}</Code>
        </Link>
    )
}
