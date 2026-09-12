import { readFile, writeFile } from 'node:fs/promises';
import { PurgeCSS } from 'purgecss';

// Keep the existing Tailwind version and remove only unused selectors.
// Include every conditional template branch and client-side class name.
const [result] = await new PurgeCSS().purge({
  content: ['templates/*.templ', 'static/*.js'],
  css: ['assets/tailwind-2.2.19.min.css'],
  defaultExtractor: content => content.match(/[A-Za-z0-9_:/.-]+/g) || [],
});
const output = result.css + '\n';
const path = 'static/app.min.css';
if (process.argv.includes('--check')) {
  if (await readFile(path, 'utf8') !== output) {
    throw new Error('CSS is stale. Run npm run build:css and commit static/app.min.css.');
  }
} else {
  await writeFile(path, output);
}
console.log(`Application CSS: ${Buffer.byteLength(output).toLocaleString('en-US')} bytes`);
