import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'minio',
  name: 'MinIO',
  icon: 'archive',
  category: 'infrastructure',
  defaults: {
    name: 'minio',
    image: 'minio/minio:latest',
    ports: ['9000:9000', '9001:9001'],
    environment: {
      MINIO_ROOT_USER: 'minioadmin',
      MINIO_ROOT_PASSWORD: 'minioadmin',
    },
    volumes: ['minio-data:/data'],
    command: 'server /data --console-address ":9001"',
    restart: 'unless-stopped',
  },
  autoEnv: [
    { key: 'S3_ENDPOINT', template: 'http://{name}:{port}' },
  ],
};

export default template;
