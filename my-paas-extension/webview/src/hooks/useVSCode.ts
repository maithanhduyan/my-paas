import { useCallback, useEffect, useRef } from 'react';
import { useCanvasStore } from '../store/canvasStore';

interface VSCodeAPI {
  postMessage(message: any): void;
  getState(): any;
  setState(state: any): void;
}

declare function acquireVsCodeApi(): VSCodeAPI;

let vscodeApi: VSCodeAPI | null = null;

function getVSCodeAPI(): VSCodeAPI {
  if (!vscodeApi) {
    vscodeApi = acquireVsCodeApi();
  }
  return vscodeApi;
}

export function useVSCode() {
  const {
    setTemplates, setStatuses, addLogLine, addToast,
    loadCanvasState, templates, addNode, getCanvasState,
  } = useCanvasStore();

  const handlersRef = useRef({ setTemplates, setStatuses, addLogLine, addToast, loadCanvasState, templates, addNode, getCanvasState });
  handlersRef.current = { setTemplates, setStatuses, addLogLine, addToast, loadCanvasState, templates, addNode, getCanvasState };

  useEffect(() => {
    const handler = (event: MessageEvent) => {
      const message = event.data;
      const h = handlersRef.current;

      switch (message.type) {
        case 'templates:data':
          h.setTemplates(message.templates);
          break;
        case 'canvas:loaded':
          if (message.state) {
            h.loadCanvasState(message.state);
          }
          break;
        case 'docker:statuses':
          h.setStatuses(message.data);
          break;
        case 'docker:log':
          h.addLogLine(message.serviceId, message.line);
          break;
        case 'compose:generated':
          h.addToast('info', `Compose file generated: ${message.path}`);
          break;
        case 'info':
          h.addToast('info', message.message);
          break;
        case 'error':
          h.addToast('error', message.message);
          break;
        case 'canvas:requestState': {
          const state = h.getCanvasState();
          getVSCodeAPI().postMessage({ type: 'compose:generate', state });
          break;
        }
        case 'template:add': {
          const tpl = h.templates.find((t: any) => t.id === message.templateId);
          if (tpl) {
            const id = `${tpl.id}-${Date.now()}`;
            h.addNode({
              id,
              type: 'serviceNode',
              position: { x: 200 + Math.random() * 200, y: 100 + Math.random() * 200 },
              data: {
                name: tpl.defaults.name,
                image: tpl.defaults.image,
                build: tpl.defaults.build,
                ports: [...tpl.defaults.ports],
                environment: { ...tpl.defaults.environment },
                volumes: [...tpl.defaults.volumes],
                healthcheck: tpl.defaults.healthcheck ? { ...tpl.defaults.healthcheck } : undefined,
                command: tpl.defaults.command,
                restart: tpl.defaults.restart,
              },
            } as any);
            h.addToast('info', `Added ${tpl.name} to canvas`);
          }
          break;
        }
        case 'core:detected': {
          const plan = message.plan;
          const id = `auto-${Date.now()}`;
          const ports = (plan.ports || []).map((p: string) => `${p}:${p}`);
          h.addNode({
            id,
            type: 'serviceNode',
            position: { x: 300 + Math.random() * 150, y: 150 + Math.random() * 150 },
            data: {
              name: plan.framework || plan.provider || 'app',
              build: { context: './', dockerfile: 'Dockerfile' },
              ports,
              environment: plan.envVars || {},
              volumes: [],
              restart: 'unless-stopped',
            },
          } as any);
          h.addToast('info', `Auto-detected: ${plan.language}${plan.framework ? ` (${plan.framework})` : ''}`);
          break;
        }
      }
    };

    window.addEventListener('message', handler);

    // Signal ready
    getVSCodeAPI().postMessage({ type: 'ready' });

    return () => window.removeEventListener('message', handler);
  }, []);

  const postMessage = useCallback((message: any) => {
    getVSCodeAPI().postMessage(message);
  }, []);

  return { postMessage };
}
