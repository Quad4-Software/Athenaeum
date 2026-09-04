import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {BookOpen, Github} from 'lucide-react';

import styles from './styles.module.css';

type FooterLink = {
  label: string;
  to?: string;
  href?: string;
  target?: string;
};

type FooterColumn = {
  title: string;
  items: FooterLink[];
};

function FooterLinkItem({item}: {item: FooterLink}): ReactNode {
  if (item.href) {
    return (
      <Link
        className={styles.link}
        href={item.href}
        target={item.target}
        rel={item.target === '_blank' ? 'noopener noreferrer' : undefined}>
        {item.label}
      </Link>
    );
  }
  return (
    <Link className={styles.link} to={item.to}>
      {item.label}
    </Link>
  );
}

export default function SiteFooter(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  const logoSrc = useBaseUrl('/img/logo.svg');
  const demoHref = useBaseUrl('pathname:///demo/');
  const githubUrl = `https://github.com/${siteConfig.organizationName}/${siteConfig.projectName}`;
  const year = new Date().getFullYear();

  const columns: FooterColumn[] = [
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
        {label: 'Privacy', to: '/privacy'},
        {label: 'Changelog', href: `${githubUrl}/blob/main/CHANGELOG.md`},
        {label: 'Contributing', href: `${githubUrl}/blob/main/CONTRIBUTING.md`},
      ],
    },
  ];

  return (
    <footer className={styles.footer}>
      <div className={clsx('container', styles.inner)}>
        <div className={styles.brandBlock}>
          <Link className={styles.brand} to="/">
            <img
              className={styles.brandLogo}
              src={logoSrc}
              alt=""
              width={32}
              height={32}
            />
            <span className={styles.brandName}>{siteConfig.title}</span>
          </Link>
          <p className={styles.tagline}>{siteConfig.tagline}</p>
          <div className={styles.brandActions}>
            <Link className={styles.primaryAction} to="/docs/getting-started">
              <BookOpen size={16} strokeWidth={1.75} aria-hidden="true" />
              Get started
            </Link>
            <Link
              className={styles.secondaryAction}
              href={githubUrl}
              target="_blank"
              rel="noopener noreferrer">
              <Github size={16} strokeWidth={1.75} aria-hidden="true" />
              GitHub
            </Link>
            <Link className={styles.secondaryAction} href={demoHref} target="_self">
              Live demo
            </Link>
          </div>
        </div>

        <div className={styles.columns}>
          {columns.map((column) => (
            <div key={column.title} className={styles.column}>
              <h2 className={styles.columnTitle}>{column.title}</h2>
              <ul className={styles.columnList}>
                {column.items.map((item) => (
                  <li key={item.label}>
                    <FooterLinkItem item={item} />
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </div>

      <div className={styles.bottom}>
        <div className={clsx('container', styles.bottomInner)}>
          <p className={styles.copyright}>
            Copyright {year} Athenaeum contributors. MIT licensed.
          </p>
          <p className={styles.note}>Self-hosted. No telemetry by default.</p>
        </div>
      </div>
    </footer>
  );
}
