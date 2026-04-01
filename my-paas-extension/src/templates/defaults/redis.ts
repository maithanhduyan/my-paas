import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'redis',
  name: 'Redis',
  icon: 'redis',
  category: 'database',
  defaults: {
    name: 'redis',
    image: 'redis:7-alpine',
    ports: ['6379:6379'],
    environment: {},
    volumes: ['redis-data:/data'],
    healthcheck: {
      test: ['CMD', 'redis-cli', 'ping'],
      interval: '10s',
      timeout: '5s',
      retries: 5,
    },
    restart: 'unless-stopped',
  },
  autoEnv: [
    { key: 'REDIS_URL', template: 'redis://{name}:{port}' },
  ],
};

export default template;
