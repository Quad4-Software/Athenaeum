import type {ReactNode} from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import Heading from '@theme/Heading';
import {EyeOff, Lock, Server, ShieldCheck} from 'lucide-react';

import styles from './privacy.module.css';

const points = [
  {
    Icon: EyeOff,
    title: 'No telemetry',
    body: 'Athenaeum does not phone home. Release binaries and container images ship with no product analytics, no usage beacons, and no crash reporting turned on.',
  },
  {
    Icon: ShieldCheck,
    title: 'No ads or trackers',
    body: 'The web UI, docs site packaging for the product, and shipped clients do not embed advertising networks or third-party tracking pixels.',
  },
  {
    Icon: Lock,
    title: 'Your library stays yours',
    body: 'Books, covers, progress, and accounts live on the host you run. The project does not operate a cloud that receives your catalog by default.',
  },
  {
    Icon: Server,
    title: 'Optional ops tools are yours',
    body: 'Prometheus metrics and Sentry or GlitchTip error reporting exist only when you configure them for your own instance. They send to endpoints you choose, not to Athenaeum.',
  },
];

export default function PrivacyPage(): ReactNode {
  return (
    <Layout
      title="Privacy"
      description="Athenaeum does not collect telemetry, show ads, or track usage. Binaries and containers ship with none of that enabled.">
      <header className={styles.hero}>
        <div className={clsx('container', styles.heroInner)}>
          <p className={styles.kicker}>Privacy</p>
          <Heading as="h1" className={styles.title}>
            We do not collect your data
          </Heading>
          <p className={styles.lead}>
            Athenaeum is self-hosted software. We do not collect telemetry,
            display ads, or track how you use the library. Binaries and
            containers ship with none of that nonsense enabled.
          </p>
        </div>
      </header>

      <main>
        <section className={styles.gridSection}>
          <div className="container">
            <div className={styles.grid}>
              {points.map(({Icon, title, body}) => (
                <article key={title} className={styles.card}>
                  <div className={styles.cardIcon} aria-hidden="true">
                    <Icon size={20} strokeWidth={1.75} />
                  </div>
                  <Heading as="h2" className={styles.cardTitle}>
                    {title}
                  </Heading>
                  <p className={styles.cardBody}>{body}</p>
                </article>
              ))}
            </div>
          </div>
        </section>

        <section className={styles.detail}>
          <div className={clsx('container', styles.detailInner)}>
            <Heading as="h2" className={styles.detailTitle}>
              What this covers
            </Heading>
            <ul className={styles.list}>
              <li>
                Official release binaries and GHCR container images do not
                include enabled telemetry, advertising, or usage tracking.
              </li>
              <li>
                The Athenaeum project does not receive browsing history, library
                contents, reading progress, or account data from default
                installs.
              </li>
              <li>
                Planned AI features (research chat, tagging, metadata help) are
                opt-in. Defaults prefer local backends such as Ollama or LM
                Studio. Cloud or OpenAI-compatible endpoints only run when you
                configure them. Prompts then go to the provider you set, not to
                Athenaeum.
              </li>
              <li>
                If you point your instance at an external identity provider,
                SMTP host, object store, or metadata API, that traffic follows
                those services&apos; policies. Athenaeum only makes the
                connections you configure.
              </li>
            </ul>
            <p className={styles.aside}>
              This page describes how Athenaeum is built and shipped. It is not
              legal advice for every jurisdiction. Your deploy may add logging
              or reverse proxies that record access on your network. That is
              under your control.
            </p>
            <div className={styles.actions}>
              <Link
                className="button button--primary button--lg"
                to="/docs/getting-started">
                Getting started
              </Link>
              <Link
                className={clsx('button button--lg', styles.ghost)}
                to="/roadmap">
                Roadmap
              </Link>
            </div>
          </div>
        </section>
      </main>
    </Layout>
  );
}
