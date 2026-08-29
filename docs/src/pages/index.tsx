import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import ShowcaseGallery from '@site/src/components/ShowcaseGallery';

import styles from './index.module.css';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  const demoHref = useBaseUrl('pathname:///demo/');
  const heroSrc = useBaseUrl('/img/showcase/library-theme-split.png');

  return (
    <header className={styles.hero}>
      <div className={clsx('container', styles.heroInner)}>
        <div className={styles.heroCopy}>
          <p className={styles.kicker}>Self-hosted library server</p>
          <Heading as="h1" className={styles.title}>
            {siteConfig.title}
          </Heading>
          <p className={styles.subtitle}>{siteConfig.tagline}</p>
          <div className={styles.actions}>
            <Link
              className="button button--primary button--lg"
              to="/docs/getting-started">
              Get started
            </Link>
            <Link
              className={clsx('button button--lg', styles.ghost)}
              href={demoHref}
              target="_self">
              Try the demo
            </Link>
          </div>
          <ul className={styles.meta}>
            <li>Single static binary</li>
            <li>EPUB · PDF · Audiobooks</li>
            <li>OPDS · PWA · Multi-user</li>
          </ul>
        </div>
        <div className={styles.heroVisual}>
          <div className={styles.frame}>
            <img
              src={heroSrc}
              alt="Athenaeum library in dark and light themes"
              width={1440}
              height={900}
              loading="eager"
            />
          </div>
        </div>
      </div>
    </header>
  );
}

export default function Home(): ReactNode {
  return (
    <Layout
      title="Home"
      description="Self-hosted EPUB, PDF, and audiobook library server. Single binary, modern reader, OPDS, and optional multi-user auth.">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
        <ShowcaseGallery />
        <section className={styles.cta}>
          <div className="container">
            <Heading as="h2" className={styles.ctaTitle}>
              Ready to run your library
            </Heading>
            <p className={styles.ctaText}>
              Install with Docker, a release binary, or build from source with
              Make or plain Go and pnpm. Point Athenaeum at a folder of media and
              open the web UI.
            </p>
            <div className={styles.actions}>
              <Link
                className="button button--primary button--lg"
                to="/docs/deploying">
                Deploying guide
              </Link>
              <Link
                className={clsx('button button--lg', styles.ghost)}
                to="/docs/features">
                Full feature list
              </Link>
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
