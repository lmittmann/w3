import { Link } from 'nextra-theme-docs'
import { Code } from 'nextra/components'

const pkgNameToPath = {
    'w3': 'w3',
    'module': 'w3/module',
    'debug': 'w3/module/debug',
    'eth': 'w3/module/eth',
    'txpool': 'w3/module/txpool',
    'web3': 'w3/module/web3',
    'w3types': 'w3/w3types',
    'w3vm': 'w3/w3vm',
}

export const RefLink = ({ title }) => {
    let [pkg, comp] = title.split('.', 2)
    let url = `https://pkg.go.dev/github.com/lmittmann/${pkgNameToPath[pkg]}#${comp}`
    return (
        <Link href={url}>
            <Code>{title}</Code>
        </Link>
    )
}
