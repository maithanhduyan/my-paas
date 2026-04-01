import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'adminer',
  name: 'Adminer',
  icon: 'tools',
  category: 'tools',
  defaults: {
    name: 'adminer',
    image: 'adminer:latest',
    ports: ['8080:8080'],
    environment: {},
    volumes: [],
    restart: 'unless-stopped',
  },
};

export default template;
