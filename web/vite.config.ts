import { fileURLToPath, URL } from 'node:url'

import { defineConfig, loadEnv, type HtmlTagDescriptor, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

import tailwindcss from '@tailwindcss/vite'

const SITE = {
  name: 'birdwatcher',
  description:
    "Crawl websites and collect bird sightings matched against Audubon's 435 plates of The Birds of America.",
  author: 'Umut Karakulak',
  locale: 'en_US',
  ogImagePath: '/og-image.png',
  ogImageWidth: '1200',
  ogImageHeight: '630',
  ogImageAlt:
    "birdwatcher — Crawl websites and catalog bird sightings against Audubon's 435 plates of The Birds of America.",
} as const

function escapeXml(value: string): string {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&apos;')
}

function buildRobotsTxt(siteUrl: string): string {
  const lines = ['User-agent: *', 'Allow: /', 'Disallow: /api/']
  if (siteUrl) {
    lines.push('', `Sitemap: ${siteUrl}/sitemap.xml`)
  }
  return `${lines.join('\n')}\n`
}

function buildSitemapXml(siteUrl: string): string {
  return [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">',
    '  <url>',
    `    <loc>${escapeXml(siteUrl)}/</loc>`,
    '  </url>',
    '</urlset>',
    '',
  ].join('\n')
}

function buildJsonLd(siteUrl: string): string {
  const schema: Record<string, unknown> = {
    '@context': 'https://schema.org',
    '@type': 'WebApplication',
    name: SITE.name,
    description: SITE.description,
    applicationCategory: 'ReferenceApplication',
    operatingSystem: 'Any',
    author: {
      '@type': 'Person',
      name: SITE.author,
    },
  }

  if (siteUrl) {
    schema.url = `${siteUrl}/`
    schema.image = `${siteUrl}${SITE.ogImagePath}`
  }

  return JSON.stringify(schema).replaceAll('<', '\\u003c')
}

function buildSeoHeadTags(siteUrl: string): HtmlTagDescriptor[] {
  const ogImageUrl = siteUrl ? `${siteUrl}${SITE.ogImagePath}` : SITE.ogImagePath
  const meta = (attrs: Record<string, string>): HtmlTagDescriptor => ({
    tag: 'meta',
    attrs,
    injectTo: 'head',
  })

  const tags: HtmlTagDescriptor[] = [
    meta({ name: 'description', content: SITE.description }),
    meta({ name: 'author', content: SITE.author }),
  ]

  if (siteUrl) {
    tags.push({ tag: 'link', attrs: { rel: 'canonical', href: `${siteUrl}/` }, injectTo: 'head' })
  }

  tags.push(
    meta({ property: 'og:type', content: 'website' }),
    meta({ property: 'og:site_name', content: SITE.name }),
    meta({ property: 'og:locale', content: SITE.locale }),
    meta({ property: 'og:title', content: SITE.name }),
    meta({ property: 'og:description', content: SITE.description }),
  )

  if (siteUrl) {
    tags.push(meta({ property: 'og:url', content: `${siteUrl}/` }))
  }

  tags.push(
    meta({ property: 'og:image', content: ogImageUrl }),
    meta({ property: 'og:image:width', content: SITE.ogImageWidth }),
    meta({ property: 'og:image:height', content: SITE.ogImageHeight }),
    meta({ property: 'og:image:alt', content: SITE.ogImageAlt }),
    meta({ name: 'twitter:card', content: 'summary_large_image' }),
    meta({ name: 'twitter:title', content: SITE.name }),
    meta({ name: 'twitter:description', content: SITE.description }),
    meta({ name: 'twitter:image', content: ogImageUrl }),
    meta({ name: 'twitter:image:alt', content: SITE.ogImageAlt }),
    {
      tag: 'script',
      attrs: { type: 'application/ld+json' },
      children: buildJsonLd(siteUrl),
      injectTo: 'head',
    },
  )

  return tags
}

function seoMetadataPlugin(siteUrl: string): Plugin {
  return {
    name: 'birdwatcher-seo-metadata',
    transformIndexHtml() {
      return buildSeoHeadTags(siteUrl)
    },
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const pathname = req.url?.split('?')[0]
        if (pathname === '/robots.txt') {
          res.setHeader('Content-Type', 'text/plain; charset=utf-8')
          res.end(buildRobotsTxt(siteUrl))
          return
        }
        if (pathname === '/sitemap.xml' && siteUrl) {
          res.setHeader('Content-Type', 'application/xml; charset=utf-8')
          res.end(buildSitemapXml(siteUrl))
          return
        }
        next()
      })
    },
    generateBundle() {
      this.emitFile({
        type: 'asset',
        fileName: 'robots.txt',
        source: buildRobotsTxt(siteUrl),
      })

      if (siteUrl) {
        this.emitFile({
          type: 'asset',
          fileName: 'sitemap.xml',
          source: buildSitemapXml(siteUrl),
        })
      }
    },
  }
}

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const rootDir = fileURLToPath(new URL('.', import.meta.url))
  const env = loadEnv(mode, rootDir, 'VITE_')
  const siteUrl = (env.VITE_SITE_URL ?? '').trim().replace(/\/+$/, '')

  return {
    plugins: [vue(), vueDevTools(), tailwindcss(), seoMetadataPlugin(siteUrl)],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
        },
      },
    },
  }
})
