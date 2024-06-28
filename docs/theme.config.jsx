import Image from 'next/image'
import { useConfig } from 'nextra-theme-docs'
import { Cards, Steps } from 'nextra/components'
import { DocLink } from './components/DocLink'

export default {
    logo: <>
        <div className="rounded-full h-10 w-10 mr-2 overflow-hidden bg-black/10 dark:bg-white/10">
            <Image src="/gopher.png" alt="w3" width={48} height={48} className='w-11/12 mx-auto' />
        </div>
        <span className="text-2xl font-bold">w3</span>
    </>,
    head: () => {
        const { frontMatter } = useConfig()
        const title = frontMatter.title ? `${frontMatter.title} – w3` : 'w3'
        return (
            <>
                <title>{title}</title>
                <meta property="og:title" content={title} />
                <meta
                    property="og:description"
                    content={frontMatter.description || 'w3'}
                />
            </>
        )
    },
    footer: {
        component: null,
    },
    components: {
        Card: Cards.Card,
        Cards: Cards,
        Steps: Steps,
        DocLink: DocLink,
    },
    project: {
        link: 'https://github.com/lmittmann/w3',
    },
    feedback: {
        content: null,
    },
    docsRepositoryBase: 'https://github.com/lmittmann/w3/blob/main/docs',
    color: {
        hue: 189,
        saturation: 100,
    },
}
