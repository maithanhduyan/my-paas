const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

let visits = 0;

app.get('/', (req, res) => {
  visits++;
  res.json({
    message: 'Hello from My PaaS!',
    service: 'sample-node-app',
    visits,
    uptime: process.uptime(),
    env: process.env.NODE_ENV || 'development',
    timestamp: new Date().toISOString(),
  });
});

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.listen(PORT, '0.0.0.0', () => {
  console.log(`Server running on port ${PORT}`);
});
