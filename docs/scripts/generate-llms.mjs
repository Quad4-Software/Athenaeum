/**
 * Generate /llms.txt and /llms-full.txt into the production build output.
 * Runs after `docusaurus build` (see the build script in package.json).
 *
 * llms.txt is a curated index aimed at agents: deployers first, then
 * operators and contributors. llms-full.txt inlines every doc page in
 * sidebar order so an agent can pull the whole corpus in one request.
 */
import {readFile, writeFile, mkdir} from 'node:fs/promises';
import {join} from 'node:path';

const siteUrl = (process.env.DOCUSAURUS_URL ?? 'https://athenaeum.quad4.io').replace(/\/$/, '');
const baseUrl = process.env.DOCUSAURUS_BASE_URL ?? '/';
const docsDir = new URL('../docs/', import.meta.url).pathname;
const outDir = new URL('../build/', import.meta.url).pathname;

// Mirrors the order in sidebars.ts.
const pages = [
  'intro',
  'getting-started',
  'features',
  'deploying',
  'configuration',
  'authentication',
  'cli-users',
  'library',
  'catalogs',
  'operations',
  'http-api',
  'development',
  'project-layout',
];

function docUrl(slug) {
  return `${siteUrl}${baseUrl}docs/${slug}`.replace(/(?<!:)\/{2,}/g, '/');
}

function stripFrontmatter(source) {
  return source.replace(/^---\n[\s\S]*?\n---\n*/, '');
}

function absolutizeLinks(markdown) {
  return markdown
    .replace(/\[([^\]]*)\]\((\.\/[^)]+)\)/g, (_m, text, href) => {
      const target = href.replace(/^\.\//, '').replace(/#.*$/, '');
      return `[${text}](${docUrl(target)})`;
    })
    .replace(/\[([^\]]*)\]\((\/[^)]+)\)/g, (_m, text, href) => {
      return `[${text}](${siteUrl}${href})`;
    });
}

const llmsIndex = `# Athenaeum

> Self-hosted library for EPUB, PDF, comics, and audiobooks. One Go binary or one Docker image; point it at a media folder to get browser readers, search, shelves, optional multi-user auth, and OPDS.

Athenaeum listens on :8080 by default and keeps state in its --data directory. Server settings are CLI flags or ATHENAEUM_* environment variables. The container image is ghcr.io/quad4-software/athenaeum and the source is at github.com/Quad4-Software/Athenaeum.

## Deploy

- [Getting started](${docUrl('getting-started')}): install.sh, docker run, Compose, release binary, or source build
- [Deploying](${docUrl('deploying')}): installer flags, Compose profiles (altcha, kokoro, postgres), host service units, backups
- [Configuration](${docUrl('configuration')}): every CLI flag and ATHENAEUM_* env var with defaults
- [Operations](${docUrl('operations')}): metrics, health checks, maintenance tasks, webhooks, backup API

## Operate

- [Authentication](${docUrl('authentication')}): first admin, sessions, guests, invites, TOTP, OIDC/Pocket ID, API keys, ALTCHA
- [CLI users](${docUrl('cli-users')}): offline account management via \`athenaeum users\`
- [Library and readers](${docUrl('library')}): formats, S3 mounts, metadata providers, BibTeX, narration, sharing
- [OPDS and KOSync](${docUrl('catalogs')}): e-reader catalog feeds and KOReader progress sync endpoints
- [HTTP API](${docUrl('http-api')}): full REST route reference

## Contribute

- [Features](${docUrl('features')}): capability list and tech stack
- [Development](${docUrl('development')}): Task targets, test methods, and CI workflows
- [Project layout](${docUrl('project-layout')}): repository directory map

## Optional

- [Introduction](${docUrl('intro')}): product overview and fit
- [Roadmap](${siteUrl}/roadmap): planned work
- [Privacy](${siteUrl}/privacy): data handling stance
- [GitHub](https://github.com/Quad4-Software/Athenaeum): source, issues, and releases
`;

const header = `# Athenaeum documentation

> Full docs corpus for Athenaeum, a self-hosted library for EPUB, PDF, comics, and audiobooks. Curated index: ${siteUrl}/llms.txt
`;

async function main() {
  const sections = [header];
  for (const slug of pages) {
    const source = await readFile(join(docsDir, `${slug}.md`), 'utf8');
    const body = absolutizeLinks(stripFrontmatter(source)).trim();
    sections.push(`## ${docUrl(slug)}\n\n${body}\n\n---\n`);
  }

  await mkdir(outDir, {recursive: true});
  await writeFile(join(outDir, 'llms.txt'), llmsIndex);
  await writeFile(join(outDir, 'llms-full.txt'), sections.join('\n'));
  console.log(`Generated llms.txt and llms-full.txt (${pages.length} pages)`);
}

await main();
