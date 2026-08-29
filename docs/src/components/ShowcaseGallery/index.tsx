import type {ReactNode} from 'react';
import useBaseUrl from '@docusaurus/useBaseUrl';
import Heading from '@theme/Heading';

import styles from './styles.module.css';

type Shot = {
  src: string;
  alt: string;
  caption: string;
  wide?: boolean;
};

const shots: Shot[] = [
  {
    src: '/img/showcase/library-light.png',
    alt: 'Library grid in light theme',
    caption: 'Library · light',
    wide: true,
  },
  {
    src: '/img/showcase/book-detail-dark.png',
    alt: 'Book detail page in dark theme',
    caption: 'Book detail · dark',
  },
  {
    src: '/img/showcase/settings-light.png',
    alt: 'Settings page in light theme',
    caption: 'Settings · light',
  },
  {
    src: '/img/showcase/library-mobile-dark.png',
    alt: 'Mobile library in dark theme',
    caption: 'Mobile · dark',
  },
  {
    src: '/img/showcase/library-mobile-light.png',
    alt: 'Mobile library in light theme',
    caption: 'Mobile · light',
  },
];

function ShotCard({src, alt, caption, wide}: Shot) {
  const url = useBaseUrl(src);
  return (
    <figure className={wide ? styles.wide : styles.shot}>
      <div className={styles.frame}>
        <img src={url} alt={alt} loading="lazy" width={1440} height={900} />
      </div>
      <figcaption>{caption}</figcaption>
    </figure>
  );
}

export default function ShowcaseGallery(): ReactNode {
  return (
    <section className={styles.gallery}>
      <div className="container">
        <div className={styles.head}>
          <Heading as="h2" className={styles.title}>
            Product tour
          </Heading>
          <p className={styles.lead}>
            Light and dark themes, desktop and mobile. Screenshots are captured
            from the offline demo with <code>task showcase</code>.
          </p>
        </div>
        <div className={styles.grid}>
          {shots.map((shot) => (
            <ShotCard key={shot.src} {...shot} />
          ))}
        </div>
      </div>
    </section>
  );
}
