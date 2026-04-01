import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'rabbitmq',
  name: 'RabbitMQ',
  icon: 'mail',
  category: 'infrastructure',
  defaults: {
    name: 'rabbitmq',
    image: 'rabbitmq:3-management-alpine',
    ports: ['5672:5672', '15672:15672'],
    environment: {
      RABBITMQ_DEFAULT_USER: 'guest',
      RABBITMQ_DEFAULT_PASS: 'guest',
    },
    volumes: ['rabbitmq-data:/var/lib/rabbitmq'],
    healthcheck: {
      test: ['CMD', 'rabbitmq-diagnostics', '-q', 'ping'],
      interval: '10s',
      timeout: '5s',
      retries: 5,
    },
    restart: 'unless-stopped',
  },
  autoEnv: [
    { key: 'AMQP_URL', template: 'amqp://{RABBITMQ_DEFAULT_USER}:{RABBITMQ_DEFAULT_PASS}@{name}:{port}' },
  ],
};

export default template;
