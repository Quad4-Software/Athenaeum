import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const githubRepo =
  process.env.GITHUB_REPOSITORY ??
  process.env.DOCUSAURUS_GITHUB_REPO ??
  'Quad4-Software/Athenaeum';
const [organizationName, projectName] = githubRepo.split('/');
// Custom domain serves at the site root. Project Pages path (/repo/) is opt-in via env.
const baseUrl = process.env.DOCUSAURUS_BASE_URL ?? '/';
const siteUrl = process.env.DOCUSAURUS_URL ?? 'https://athenaeum.quad4.io';
const githubUrl = `https://github.com/${githubRepo}`;
const demoHref = 'pathname:///demo/';

const config: Config = {
  title: 'Athenaeum',
  tagline: 'Self-hosted library for EPUB, PDF, comics, and audiobooks.',
  favicon: 'img/favicon.svg',

  future: {
    v4: true,
  },

  url: siteUrl,
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
          editUrl: `${githubUrl}/tree/master/docs/docs/`,
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
      // Desktop/mobile chrome is custom (src/theme/Navbar). Keep items for the
      // Docusaurus mobile sidebar link list.
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
    // Footer chrome is custom (src/theme/Footer). Keep a minimal config so
    // theme hooks that read footer still resolve.
    footer: {
      style: 'dark',
      copyright: `Copyright ${new Date().getFullYear()} Athenaeum contributors. MIT licensed.`,
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
