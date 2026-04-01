import * as vscode from 'vscode';
import { CanvasPanel } from './panels/CanvasPanel';
import { DockerManager } from './docker/DockerManager';
import { ComposeGenerator } from './docker/ComposeGenerator';
import { ProjectManager } from './project/ProjectManager';
import { TemplateRegistry } from './templates/TemplateRegistry';
import { ServicesTreeProvider, TemplatesTreeProvider } from './sidebar/SidebarProviders';
import { detect } from './core/CoreBridge';

export function activate(context: vscode.ExtensionContext) {
  const docker = new DockerManager();
  const compose = new ComposeGenerator();
  const templates = new TemplateRegistry(context);

  // Sidebar tree views
  const servicesTree = new ServicesTreeProvider();
  const templatesTree = new TemplatesTreeProvider(templates);

  vscode.window.registerTreeDataProvider('myPaas.services', servicesTree);
  vscode.window.registerTreeDataProvider('myPaas.templates', templatesTree);

  // Load services from saved project on startup
  const loadServicesFromProject = async () => {
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) { return; }
    const project = new ProjectManager(workspaceFolder.uri.fsPath);
    const state = await project.load();
    if (state) {
      servicesTree.refresh(state.nodes.map(n => ({
        name: n.data.name,
        image: n.data.image,
        ports: n.data.ports,
      })));
    }
  };
  loadServicesFromProject();

  // Expose servicesTree so CanvasPanel can update it
  const updateServicesFromCanvas = (state: import('./shared/types').CanvasState) => {
    servicesTree.refresh(state.nodes.map(n => ({
      name: n.data.name,
      image: n.data.image,
      ports: n.data.ports,
    })));
  };

  context.subscriptions.push(
    vscode.commands.registerCommand('myPaas.openCanvas', () => {
      const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
      if (!workspaceFolder) {
        vscode.window.showErrorMessage('Please open a workspace folder first.');
        return;
      }
      const project = new ProjectManager(workspaceFolder.uri.fsPath);
      CanvasPanel.createOrShow(context, { docker, compose, project, templates }, updateServicesFromCanvas);
    }),

    vscode.commands.registerCommand('myPaas.addTemplate', (templateId: string) => {
      const panel = CanvasPanel.currentPanel;
      if (panel) {
        panel.addTemplateToCanvas(templateId);
      } else {
        // Open canvas first, then add
        vscode.commands.executeCommand('myPaas.openCanvas').then(() => {
          // Small delay to let the webview initialize
          setTimeout(() => {
            CanvasPanel.currentPanel?.addTemplateToCanvas(templateId);
          }, 1000);
        });
      }
    }),

    vscode.commands.registerCommand('myPaas.refreshServices', () => {
      loadServicesFromProject();
    }),

    vscode.commands.registerCommand('myPaas.composeUp', async () => {
      const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
      if (!workspaceFolder) { return; }
      await docker.composeUp(workspaceFolder.uri.fsPath);
      vscode.window.showInformationMessage('Compose Up started.');
    }),

    vscode.commands.registerCommand('myPaas.composeDown', async () => {
      const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
      if (!workspaceFolder) { return; }
      await docker.composeDown(workspaceFolder.uri.fsPath);
      vscode.window.showInformationMessage('Compose Down completed.');
    }),

    vscode.commands.registerCommand('myPaas.generateCompose', async () => {
      const panel = CanvasPanel.currentPanel;
      if (panel) {
        panel.requestCanvasState();
      } else {
        vscode.window.showWarningMessage('Open the canvas first.');
      }
    }),

    vscode.commands.registerCommand('myPaas.detectApp', async () => {
      const folder = await vscode.window.showOpenDialog({
        canSelectFolders: true,
        canSelectFiles: false,
        canSelectMany: false,
        openLabel: 'Select App Source Directory',
      });
      if (!folder || folder.length === 0) { return; }

      const sourcePath = folder[0].fsPath;

      try {
        const result = await detect(context.extensionPath, sourcePath);
        if (!result.success || !result.plan) {
          vscode.window.showWarningMessage(`Detection failed: ${result.error}`);
          return;
        }

        const plan = result.plan;
        const info = `Detected: ${plan.language}${plan.framework ? ` (${plan.framework})` : ''} v${plan.version || 'latest'}`;
        vscode.window.showInformationMessage(info);

        // Send detection result to canvas panel
        const panel = CanvasPanel.currentPanel;
        if (panel) {
          panel.sendDetectResult(plan);
        }
      } catch (err: any) {
        vscode.window.showErrorMessage(`Auto Detect error: ${err.message}`);
      }
    })
  );
}

export function deactivate() {}
