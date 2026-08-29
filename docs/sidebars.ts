import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'intro',
    {
      type: 'category',
      label: 'Start here',
      collapsed: false,
      items: ['getting-started', 'features', 'deploying'],
    },
    {
      type: 'category',
      label: 'Operate',
      items: [
        'configuration',
        'authentication',
        'cli-users',
        'library',
        'catalogs',
        'operations',
        'http-api',
      ],
    },
    {
      type: 'category',
      label: 'Contribute',
      items: ['development', 'project-layout'],
    },
  ],
};

export default sidebars;
