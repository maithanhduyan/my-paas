import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'traefik',
  name: 'Traefik',
  icon: 'globe',
  category: 'infrastructure',
  defaults: {
    name: 'traefik',
    image: 'traefik:v3.0',
    ports: ['80:80', '443:443', '8080:8080'],
    environment: {},
    volumes: [
      '/var/run/docker.sock:/var/run/docker.sock:ro',
    ],
    command: '--api.insecure=true --providers.docker=true --entrypoints.web.address=:80 --entrypoints.websecure.address=:443',
    restart: 'unless-stopped',
  },
};

export default template;
