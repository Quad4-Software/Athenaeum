import {useMemo, useState, type ReactNode} from 'react';
import clsx from 'clsx';
import useBaseUrl from '@docusaurus/useBaseUrl';
import Heading from '@theme/Heading';

import styles from './styles.module.css';

type Shot = {
  id: string;
  src: string;
  alt: string;
  label: string;
  group: 'desktop' | 'mobile';
};

const shots: Shot[] = [
  {
    id: 'split',
    src: '/img/showcase/library-theme-split.png',
    alt: 'Library grid split between dark and light themes',
    label: 'Library themes',
    group: 'desktop',
  },
  {
    id: 'library-dark',
    src: '/img/showcase/library-dark.png',
    alt: 'Library grid in dark theme',
    label: 'Library · dark',
    group: 'desktop',
  },
  {
    id: 'library-light',
    src: '/img/showcase/library-light.png',
    alt: 'Library grid in light theme',
    label: 'Library · light',
    group: 'desktop',
  },
  {
    id: 'book-dark',
    src: '/img/showcase/book-detail-dark.png',
    alt: 'Book detail page in dark theme',
    label: 'Book detail · dark',
    group: 'desktop',
  },
  {
    id: 'book-light',
    src: '/img/showcase/book-detail-light.png',
    alt: 'Book detail page in light theme',
    label: 'Book detail · light',
    group: 'desktop',
  },
  {
    id: 'settings-dark',
    src: '/img/showcase/settings-dark.png',
    alt: 'Settings page in dark theme',
    label: 'Settings · dark',
    group: 'desktop',
  },
  {
    id: 'settings-light',
    src: '/img/showcase/settings-light.png',
    alt: 'Settings page in light theme',
    label: 'Settings · light',
    group: 'desktop',
  },
  {
    id: 'mobile-dark',
    src: '/img/showcase/library-mobile-dark.png',
    alt: 'Mobile library in dark theme',
    label: 'Mobile · dark',
    group: 'mobile',
  },
  {
    id: 'mobile-light',
    src: '/img/showcase/library-mobile-light.png',
    alt: 'Mobile library in light theme',
    label: 'Mobile · light',
    group: 'mobile',
  },
];

type Filter = 'all' | 'desktop' | 'mobile';

export default function ShowcaseGallery(): ReactNode {
  const baseUrl = useBaseUrl('/');
  const [filter, setFilter] = useState<Filter>('all');
  const [activeId, setActiveId] = useState(shots[0].id);

  const resolved = useMemo(
    () =>
      shots.map((shot) => ({
        ...shot,
        url: `${baseUrl.replace(/\/$/, '')}${shot.src}`,
      })),
    [baseUrl],
  );

  const visible = resolved.filter(
    (shot) => filter === 'all' || shot.group === filter,
  );
  const active =
    visible.find((shot) => shot.id === activeId) ?? visible[0] ?? resolved[0];

  return (
    <section className={styles.gallery}>
      <div className="container">
        <div className={styles.head}>
          <Heading as="h2" className={styles.title}>
            Screenshots
          </Heading>
          <p className={styles.lead}>
            Browse the library, book detail, settings, and mobile shell in light
            and dark themes.
          </p>
          <div className={styles.filters} role="tablist" aria-label="Screenshot filter">
            {(
              [
                ['all', 'All'],
                ['desktop', 'Desktop'],
                ['mobile', 'Mobile'],
              ] as const
            ).map(([id, label]) => (
              <button
                key={id}
                type="button"
                role="tab"
                aria-selected={filter === id}
                className={clsx(styles.filter, filter === id && styles.filterActive)}
                onClick={() => {
                  setFilter(id);
                  const next = resolved.find(
                    (shot) => id === 'all' || shot.group === id,
                  );
                  if (next) setActiveId(next.id);
                }}>
                {label}
              </button>
            ))}
          </div>
        </div>

        <div className={styles.stage}>
          <figure className={styles.feature}>
            <div
              className={clsx(
                styles.featureFrame,
                active.group === 'mobile' && styles.featureMobile,
              )}>
              <img
                key={active.id}
                src={active.url}
                alt={active.alt}
                width={active.group === 'mobile' ? 390 : 1440}
                height={active.group === 'mobile' ? 844 : 900}
                loading="eager"
              />
            </div>
            <figcaption>{active.label}</figcaption>
          </figure>

          <div className={styles.thumbs} role="listbox" aria-label="Screenshot thumbnails">
            {visible.map((shot) => (
              <button
                key={shot.id}
                type="button"
                role="option"
                aria-selected={shot.id === active.id}
                className={clsx(
                  styles.thumb,
                  shot.group === 'mobile' && styles.thumbMobile,
                  shot.id === active.id && styles.thumbActive,
                )}
                onClick={() => setActiveId(shot.id)}>
                <img src={shot.url} alt="" loading="lazy" />
                <span>{shot.label}</span>
              </button>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}
