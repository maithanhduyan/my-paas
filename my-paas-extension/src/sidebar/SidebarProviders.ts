import * as vscode from 'vscode';
import type { ServiceTemplate } from '../shared/types';
import { TemplateRegistry } from '../templates/TemplateRegistry';

// ─── Services Tree ───────────────────────────────────────────

interface ServiceItem {
  name: string;
  image?: string;
  ports: string[];
  status?: string;
}

export class ServicesTreeProvider implements vscode.TreeDataProvider<ServiceItem> {
  private _onDidChange = new vscode.EventEmitter<ServiceItem | undefined>();
  readonly onDidChangeTreeData = this._onDidChange.event;

  private services: ServiceItem[] = [];

  refresh(services: ServiceItem[]) {
    this.services = services;
    this._onDidChange.fire(undefined);
  }

  getTreeItem(element: ServiceItem): vscode.TreeItem {
    const item = new vscode.TreeItem(element.name, vscode.TreeItemCollapsibleState.None);
    const parts: string[] = [];
    if (element.image) { parts.push(element.image); }
    if (element.ports.length) { parts.push(`Ports: ${element.ports.join(', ')}`); }
    item.description = parts.join(' | ');
    item.tooltip = `${element.name}\n${parts.join('\n')}`;
    item.contextValue = 'service';

    // Icon based on status
    if (element.status === 'running') {
      item.iconPath = new vscode.ThemeIcon('circle-filled', new vscode.ThemeColor('testing.iconPassed'));
    } else if (element.status === 'error') {
      item.iconPath = new vscode.ThemeIcon('circle-filled', new vscode.ThemeColor('testing.iconFailed'));
    } else {
      item.iconPath = new vscode.ThemeIcon('circle-outline');
    }

    return item;
  }

  getChildren(): ServiceItem[] {
    return this.services;
  }
}

// ─── Templates Tree ──────────────────────────────────────────

type TemplateTreeItem = CategoryNode | TemplateNode;

interface CategoryNode {
  kind: 'category';
  label: string;
  category: string;
}

interface TemplateNode {
  kind: 'template';
  template: ServiceTemplate;
}

const CATEGORY_ORDER: Record<string, number> = {
  database: 0,
  app: 1,
  infrastructure: 2,
  tools: 3,
};

const CATEGORY_LABELS: Record<string, string> = {
  database: 'Databases',
  app: 'Applications',
  infrastructure: 'Infrastructure',
  tools: 'Tools',
};

const CATEGORY_ICONS: Record<string, string> = {
  database: 'database',
  app: 'code',
  infrastructure: 'globe',
  tools: 'tools',
};

export class TemplatesTreeProvider implements vscode.TreeDataProvider<TemplateTreeItem> {
  private _onDidChange = new vscode.EventEmitter<TemplateTreeItem | undefined>();
  readonly onDidChangeTreeData = this._onDidChange.event;

  constructor(private registry: TemplateRegistry) {}

  getTreeItem(element: TemplateTreeItem): vscode.TreeItem {
    if (element.kind === 'category') {
      const item = new vscode.TreeItem(element.label, vscode.TreeItemCollapsibleState.Expanded);
      item.iconPath = new vscode.ThemeIcon(CATEGORY_ICONS[element.category] || 'folder');
      item.contextValue = 'category';
      return item;
    }

    const tpl = element.template;
    const item = new vscode.TreeItem(tpl.name, vscode.TreeItemCollapsibleState.None);
    item.description = tpl.defaults.image || 'build';
    item.tooltip = `${tpl.name}\n${tpl.defaults.image || 'Custom build'}\nPorts: ${tpl.defaults.ports.join(', ') || 'none'}`;
    item.iconPath = new vscode.ThemeIcon(tpl.icon);
    item.contextValue = 'template';
    item.command = {
      command: 'myPaas.addTemplate',
      title: 'Add to Canvas',
      arguments: [tpl.id],
    };
    return item;
  }

  getChildren(element?: TemplateTreeItem): TemplateTreeItem[] {
    if (!element) {
      // Root: return categories
      const categories = new Set(this.registry.getAll().map(t => t.category));
      return [...categories]
        .sort((a, b) => (CATEGORY_ORDER[a] ?? 99) - (CATEGORY_ORDER[b] ?? 99))
        .map(cat => ({ kind: 'category' as const, label: CATEGORY_LABELS[cat] || cat, category: cat }));
    }

    if (element.kind === 'category') {
      return this.registry
        .getAll()
        .filter(t => t.category === element.category)
        .map(t => ({ kind: 'template' as const, template: t }));
    }

    return [];
  }
}
