import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import {
  AudioLines,
  Network,
  Podcast,
  Puzzle,
  Radio,
  RefreshCw,
  Rss,
  Share2,
  type LucideIcon,
} from 'lucide-react';

import styles from './roadmap.module.css';

type Status = 'planned';

type RoadmapItem = {
  id: string;
  title: string;
  summary: string;
  detail?: string;
  Icon: LucideIcon;
  status: Status;
};

const reticulumBullets = [
  {
    title: 'Peer library links',
    body: 'Connect Athenaeum instances over Reticulum and pull or push titles without a central relay you do not control.',
  },
  {
    title: 'Encrypted transfers',
    body: 'Move book files and catalogs end to end with Reticulum identities. Paths can use internet, radio, or other RNS interfaces.',
  },
  {
    title: 'Mesh-friendly announce',
    body: 'Optional destination announce so peers can find a shareable catalog on the mesh when you choose to publish one.',
  },
  {
    title: 'Go stack',
    body: (
      <>
        Built on{' '}
        <a
          href="https://reticulum-go.quad4.io/"
          target="_blank"
          rel="noopener noreferrer">
          Reticulum-Go
        </a>
        , wire-compatible with the Python Reticulum reference.
      </>
    ),
  },
];

const listeningItems: RoadmapItem[] = [
  {
    id: 'podcasts',
    title: 'Podcasts',
    summary:
      'Subscribe, download episodes, and keep them in the library beside books and audiobooks.',
    detail: 'Transcription via Whisper or a similar local or sidecar model.',
    Icon: Podcast,
    status: 'planned',
  },
  {
    id: 'rss',
    title: 'RSS reader',
    summary:
      'In-app feed reading and aggregation, with FreshRSS as an optional backend.',
    Icon: Rss,
    status: 'planned',
  },
  {
    id: 'audiobookshelf',
    title: 'Audiobookshelf bridge',
    summary:
      'Talk to an existing Audiobookshelf server for podcasts or audiobooks instead of re-importing everything.',
    Icon: AudioLines,
    status: 'planned',
  },
];

const platformItems: RoadmapItem[] = [
  {
    id: 'extensions',
    title: 'Extensions',
    summary:
      'Load add-ons as WASM modules or native Go plugins for scrapers, importers, and custom jobs.',
    Icon: Puzzle,
    status: 'planned',
  },
  {
    id: 'updates',
    title: 'Update system',
    summary:
      'Check for releases from the UI or CLI, show notes, and apply or stage binary upgrades safely.',
    Icon: RefreshCw,
    status: 'planned',
  },
];

function StatusChip({status}: {status: Status}) {
  return (
    <span className={styles.status} data-status={status}>
      {status === 'planned' ? 'Planned' : status}
    </span>
  );
}

function TrackCard({item}: {item: RoadmapItem}) {
  const {Icon} = item;
  return (
    <article className={styles.card}>
      <div className={styles.cardTop}>
        <div className={styles.cardIcon} aria-hidden="true">
          <Icon size={20} strokeWidth={1.75} />
        </div>
        <StatusChip status={item.status} />
      </div>
      <Heading as="h3" className={styles.cardTitle}>
        {item.title}
      </Heading>
      <p className={styles.cardBody}>{item.summary}</p>
      {item.detail ? <p className={styles.cardDetail}>{item.detail}</p> : null}
    </article>
  );
}

export default function RoadmapPage(): ReactNode {
  return (
    <Layout
      title="Roadmap"
      description="Planned Athenaeum work: Reticulum peer sharing, podcasts, RSS, extensions, and updates.">
      <header className={styles.hero}>
        <div className={clsx('container', styles.heroInner)}>
          <p className={styles.kicker}>Project direction</p>
          <Heading as="h1" className={styles.title}>
            Roadmap
          </Heading>
          <p className={styles.lead}>
            What we intend to build next. Order can shift. Nothing here has a
            ship date. Existing installs keep working while these land.
          </p>
          <ul className={styles.meta}>
            <li>No promised release dates</li>
            <li>Priorities move with real use</li>
            <li>
              Track discussion on{' '}
              <a
                href="https://github.com/Quad4-Software/Athenaeum"
                target="_blank"
                rel="noopener noreferrer">
                GitHub
              </a>
            </li>
          </ul>
        </div>
      </header>

      <main>
        <section className={styles.spotlight} aria-labelledby="reticulum-heading">
          <div className="container">
            <div className={styles.spotlightGrid}>
              <div className={styles.spotlightCopy}>
                <div className={styles.spotlightBadge}>
                  <Network size={16} strokeWidth={2} aria-hidden="true" />
                  <span>Decentralization</span>
                  <StatusChip status="planned" />
                </div>
                <Heading as="h2" id="reticulum-heading" className={styles.spotlightTitle}>
                  Sharing over Reticulum
                </Heading>
                <p className={styles.spotlightLead}>
                  Let households and friends exchange library content over the
                  Reticulum Network Stack instead of only HTTPS, OPDS, or email
                  share links. Athenaeum stays the reader and catalog. Reticulum
                  carries the peer path.
                </p>
                <p className={styles.spotlightAside}>
                  <Share2 size={16} strokeWidth={1.75} aria-hidden="true" />
                  Complements OPDS and public download links. Does not replace
                  them.
                </p>
              </div>
              <ul className={styles.spotlightList}>
                {reticulumBullets.map((item) => (
                  <li key={item.title}>
                    <Heading as="h3">{item.title}</Heading>
                    <p>{item.body}</p>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </section>

        <section className={styles.track} aria-labelledby="listening-heading">
          <div className="container">
            <div className={styles.trackHead}>
              <div className={styles.trackIcon} aria-hidden="true">
                <Radio size={22} strokeWidth={1.75} />
              </div>
              <div>
                <Heading as="h2" id="listening-heading" className={styles.trackTitle}>
                  Listening and feeds
                </Heading>
                <p className={styles.trackLead}>
                  Podcasts, transcription, RSS, and bridges to tools you already
                  run.
                </p>
              </div>
            </div>
            <div className={styles.cardGrid}>
              {listeningItems.map((item) => (
                <TrackCard key={item.id} item={item} />
              ))}
            </div>
          </div>
        </section>

        <section
          className={clsx(styles.track, styles.trackAlt)}
          aria-labelledby="platform-heading">
          <div className="container">
            <div className={styles.trackHead}>
              <div className={styles.trackIcon} aria-hidden="true">
                <Puzzle size={22} strokeWidth={1.75} />
              </div>
              <div>
                <Heading as="h2" id="platform-heading" className={styles.trackTitle}>
                  Platform
                </Heading>
                <p className={styles.trackLead}>
                  Hooks for third-party code and a saner path to stay current.
                </p>
              </div>
            </div>
            <div className={styles.cardGridWide}>
              {platformItems.map((item) => (
                <TrackCard key={item.id} item={item} />
              ))}
            </div>
          </div>
        </section>

        <section className={styles.cta}>
          <div className="container">
            <Heading as="h2" className={styles.ctaTitle}>
              Run what ships today
            </Heading>
            <p className={styles.ctaText}>
              EPUB, PDF, comics, audiobooks, Kokoro narration, OPDS, and
              multi-user auth are already in the binary. The roadmap is for what
              comes after.
            </p>
            <div className={styles.actions}>
              <Link
                className="button button--primary button--lg"
                to="/docs/getting-started">
                Getting started
              </Link>
              <Link
                className={clsx('button button--lg', styles.ghost)}
                to="/docs/features">
                Features
              </Link>
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
