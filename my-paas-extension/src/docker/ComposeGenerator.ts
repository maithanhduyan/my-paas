import { stringify } from 'yaml';
import type { CanvasState, ServiceNode, ServiceEdge } from '../shared/types';

export class ComposeGenerator {

  generate(state: CanvasState, servicesBasePath?: string): string {
    const services: Record<string, any> = {};
    const volumes: Record<string, any> = {};
    const hasMultipleServices = state.nodes.length > 1;

    // Build dependency map from edges (only depends_on/network, not volume)
    const dependsOnMap = new Map<string, string[]>();
    for (const edge of state.edges) {
      if (edge.type === 'volume') { continue; }
      const sourceNode = state.nodes.find(n => n.id === edge.source);
      const targetNode = state.nodes.find(n => n.id === edge.target);
      if (!sourceNode || !targetNode) { continue; }

      // edge: source → target means target depends on source
      const deps = dependsOnMap.get(targetNode.data.name) || [];
      deps.push(sourceNode.data.name);
      dependsOnMap.set(targetNode.data.name, deps);
    }

    // Build volume-to-service map from volumeNodes + volume edges
    const serviceVolumeBinds = new Map<string, string[]>(); // serviceName -> bind mount strings
    if (state.volumeNodes) {
      for (const vol of state.volumeNodes) {
        // Find service connection via edges
        let connectedServiceName: string | undefined;
        for (const edge of state.edges) {
          if (edge.type !== 'volume') { continue; }
          // volume -> service or service -> volume
          const svcNode = state.nodes.find(n =>
            (n.id === edge.source && vol.id === edge.target) ||
            (n.id === edge.target && vol.id === edge.source)
          );
          if (svcNode) {
            connectedServiceName = svcNode.data.name;
            break;
          }
        }
        if (connectedServiceName && servicesBasePath) {
          // Use bind mount from .mypaas/services/<name>/volumes/<vol-name> -> mountPath
          const hostPath = `.mypaas/services/${connectedServiceName}/volumes/${vol.name}`;
          const bindMount = `${hostPath}:${vol.mountPath}`;
          const existing = serviceVolumeBinds.get(connectedServiceName) || [];
          existing.push(bindMount);
          serviceVolumeBinds.set(connectedServiceName, existing);
        }
      }
    }

    // Build healthcheck lookup for depends_on conditions
    const hasHealthcheck = new Set<string>();

    for (const node of state.nodes) {
      const svc: any = {};
      const d = node.data;

      if (d.image) {
        svc.image = d.image;
      }

      if (d.build) {
        svc.build = {
          context: d.build.context,
          dockerfile: d.build.dockerfile,
        };
      }

      if (d.ports && d.ports.length > 0) {
        svc.ports = d.ports;
      }

      if (d.environment && Object.keys(d.environment).length > 0) {
        svc.environment = { ...d.environment };
      }

      if (d.volumes && d.volumes.length > 0) {
        svc.volumes = [...d.volumes];
        // Extract named volumes
        for (const v of d.volumes) {
          const parts = v.split(':');
          if (parts.length >= 2 && !parts[0].startsWith('.') && !parts[0].startsWith('/')) {
            volumes[parts[0]] = null;
          }
        }
      }

      // Add bind-mount volumes from VolumeNodes
      const extraBinds = serviceVolumeBinds.get(d.name);
      if (extraBinds && extraBinds.length > 0) {
        if (!svc.volumes) { svc.volumes = []; }
        svc.volumes.push(...extraBinds);
      }

      // Add log bind-mount: .mypaas/services/<name>/logs -> /var/log/app
      if (servicesBasePath) {
        if (!svc.volumes) { svc.volumes = []; }
        svc.volumes.push(`.mypaas/services/${d.name}/logs:/var/log/app`);
      }

      if (d.command) {
        svc.command = d.command;
      }

      if (d.restart) {
        svc.restart = d.restart;
      }

      if (d.healthcheck) {
        svc.healthcheck = { ...d.healthcheck };
        hasHealthcheck.add(d.name);
      }

      if (d.labels && Object.keys(d.labels).length > 0) {
        svc.labels = d.labels;
      }

      // Add depends_on
      const deps = dependsOnMap.get(d.name);
      if (deps && deps.length > 0) {
        const dependsOn: Record<string, any> = {};
        for (const dep of deps) {
          if (hasHealthcheck.has(dep)) {
            dependsOn[dep] = { condition: 'service_healthy' };
          } else {
            dependsOn[dep] = { condition: 'service_started' };
          }
        }
        svc.depends_on = dependsOn;
      }

      // Add network if multiple services
      if (hasMultipleServices) {
        svc.networks = ['app-network'];
      }

      services[d.name] = svc;
    }

    // Second pass: fix depends_on conditions now that we know all healthchecks
    for (const node of state.nodes) {
      const svc = services[node.data.name];
      if (svc?.depends_on) {
        for (const dep of Object.keys(svc.depends_on)) {
          if (hasHealthcheck.has(dep)) {
            svc.depends_on[dep] = { condition: 'service_healthy' };
          }
        }
      }
    }

    const composeObj: any = { services };

    if (Object.keys(volumes).length > 0) {
      composeObj.volumes = volumes;
    }

    if (hasMultipleServices) {
      composeObj.networks = {
        'app-network': { driver: 'bridge' },
      };
    }

    const header = `# Auto-generated by My PaaS Extension\n# Project: ${state.name}\n\n`;
    return header + stringify(composeObj, { lineWidth: 120 });
  }
}
