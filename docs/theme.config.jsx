import { useRouter } from 'next/router'
import Image from 'next/image'

export default {
    logo: <>
        <div className="rounded-full h-10 w-10 mr-2 overflow-hidden bg-black/10 dark:bg-white/10">
            <Image src="/gopher.png" alt="w3" width={48} height={48} className='w-11/12 mx-auto' />
        </div>
        <span className="text-2xl font-bold">w3</span>
    </>,
    banner: {
        text: '🚧 This site is currently under construction 🚧'
    },
    useNextSeoProps() {
        const { pathname } = useRouter()
        return { titleTemplate: pathname === '/' ? 'w3' : '%s – w3' }
    },
    footer: {
        component: null,
    },
    project: {
        link: 'https://github.com/lmittmann/w3',
    },
    editLink: {
        text: 'Edit this page on GitHub'
    },
    feedback: {
        content: null,
    },
    docsRepositoryBase: 'https://github.com/lmittmann/w3/blob/main/docs/pages',
    primaryHue: {
        dark: 189,
        light: 191
    }
}
