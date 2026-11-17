import { access, cp, rm } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';

// Vditor loads its i18n, icon and language assets at runtime from a configurable
// CDN. Serving them from our own origin removes a third-party script path into
// the admin session and lets the strict CSP in deploy/nginx.conf stay in force.
const source = new URL('../node_modules/vditor/dist', import.meta.url);
const target = new URL('../public/vditor/dist', import.meta.url);

try {
  await access(source);
} catch {
  console.warn('[vditor] node_modules/vditor is missing; run `npm ci` before building.');
  process.exit(0);
}

await rm(target, { recursive: true, force: true });
await cp(source, target, { recursive: true });
console.log(`[vditor] copied runtime assets to ${fileURLToPath(target)}`);
