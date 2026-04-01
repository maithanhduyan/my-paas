import { exec } from 'child_process';
import { promisify } from 'util';
import * as fs from 'fs';
import * as path from 'path';
import type { ContainerStatus } from '../shared/types';

const execAsync = promisify(exec);

export class DockerManager {

  async composeUp(workspacePath: string): Promise<void> {
    await execAsync('docker compose up -d', { cwd: workspacePath });
  }

  async composeDown(workspacePath: string): Promise<void> {
    await execAsync('docker compose down', { cwd: workspacePath });
  }

  async startService(workspacePath: string, serviceName: string): Promise<void> {
    await execAsync(`docker compose start ${serviceName}`, { cwd: workspacePath });
  }

  async stopService(workspacePath: string, serviceName: string): Promise<void> {
    await execAsync(`docker compose stop ${serviceName}`, { cwd: workspacePath });
  }

  async restartService(workspacePath: string, serviceName: string): Promise<void> {
    await execAsync(`docker compose restart ${serviceName}`, { cwd: workspacePath });
  }

  async removeService(workspacePath: string, serviceName: string): Promise<void> {
    await execAsync(`docker compose rm -f -s ${serviceName}`, { cwd: workspacePath });
  }

  async getStatuses(workspacePath: string): Promise<ContainerStatus[]> {
    try {
      const { stdout } = await execAsync(
        'docker compose ps --format json',
        { cwd: workspacePath }
      );

      if (!stdout.trim()) { return []; }

      const lines = stdout.trim().split('\n');
      const statuses: ContainerStatus[] = [];

      for (const line of lines) {
        try {
          const container = JSON.parse(line);
          const state = container.State === 'running' ? 'running'
            : container.State === 'exited' ? 'stopped'
            : 'error';

          statuses.push({
            serviceId: container.Service || container.Name,
            state,
            containerId: container.ID,
            ports: container.Publishers?.map((p: any) =>
              `${p.PublishedPort}:${p.TargetPort}`
            ) || [],
          });
        } catch {
          // skip malformed lines
        }
      }

      return statuses;
    } catch {
      return [];
    }
  }

  streamLogs(
    workspacePath: string,
    serviceName: string,
    onLine: (line: string) => void,
    signal: AbortSignal,
    logDir?: string,
  ): void {
    const { spawn } = require('child_process');
    const proc = spawn('docker', ['compose', 'logs', '-f', '--tail', '100', serviceName], {
      cwd: workspacePath,
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    // Open a log file stream if a log directory is provided
    let logStream: fs.WriteStream | null = null;
    if (logDir) {
      try {
        fs.mkdirSync(logDir, { recursive: true });
        const logFile = path.join(logDir, `${serviceName}.log`);
        logStream = fs.createWriteStream(logFile, { flags: 'a' });
      } catch { /* ignore */ }
    }

    const handleData = (data: Buffer) => {
      const lines = data.toString().split('\n');
      for (const line of lines) {
        if (line.trim()) {
          onLine(line);
          logStream?.write(line + '\n');
        }
      }
    };

    proc.stdout.on('data', handleData);
    proc.stderr.on('data', handleData);

    signal.addEventListener('abort', () => {
      proc.kill();
      logStream?.end();
    });

    proc.on('error', () => { logStream?.end(); });
    proc.on('exit', () => { logStream?.end(); });
  }
}
