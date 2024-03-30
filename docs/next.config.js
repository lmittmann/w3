import nextra from 'nextra'

const withNextra = nextra({
	theme: 'nextra-theme-docs',
	themeConfig: './theme.config.jsx',
	staticImage: true,
	search: {
		codeblocks: true
	},
	defaultShowCopyCode: true
})

export default withNextra({
	output: 'export',
	reactStrictMode: true,
	images: {
		unoptimized: true,
	}
})
