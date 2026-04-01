import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'mongodb',
  name: 'MongoDB',
  icon: 'database',
  category: 'database',
  defaults: {
    name: 'mongodb',
    image: 'mongo:7',
    ports: ['27017:27017'],
    environment: {
      MONGO_INITDB_ROOT_USERNAME: 'root',
      MONGO_INITDB_ROOT_PASSWORD: 'password',
    },
    volumes: ['mongodb-data:/data/db'],
    restart: 'unless-stopped',
  },
  autoEnv: [
    { key: 'MONGODB_URI', template: 'mongodb://{MONGO_INITDB_ROOT_USERNAME}:{MONGO_INITDB_ROOT_PASSWORD}@{name}:{port}' },
  ],
};

export default template;
