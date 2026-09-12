import type {ReactNode} from 'react';
import clsx from 'clsx';
import Heading from '@theme/Heading';
import {
  BookOpen,
  FolderSearch,
  MonitorSmartphone,
  Network,
  Package,
  Podcast,
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
  soon?: boolean;
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
  {
    title: 'Podcasts',
    Icon: Podcast,
    soon: true,
    description: (
      <>
        Subscribe to feeds, download episodes, and keep them beside books and
        audiobooks in the library.
      </>
    ),
  },
  {
    title: 'Desktop and mobile apps',
    Icon: MonitorSmartphone,
    soon: true,
    description: (
      <>
        Installable apps for Windows, macOS, Linux, iOS, and Android, running
        the same library and UI.
      </>
    ),
  },
  {
    title: 'Sharing over Reticulum',
    Icon: Network,
    soon: true,
    description: (
      <>
        Exchange titles with other Athenaeum instances over the Reticulum
        network. No central relay.
      </>
    ),
  },
];

function Feature({title, description, Icon, soon}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className={clsx(styles.card, soon && styles.cardSoon)}>
        <div className={styles.cardTop}>
          <div className={styles.cardIcon} aria-hidden="true">
            <Icon size={22} strokeWidth={1.75} />
          </div>
          {soon ? <span className={styles.soonBadge}>Soon</span> : null}
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
            Browser readers, Kokoro narration, OPDS, and the usual self-host
            knobs.
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
