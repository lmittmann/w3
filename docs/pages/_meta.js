export default {
    index: {
        title: 'Overview',
        display: 'hidden',
        theme: { breadcrumb: false },
    },
    '--- RPC Client': {
        type: 'separator',
        title: 'RPC Client'
    },
    'rpc-overview': {
        title: 'Overview',
        theme: { breadcrumb: false },
    },
    'rpc-methods': {
        title: 'Methods',
        theme: { breadcrumb: false },
    },
    'rpc-extension': {
        title: 'Extension',
        theme: { breadcrumb: false },
    },
    '--- VM': {
        type: 'separator',
        title: 'VM'
    },
    'vm-overview': {
        title: 'Overview',
        theme: { breadcrumb: false },
    },
    '--- HELPER': {
        type: 'separator',
        title: 'Helper'
    },
    'helper-abi': {
        title: 'ABI',
        theme: { breadcrumb: false },
    },
    'helper-utils': {
        title: 'Utils',
        theme: { breadcrumb: false },
    },
    examples: {
        title: 'Examples',
        type: 'page',
        href: '/examples',
        newWindow: true
    },
    godoc: {
        title: 'GoDoc',
        type: 'page',
        href: 'https://pkg.go.dev/github.com/lmittmann/w3#section-documentation',
        newWindow: true
    },
}
