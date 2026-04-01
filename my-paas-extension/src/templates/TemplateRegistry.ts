import * as vscode from 'vscode';
import type { ServiceTemplate } from '../shared/types';

import postgres from './defaults/postgres';
import redis from './defaults/redis';
import mongodb from './defaults/mongodb';
import mysql from './defaults/mysql';
import nginx from './defaults/nginx';
import nodejsApp from './defaults/nodejs-app';
import goApp from './defaults/go-app';
import pythonApp from './defaults/python-app';
import minio from './defaults/minio';
import pgadmin from './defaults/pgadmin';
import adminer from './defaults/adminer';
import traefik from './defaults/traefik';
import rabbitmq from './defaults/rabbitmq';
import autoDetect from './defaults/auto-detect';

const builtinTemplates: ServiceTemplate[] = [
  autoDetect,
  postgres,
  redis,
  mongodb,
  mysql,
  nginx,
  nodejsApp,
  goApp,
  pythonApp,
  minio,
  pgadmin,
  adminer,
  traefik,
  rabbitmq,
];

export class TemplateRegistry {
  private templates: ServiceTemplate[];

  constructor(_context: vscode.ExtensionContext) {
    this.templates = [...builtinTemplates];
  }

  getAll(): ServiceTemplate[] {
    return this.templates;
  }

  getById(id: string): ServiceTemplate | undefined {
    return this.templates.find(t => t.id === id);
  }
}
