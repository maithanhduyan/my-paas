import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'nginx',
  name: 'Nginx',
  icon: 'globe',
  category: 'infrastructure',
  defaults: {
    name: 'nginx',
    image: 'nginx:alpine',
    ports: ['80:80', '443:443'],
    environment: {},
    volumes: ['./nginx.conf:/etc/nginx/nginx.conf:ro'],
    restart: 'unless-stopped',
  },
  files: [
    {
      path: 'nginx.conf',
      content: `events {
    worker_connections 1024;
}

http {
    include       mime.types;
    sendfile      on;

    upstream app_server {
        server app:3000;
    }

    server {
        listen 80;
        server_name localhost;

        location / {
            proxy_pass http://app_server;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
`,
    },
  ],
};

export default template;
