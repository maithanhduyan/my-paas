import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'mysql',
  name: 'MySQL',
  icon: 'database',
  category: 'database',
  defaults: {
    name: 'mysql',
    image: 'mysql:8',
    ports: ['3306:3306'],
    environment: {
      MYSQL_ROOT_PASSWORD: 'password',
      MYSQL_DATABASE: 'mydb',
      MYSQL_USER: 'user',
      MYSQL_PASSWORD: 'password',
    },
    volumes: ['mysql-data:/var/lib/mysql'],
    healthcheck: {
      test: ['CMD', 'mysqladmin', 'ping', '-h', 'localhost'],
      interval: '10s',
      timeout: '5s',
      retries: 5,
    },
    restart: 'unless-stopped',
  },
  autoEnv: [
    { key: 'DATABASE_URL', template: 'mysql://{MYSQL_USER}:{MYSQL_PASSWORD}@{name}:{port}/{MYSQL_DATABASE}' },
  ],
};

export default template;
