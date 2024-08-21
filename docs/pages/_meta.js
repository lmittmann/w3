export default {
    index: {
        title: 'Overview',
        display: 'hidden',
        theme: { breadcrumb: false },
    },
    '404': {
        title: '404',
        display: 'hidden',
        'theme': {
            breadcrumb: false,
            toc: false,
            layout: 'full',
            pagination: false,
        }
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
    '+++ VM': {
        'title': '',
        'type': 'separator'
    },
    '--- VM': {
        type: 'separator',
        title: 'VM'
    },
    'vm-overview': {
        title: 'Overview',
        theme: { breadcrumb: false },
    },
    'vm-tracing': {
        title: 'Tracing',
        theme: { breadcrumb: false },
    },
    'vm-testing': {
        title: 'Testing',
        theme: { breadcrumb: false },
    },
    '+++ HELPER': {
        'title': '',
        'type': 'separator'
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
        href: 'https://github.com/lmittmann/w3/tree/main/examples',
        newWindow: true
    },
    godoc: {
        title: 'GoDoc',
        type: 'page',
        href: 'https://pkg.go.dev/github.com/lmittmann/w3#section-documentation',
        newWindow: true
    },
}
