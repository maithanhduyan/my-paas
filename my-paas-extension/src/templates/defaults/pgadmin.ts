import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'pgadmin',
  name: 'pgAdmin',
  icon: 'tools',
  category: 'tools',
  defaults: {
    name: 'pgadmin',
    image: 'dpage/pgadmin4:latest',
    ports: ['5050:80'],
    environment: {
      PGADMIN_DEFAULT_EMAIL: 'admin@admin.com',
      PGADMIN_DEFAULT_PASSWORD: 'admin',
    },
    volumes: ['pgadmin-data:/var/lib/pgadmin'],
    restart: 'unless-stopped',
  },
};

export default template;
