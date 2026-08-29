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
  const heroSrc = useBaseUrl('/img/showcase/library-dark.png');

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
              className={clsx('button button--secondary button--lg', styles.ghost)}
              href={demoHref}
              target="_self">
              Try the demo
            </Link>
          </div>
          <ul className={styles.meta}>
            <li>Single static binary</li>
            <li>EPUB · PDF · audiobooks</li>
            <li>OPDS · PWA · multi-user</li>
          </ul>
        </div>
        <div className={styles.heroVisual}>
          <div className={styles.frame}>
            <img
              src={heroSrc}
              alt="Athenaeum library view in dark theme"
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
      title="Docs"
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
              Install with Docker, a release binary, or build from source. Point
              Athenaeum at a folder of media and open the web UI.
            </p>
            <div className={styles.actions}>
              <Link
                className="button button--primary button--lg"
                to="/docs/deploying">
                Deploying guide
              </Link>
              <Link
                className="button button--secondary button--lg"
                to="/docs/configuration">
                Configuration reference
              </Link>
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
