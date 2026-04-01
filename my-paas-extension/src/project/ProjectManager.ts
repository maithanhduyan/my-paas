import * as fs from 'fs/promises';
import * as path from 'path';
import type { CanvasState, TemplateFile } from '../shared/types';

const PROJECT_DIR = '.mypaas';
const SERVICES_DIR = 'services';
const PROJECT_FILE = 'project.json';
const COMPOSE_FILE = 'docker-compose.yml';

export class ProjectManager {
  public readonly workspacePath: string;

  constructor(workspacePath: string) {
    this.workspacePath = workspacePath;
  }

  private get projectDir(): string {
    return path.join(this.workspacePath, PROJECT_DIR);
  }

  private get servicesDir(): string {
    return path.join(this.projectDir, SERVICES_DIR);
  }

  private get projectFile(): string {
    return path.join(this.projectDir, PROJECT_FILE);
  }

  private get composeFile(): string {
    return path.join(this.workspacePath, COMPOSE_FILE);
  }

  // --- Per-service directory management ---

  /** Get the directory for a specific service: .mypaas/services/<service-name>/ */
  getServiceDir(serviceName: string): string {
    return path.join(this.servicesDir, serviceName);
  }

  /** Get the logs directory for a service: .mypaas/services/<service-name>/logs/ */
  getServiceLogDir(serviceName: string): string {
    return path.join(this.getServiceDir(serviceName), 'logs');
  }

  /** Get the volumes directory for a service: .mypaas/services/<service-name>/volumes/ */
  getServiceVolumeDir(serviceName: string): string {
    return path.join(this.getServiceDir(serviceName), 'volumes');
  }

  /** Get a named volume path: .mypaas/services/<service-name>/volumes/<volume-name>/ */
  getVolumePath(serviceName: string, volumeName: string): string {
    return path.join(this.getServiceVolumeDir(serviceName), volumeName);
  }

  /** Ensure the service directory structure exists */
  async ensureServiceDir(serviceName: string): Promise<string> {
    const dir = this.getServiceDir(serviceName);
    await fs.mkdir(path.join(dir, 'logs'), { recursive: true });
    await fs.mkdir(path.join(dir, 'volumes'), { recursive: true });
    return dir;
  }

  /** Ensure all service directories exist for the current canvas state */
  async ensureAllServiceDirs(state: CanvasState): Promise<void> {
    for (const node of state.nodes) {
      await this.ensureServiceDir(node.data.name);
    }
    // Also ensure volume node directories
    if (state.volumeNodes) {
      for (const vol of state.volumeNodes) {
        if (vol.connectedServiceId) {
          const serviceNode = state.nodes.find(n => n.id === vol.connectedServiceId);
          if (serviceNode) {
            const volPath = this.getVolumePath(serviceNode.data.name, vol.name);
            await fs.mkdir(volPath, { recursive: true });
          }
        }
      }
    }
  }

  /** Remove a service directory (only removes the service dir, NOT volume data) */
  async removeServiceContainer(serviceName: string): Promise<void> {
    const logDir = this.getServiceLogDir(serviceName);
    try {
      await fs.rm(logDir, { recursive: true, force: true });
    } catch { /* ignore */ }
  }

  /** List all volume dirs for a service */
  async listServiceVolumes(serviceName: string): Promise<string[]> {
    const volDir = this.getServiceVolumeDir(serviceName);
    try {
      const entries = await fs.readdir(volDir, { withFileTypes: true });
      return entries.filter(e => e.isDirectory()).map(e => e.name);
    } catch {
      return [];
    }
  }

  // --- Project persistence ---

  async save(state: CanvasState): Promise<void> {
    await fs.mkdir(this.projectDir, { recursive: true });
    await fs.writeFile(this.projectFile, JSON.stringify(state, null, 2), 'utf-8');
    // Ensure service directories exist
    await this.ensureAllServiceDirs(state);
  }

  async load(): Promise<CanvasState | null> {
    try {
      const raw = await fs.readFile(this.projectFile, 'utf-8');
      return JSON.parse(raw) as CanvasState;
    } catch {
      return null;
    }
  }

  async saveCompose(content: string): Promise<string> {
    await fs.writeFile(this.composeFile, content, 'utf-8');
    return this.composeFile;
  }

  async saveAuxiliaryFiles(files: TemplateFile[]): Promise<string[]> {
    const written: string[] = [];
    for (const file of files) {
      const filePath = path.join(this.workspacePath, file.path);
      // Don't overwrite existing files (user may have customized them)
      try {
        await fs.access(filePath);
        // File exists, skip
        continue;
      } catch {
        // File doesn't exist, create it
      }
      await fs.mkdir(path.dirname(filePath), { recursive: true });
      await fs.writeFile(filePath, file.content, 'utf-8');
      written.push(filePath);
    }
    return written;
  }
}
