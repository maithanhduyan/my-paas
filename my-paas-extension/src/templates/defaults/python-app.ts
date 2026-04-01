import type { ServiceTemplate } from '../../shared/types';

const template: ServiceTemplate = {
  id: 'python-app',
  name: 'Python App',
  icon: 'code',
  category: 'app',
  defaults: {
    name: 'python-app',
    ports: ['8000:8000'],
    environment: {},
    volumes: [],
    build: { context: './', dockerfile: 'Dockerfile' },
    restart: 'unless-stopped',
  },
  files: [
    {
      path: 'Dockerfile',
      content: `FROM python:3.12-slim
WORKDIR /app
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY . .
EXPOSE 8000
CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
`,
    },
  ],
};

export default template;
