import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import {useLocation} from '@docusaurus/router';
import NavbarColorModeToggle from '@theme/Navbar/ColorModeToggle';
import NavbarMobileSidebarToggle from '@theme/Navbar/MobileSidebar/Toggle';
import {Github} from 'lucide-react';

import styles from './styles.module.css';

type NavLink = {
  label: string;
  to?: string;
  href?: string;
  target?: string;
  /** Path used for active styling when it differs from `to`. */
  activePath?: string;
  match?: 'exact' | 'prefix';
};

const PRIMARY_CTA_TO = '/docs/getting-started';

function linkIsActive(pathname: string, link: NavLink): boolean {
  const path = link.activePath ?? link.to;
  if (!path) {
    return false;
  }
  if (link.match === 'exact') {
    return pathname === path;
  }
  if (path === '/') {
    return pathname === '/' || pathname === '';
  }
  return pathname === path || pathname.startsWith(`${path}/`);
}

function Brand(): ReactNode {
  const {siteConfig} = useDocusaurusContext();
  const logoSrc = useBaseUrl('/img/logo.svg');

  return (
    <Link className={styles.brand} to="/" aria-label={`${siteConfig.title} home`}>
      <img
        className={styles.brandLogo}
        src={logoSrc}
        alt=""
        width={28}
        height={28}
      />
      <span className={styles.brandName}>{siteConfig.title}</span>
    </Link>
  );
}

function DesktopLinks({links}: {links: NavLink[]}): ReactNode {
  const {pathname} = useLocation();

  return (
    <ul className={styles.links}>
      {links.map((link) => {
        const active = linkIsActive(pathname, link);
        const className = clsx(styles.link, active && styles.linkActive);
        if (link.href) {
          return (
            <li key={link.label}>
              <Link
                className={className}
                href={link.href}
                target={link.target}
                rel={link.target === '_blank' ? 'noopener noreferrer' : undefined}>
                {link.label}
              </Link>
            </li>
          );
        }
        return (
          <li key={link.label}>
            <Link className={className} to={link.to}>
              {link.label}
            </Link>
          </li>
        );
      })}
    </ul>
  );
}

function useSiteNav(): {navLinks: NavLink[]; githubUrl: string} {
  const {siteConfig} = useDocusaurusContext();
  const demoHref = useBaseUrl('pathname:///demo/');
  const githubUrl = `https://github.com/${siteConfig.organizationName}/${siteConfig.projectName}`;

  const navLinks: NavLink[] = [
    {
      label: 'Docs',
      to: '/docs/intro',
      activePath: '/docs',
      match: 'prefix',
    },
    {label: 'Live demo', href: demoHref, target: '_self'},
    {label: 'Roadmap', to: '/roadmap', match: 'prefix'},
  ];

  return {navLinks, githubUrl};
}

export default function NavbarContent(): ReactNode {
  const {navLinks, githubUrl} = useSiteNav();

  return (
    <div className={styles.bar}>
      <div className={styles.inner}>
        <div className={styles.left}>
          <div className={styles.mobileToggle}>
            <NavbarMobileSidebarToggle />
          </div>
          <Brand />
          <nav className={styles.desktopNav} aria-label="Primary">
            <DesktopLinks links={navLinks} />
          </nav>
        </div>

        <div className={styles.right}>
          <NavbarColorModeToggle className={styles.colorToggle} />
          <Link
            className={styles.iconButton}
            href={githubUrl}
            target="_blank"
            rel="noopener noreferrer"
            aria-label="GitHub repository">
            <Github size={18} strokeWidth={1.75} aria-hidden="true" />
          </Link>
          <Link className={styles.cta} to={PRIMARY_CTA_TO}>
            Get started
          </Link>
        </div>
      </div>
    </div>
  );
}
