import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const githubRepo =
  process.env.GITHUB_REPOSITORY ??
  process.env.DOCUSAURUS_GITHUB_REPO ??
  'ivan/reader';
const [organizationName, projectName] = githubRepo.split('/');
const baseUrl =
  process.env.DOCUSAURUS_BASE_URL ??
  (process.env.CI ? `/${projectName}/` : '/');
const githubUrl = `https://github.com/${githubRepo}`;
const demoHref = 'pathname:///demo/';

const config: Config = {
  title: 'Athenaeum',
  tagline:
    'Self-hosted EPUB, PDF, and audiobook library in a single static binary.',
  favicon: 'img/favicon.svg',

  future: {
    v4: true,
  },

  url: process.env.DOCUSAURUS_URL ?? `https://${organizationName}.github.io`,
  baseUrl,
  organizationName,
  projectName,
  trailingSlash: false,

  onBrokenLinks: 'throw',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          routeBasePath: 'docs',
          editUrl: `${githubUrl}/tree/main/docs/docs/`,
          showLastUpdateTime: false,
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/showcase/library-dark.png',
    colorMode: {
      defaultMode: 'dark',
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },
    docs: {
      sidebar: {
        hideable: true,
        autoCollapseCategories: true,
      },
    },
    navbar: {
      title: 'Athenaeum',
      logo: {
        alt: 'Athenaeum',
        src: 'img/logo.svg',
        srcDark: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
        },
        {
          to: '/docs/getting-started',
          label: 'Get started',
          position: 'left',
        },
        {
          href: demoHref,
          label: 'Live demo',
          position: 'left',
          target: '_self',
        },
        {
          to: '/roadmap',
          label: 'Roadmap',
          position: 'left',
        },
        {
          href: githubUrl,
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Docs',
          items: [
            {label: 'Getting started', to: '/docs/getting-started'},
            {label: 'Features', to: '/docs/features'},
            {label: 'Deploying', to: '/docs/deploying'},
            {label: 'Configuration', to: '/docs/configuration'},
          ],
        },
        {
          title: 'Guides',
          items: [
            {label: 'Authentication', to: '/docs/authentication'},
            {label: 'Library', to: '/docs/library'},
            {label: 'OPDS and KOSync', to: '/docs/catalogs'},
            {label: 'HTTP API', to: '/docs/http-api'},
          ],
        },
        {
          title: 'Project',
          items: [
            {label: 'Roadmap', to: '/roadmap'},
            {label: 'GitHub', href: githubUrl},
            {label: 'Changelog', href: `${githubUrl}/blob/main/CHANGELOG.md`},
            {label: 'Contributing', href: `${githubUrl}/blob/main/CONTRIBUTING.md`},
            {label: 'Live demo', href: demoHref},
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Athenaeum contributors. MIT licensed.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.oneDark,
      additionalLanguages: ['bash', 'ini', 'json', 'docker'],
    },
    tableOfContents: {
      minHeadingLevel: 2,
      maxHeadingLevel: 3,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
