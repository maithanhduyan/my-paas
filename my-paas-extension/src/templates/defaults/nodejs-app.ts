import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'nodejs-app',
  name: 'Node.js App',
  icon: 'code',
  category: 'app',
  defaults: {
    name: 'app',
    ports: ['3000:3000'],
    environment: {
      NODE_ENV: 'production',
    },
    volumes: [],
    build: { context: './', dockerfile: 'Dockerfile' },
    restart: 'unless-stopped',
  },
  files: [
    {
      path: 'Dockerfile',
      content: `FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./
EXPOSE 3000
CMD ["node", "dist/index.js"]
`,
    },
  ],
};

export default template;
