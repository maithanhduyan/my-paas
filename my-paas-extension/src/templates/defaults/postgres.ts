import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'postgres',
  name: 'PostgreSQL',
  icon: 'database',
  category: 'database',
  defaults: {
    name: 'postgres',
    image: 'postgres:16-alpine',
    ports: ['5432:5432'],
    environment: {
      POSTGRES_USER: 'postgres',
      POSTGRES_PASSWORD: 'postgres',
      POSTGRES_DB: 'mydb',
    },
    volumes: ['postgres-data:/var/lib/postgresql/data'],
    healthcheck: {
      test: ['CMD-SHELL', 'pg_isready -U postgres'],
      interval: '10s',
      timeout: '5s',
      retries: 5,
    },
    restart: 'unless-stopped',
  },
  autoEnv: [
    {
      key: 'DATABASE_URL',
      template: 'postgresql://{POSTGRES_USER}:{POSTGRES_PASSWORD}@{name}:{port}/{POSTGRES_DB}',
    },
  ],
};

export default template;
