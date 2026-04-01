import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'auto-detect',
  name: 'Auto Detect',
  icon: 'search',
  category: 'app',
  defaults: {
    name: 'app',
    ports: ['3000:3000'],
    environment: {},
    volumes: [],
    build: { context: './', dockerfile: 'Dockerfile' },
    restart: 'unless-stopped',
  },
};

export default template;
