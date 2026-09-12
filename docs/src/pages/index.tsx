import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import CodeBlock from '@theme/CodeBlock';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import ShowcaseGallery from '@site/src/components/ShowcaseGallery';

import styles from './index.module.css';

const FORMAT_CHIPS = ['EPUB', 'PDF', 'MOBI / AZW3', 'CBZ / CBR', 'MP3 / M4B / FLAC'];

const FACTS = ['One binary or one container', 'MIT licensed', 'No telemetry'];

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  const demoHref = useBaseUrl('pathname:///demo/');
  const heroSrc = useBaseUrl('/img/showcase/library-theme-split.png');

  return (
    <header className={styles.hero}>
      <div className={clsx('container', styles.heroInner)}>
        <div className={styles.heroCopy}>
          <p className={styles.kicker}>Self-hosted library</p>
          <Heading as="h1" className={styles.title}>
            {siteConfig.title}
          </Heading>
          <p className={styles.subtitle}>{siteConfig.tagline}</p>
          <ul className={styles.chips} aria-label="Supported formats">
            {FORMAT_CHIPS.map((chip) => (
              <li key={chip}>{chip}</li>
            ))}
          </ul>
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
          <ul className={styles.facts}>
            {FACTS.map((fact) => (
              <li key={fact}>{fact}</li>
            ))}
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
      description="Self-hosted library for EPUB, PDF, comics, and audiobooks.">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
        <ShowcaseGallery />
        <section className={styles.cta}>
          <div className="container">
            <Heading as="h2" className={styles.ctaTitle}>
              Install Athenaeum
            </Heading>
            <p className={styles.ctaText}>
              Use Docker Compose, a release binary, or{' '}
              <code>./install.sh</code>. Point it at a folder of books and open
              the web UI. Build from source when you want to hack on it.
            </p>
            <div className={styles.ctaCode}>
              <CodeBlock language="bash">
                {`docker run -d --name athenaeum \\
  -p 8080:8080 \\
  -v athenaeum-data:/data \\
  -v /path/to/books:/library \\
  ghcr.io/quad4-software/athenaeum:latest`}
              </CodeBlock>
            </div>
            <div className={styles.actions}>
              <Link
                className="button button--primary button--lg"
                to="/docs/getting-started">
                Getting started
              </Link>
              <Link
                className={clsx('button button--lg', styles.ghost)}
                to="/docs/deploying">
                Deploying
              </Link>
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
