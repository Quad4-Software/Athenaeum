import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import {
  BookOpen,
  FolderSearch,
  Package,
  Shield,
  Smartphone,
  Volume2,
  type LucideIcon,
} from 'lucide-react';

import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  description: ReactNode;
  Icon: LucideIcon;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Easy to run',
    Icon: Package,
    description: (
      <>
        One static binary or a Docker Compose file. Point it at your books
        folder and open the web UI. No Node at runtime.
      </>
    ),
  },
  {
    title: 'Read in the browser',
    Icon: BookOpen,
    description: (
      <>
        EPUB, PDF, MOBI/AZW/AZW3, comics (CBZ/CBR), and audiobooks play in the
        browser. Multi-file audiobook folders merge into one book.
      </>
    ),
  },
  {
    title: 'Narration with Kokoro',
    Icon: Volume2,
    description: (
      <>
        Listen to EPUBs with in-browser Kokoro TTS, or use your browser voice.
        An optional Kokoro sidecar is available for server-side narration.
      </>
    ),
  },
  {
    title: 'Search and shelves',
    Icon: FolderSearch,
    description: (
      <>
        Full-text search, manual and smart shelves, continue reading, tags, and
        ratings. Progress stays per user when auth is on.
      </>
    ),
  },
  {
    title: 'Works with e-readers',
    Icon: Smartphone,
    description: (
      <>
        OPDS catalogs for KOReader and similar apps, KOSync progress sync, share
        links, and optional send-to-Kindle over email.
      </>
    ),
  },
  {
    title: 'Household and ops',
    Icon: Shield,
    description: (
      <>
        Optional multi-user auth, guests, and SSO. Backups, metrics, sandboxing,
        and release binaries for Linux, macOS, Windows, and BSD.
      </>
    ),
  },
];

function Feature({title, description, Icon}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className={styles.card}>
        <div className={styles.cardIcon} aria-hidden="true">
          <Icon size={22} strokeWidth={1.75} />
        </div>
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
            What you get
          </Heading>
          <p className={styles.sectionLead}>
            A private library on your own machine: browser readers, Kokoro
            narration, OPDS for e-readers, and the usual self-host knobs.
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
