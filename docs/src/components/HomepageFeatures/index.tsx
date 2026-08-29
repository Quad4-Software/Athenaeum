import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';

import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  description: ReactNode;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'One binary',
    description: (
      <>
        Go server plus embedded Svelte UI. No CDN, no Node at runtime. Optional
        <code> --web-dir</code> if you prefer serving a built SPA from disk.
      </>
    ),
  },
  {
    title: 'Readers that feel native',
    description: (
      <>
        EPUB and PDF in the browser, audiobooks with range streaming, comics
        with dual-page and RTL modes, plus optional EPUB narration.
      </>
    ),
  },
  {
    title: 'Built for real libraries',
    description: (
      <>
        Concurrent scanning, SQLite FTS5 search, shelves, progress, OPDS for
        e-readers, PWA install, and optional multi-user auth with guests and
        OIDC.
      </>
    ),
  },
  {
    title: 'Formats that matter',
    description: (
      <>
        EPUB, PDF, MOBI/AZW/AZW3 in-browser, KFX download, CBZ/CBR comics, and
        multi-file audiobook folders merged automatically.
      </>
    ),
  },
  {
    title: 'Catalogs and sync',
    description: (
      <>
        OPDS 1.2 and OPDS 2 for e-reader apps, KOSync progress sync, share
        links, and optional send-to-Kindle over SMTP.
      </>
    ),
  },
  {
    title: 'Ops-ready self-hosting',
    description: (
      <>
        Docker and host installers, Prometheus metrics, backup/restore,
        sandboxing, and release binaries across Linux, macOS, Windows, and BSD.
      </>
    ),
  },
];

function Feature({title, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className={styles.card}>
        <Heading as="h3" className={styles.cardTitle}>
          {title}
        </Heading>
        <p className={styles.cardBody}>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactNode {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className={styles.sectionHead}>
          <Heading as="h2" className={styles.sectionTitle}>
            Everything you need on one shelf
          </Heading>
          <p className={styles.sectionLead}>
            Fast indexing, a clean reader, catalogs for e-readers, and the ops
            knobs you need when you self-host.
          </p>
        </div>
        <div className="row">
          {FeatureList.map((props) => (
            <Feature key={props.title} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
