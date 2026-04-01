import { execFile } from 'child_process';
import * as path from 'path';

export interface BuildPlan {
  provider: string;
  language: string;
  version?: string;
  framework?: string;
  baseImage: string;
  installCmd?: string;
  buildCmd?: string;
  startCmd: string;
  ports: string[];
  envVars?: Record<string, string>;
  paths?: string[];
  caches?: string[];
  static?: boolean;
}

export interface DetectResult {
  success: boolean;
  plan?: BuildPlan;
  error?: string;
}

function getCoreBinaryPath(extensionPath: string): string {
  const platform = process.platform;
  const ext = platform === 'win32' ? '.exe' : '';
  return path.join(extensionPath, 'dist', `mypaas-core${ext}`);
}

function runCore(extensionPath: string, args: string[]): Promise<string> {
  return new Promise((resolve, reject) => {
    const bin = getCoreBinaryPath(extensionPath);
    execFile(bin, args, { timeout: 15000 }, (err, stdout, stderr) => {
      if (err) {
        reject(new Error(stderr || err.message));
        return;
      }
      resolve(stdout);
    });
  });
}

/**
 * Detect language/framework from a source directory.
 */
export async function detect(extensionPath: string, sourcePath: string): Promise<DetectResult> {
  const output = await runCore(extensionPath, ['detect', sourcePath]);
  return JSON.parse(output) as DetectResult;
}

/**
 * Generate a Dockerfile for the source directory.
 */
export async function generateDockerfile(extensionPath: string, sourcePath: string): Promise<string> {
  return runCore(extensionPath, ['dockerfile', sourcePath]);
}
