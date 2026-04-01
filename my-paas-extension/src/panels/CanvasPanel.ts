import * as vscode from 'vscode';
import { DockerManager } from '../docker/DockerManager';
import { ComposeGenerator } from '../docker/ComposeGenerator';
import { ProjectManager } from '../project/ProjectManager';
import { TemplateRegistry } from '../templates/TemplateRegistry';
import type { WebviewMessage, CanvasState, TemplateFile } from '../shared/types';
import type { BuildPlan } from '../core/CoreBridge';
import { generateDockerfile } from '../core/CoreBridge';

interface Services {
  docker: DockerManager;
  compose: ComposeGenerator;
  project: ProjectManager;
  templates: TemplateRegistry;
}

export class CanvasPanel {
  public static currentPanel: CanvasPanel | undefined;
  private static readonly viewType = 'myPaas.canvas';

  private readonly _panel: vscode.WebviewPanel;
  private readonly _context: vscode.ExtensionContext;
  private readonly _services: Services;
  private _onStateChange?: (state: CanvasState) => void;
  private _disposables: vscode.Disposable[] = [];
  private _logDisposables: Map<string, { dispose: () => void }> = new Map();

  public static createOrShow(
    context: vscode.ExtensionContext,
    services: Services,
    onStateChange?: (state: CanvasState) => void,
  ) {
    const column = vscode.window.activeTextEditor
      ? vscode.window.activeTextEditor.viewColumn
      : undefined;

    if (CanvasPanel.currentPanel) {
      CanvasPanel.currentPanel._onStateChange = onStateChange;
      CanvasPanel.currentPanel._panel.reveal(column);
      return;
    }

    const panel = vscode.window.createWebviewPanel(
      CanvasPanel.viewType,
      'My PaaS Canvas',
      column || vscode.ViewColumn.One,
      {
        enableScripts: true,
        retainContextWhenHidden: true,
        localResourceRoots: [
          vscode.Uri.joinPath(context.extensionUri, 'webview', 'dist'),
        ],
      }
    );

    CanvasPanel.currentPanel = new CanvasPanel(panel, context, services, onStateChange);
  }

  private constructor(
    panel: vscode.WebviewPanel,
    context: vscode.ExtensionContext,
    services: Services,
    onStateChange?: (state: CanvasState) => void,
  ) {
    this._panel = panel;
    this._context = context;
    this._services = services;
    this._onStateChange = onStateChange;

    this._panel.webview.html = this._getHtmlForWebview();

    this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

    this._panel.webview.onDidReceiveMessage(
      (message: WebviewMessage) => this._handleMessage(message),
      null,
      this._disposables
    );
  }

  public requestCanvasState() {
    this._panel.webview.postMessage({ type: 'canvas:requestState' });
  }

  public addTemplateToCanvas(templateId: string) {
    this._panel.webview.postMessage({ type: 'template:add', templateId });
  }

  public sendDetectResult(plan: BuildPlan) {
    this._panel.webview.postMessage({ type: 'core:detected', plan });
  }

  private async _handleMessage(message: WebviewMessage) {
    const { docker, compose, project, templates } = this._services;

    switch (message.type) {
      case 'ready': {
        const allTemplates = templates.getAll();
        this._panel.webview.postMessage({ type: 'templates:data', templates: allTemplates });
        const state = await project.load();
        this._panel.webview.postMessage({ type: 'canvas:loaded', state });
        break;
      }

      case 'canvas:save': {
        await project.save(message.state);
        this._onStateChange?.(message.state);
        this._panel.webview.postMessage({ type: 'info', message: 'Project saved.' });
        break;
      }

      case 'canvas:load': {
        const state = await project.load();
        this._panel.webview.postMessage({ type: 'canvas:loaded', state });
        break;
      }

      case 'compose:generate': {
        // Ensure service directories exist first
        await project.ensureAllServiceDirs(message.state);
        const yml = compose.generate(message.state, project.workspacePath);
        const composePath = await project.saveCompose(yml);

        // Collect auxiliary files from templates for each node
        const auxFiles: TemplateFile[] = [];
        for (const node of message.state.nodes) {
          const tpl = templates.getById(node.templateId);
          if (tpl?.files) {
            auxFiles.push(...tpl.files);
          }

          // For auto-detect nodes, generate Dockerfile via Go core
          if (node.templateId === 'auto-detect' && node.data.build) {
            try {
              const sourcePath = require('path').resolve(project.workspacePath, node.data.build.context);
              const dockerfile = await generateDockerfile(this._context.extensionPath, sourcePath);
              auxFiles.push({ path: node.data.build.dockerfile || 'Dockerfile', content: dockerfile });
            } catch (err: any) {
              this._panel.webview.postMessage({ type: 'error', message: `Auto-detect Dockerfile failed: ${err.message}` });
            }
          }
        }
        const writtenFiles = await project.saveAuxiliaryFiles(auxFiles);

        this._panel.webview.postMessage({ type: 'compose:generated', path: composePath });
        let infoMsg = `docker-compose.yml generated at ${composePath}`;
        if (writtenFiles.length > 0) {
          infoMsg += ` | Also created: ${writtenFiles.map(f => f.split(/[\\/]/).pop()).join(', ')}`;
        }
        this._panel.webview.postMessage({ type: 'info', message: infoMsg });
        break;
      }

      case 'compose:up': {
        try {
          await docker.composeUp(project.workspacePath);
          this._panel.webview.postMessage({ type: 'info', message: 'Compose Up started.' });
          this._refreshStatuses();
        } catch (err: any) {
          this._panel.webview.postMessage({ type: 'error', message: err.message });
        }
        break;
      }

      case 'compose:down': {
        try {
          await docker.composeDown(project.workspacePath);
          this._panel.webview.postMessage({ type: 'info', message: 'Compose Down completed.' });
          this._refreshStatuses();
        } catch (err: any) {
          this._panel.webview.postMessage({ type: 'error', message: err.message });
        }
        break;
      }

      case 'docker:start': {
        try {
          await docker.startService(project.workspacePath, message.serviceId);
          this._refreshStatuses();
        } catch (err: any) {
          this._panel.webview.postMessage({ type: 'error', message: err.message });
        }
        break;
      }

      case 'docker:stop': {
        try {
          await docker.stopService(project.workspacePath, message.serviceId);
          this._refreshStatuses();
        } catch (err: any) {
          this._panel.webview.postMessage({ type: 'error', message: err.message });
        }
        break;
      }

      case 'docker:restart': {
        try {
          await docker.restartService(project.workspacePath, message.serviceId);
          this._refreshStatuses();
        } catch (err: any) {
          this._panel.webview.postMessage({ type: 'error', message: err.message });
        }
        break;
      }

      case 'docker:remove': {
        try {
          await docker.removeService(project.workspacePath, message.serviceId);
          // Remove logs but keep volumes
          await project.removeServiceContainer(message.serviceId);
          this._panel.webview.postMessage({ type: 'info', message: `Service ${message.serviceId} removed. Volumes preserved.` });
          this._refreshStatuses();
        } catch (err: any) {
          this._panel.webview.postMessage({ type: 'error', message: err.message });
        }
        break;
      }

      case 'service:openDir': {
        const serviceDir = project.getServiceDir(message.serviceId);
        const uri = vscode.Uri.file(serviceDir);
        try {
          await vscode.commands.executeCommand('revealFileInOS', uri);
        } catch {
          vscode.window.showInformationMessage(`Service dir: ${serviceDir}`);
        }
        break;
      }

      case 'docker:logs': {
        this._streamLogs(message.serviceId);
        break;
      }

      case 'docker:status': {
        this._refreshStatuses();
        break;
      }

      case 'templates:list': {
        const allTemplates = templates.getAll();
        this._panel.webview.postMessage({ type: 'templates:data', templates: allTemplates });
        break;
      }
    }
  }

  private async _refreshStatuses() {
    try {
      const statuses = await this._services.docker.getStatuses(this._services.project.workspacePath);
      this._panel.webview.postMessage({ type: 'docker:statuses', data: statuses });
    } catch {
      // Docker may not be running
    }
  }

  private _streamLogs(serviceId: string) {
    // Stop existing stream for this service
    const existing = this._logDisposables.get(serviceId);
    if (existing) { existing.dispose(); }

    const { docker, project } = this._services;
    const abortController = new AbortController();
    const logDir = project.getServiceLogDir(serviceId);

    docker.streamLogs(project.workspacePath, serviceId, (line) => {
      this._panel.webview.postMessage({ type: 'docker:log', serviceId, line });
    }, abortController.signal, logDir);

    this._logDisposables.set(serviceId, {
      dispose: () => abortController.abort(),
    });
  }

  private _getHtmlForWebview(): string {
    const webview = this._panel.webview;
    const distUri = vscode.Uri.joinPath(this._context.extensionUri, 'webview', 'dist');

    const scriptUri = webview.asWebviewUri(vscode.Uri.joinPath(distUri, 'index.js'));
    const styleUri = webview.asWebviewUri(vscode.Uri.joinPath(distUri, 'index.css'));

    const nonce = getNonce();

    return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="Content-Security-Policy" content="default-src 'none'; style-src ${webview.cspSource} 'unsafe-inline'; script-src 'nonce-${nonce}'; font-src ${webview.cspSource}; img-src ${webview.cspSource} https:;">
  <link rel="stylesheet" href="${styleUri}">
  <title>My PaaS Canvas</title>
</head>
<body>
  <div id="root"></div>
  <script nonce="${nonce}" src="${scriptUri}"></script>
</body>
</html>`;
  }

  public dispose() {
    CanvasPanel.currentPanel = undefined;
    this._logDisposables.forEach(d => d.dispose());
    this._logDisposables.clear();
    this._panel.dispose();
    while (this._disposables.length) {
      const d = this._disposables.pop();
      if (d) { d.dispose(); }
    }
  }
}

function getNonce(): string {
  let text = '';
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  for (let i = 0; i < 32; i++) {
    text += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return text;
}
