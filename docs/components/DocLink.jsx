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

export const DocLink = ({ title, id }) => {
    if (typeof id === 'undefined') { id = title }
    let dotIndex = id.indexOf('.');
    let pkg = id.substring(0, dotIndex);
    let comp = id.substring(dotIndex + 1);

    return (
        <Link href={`https://pkg.go.dev/${pkgNameToPath[pkg]}#${comp}`}>
            <Code>{title}</Code>
        </Link>
    )
}
