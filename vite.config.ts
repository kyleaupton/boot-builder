import { rmSync } from 'node:fs';
import url from 'node:url';
import path from 'node:path';
import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import electron from 'vite-plugin-electron/simple';
import alias from '@rollup/plugin-alias';
import { globSync } from 'glob';

import pkg from './package.json';

const __filename = url.fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const globImport = (g: string) => {
  return globSync(g).map((item) =>
    url.fileURLToPath(new URL(item, import.meta.url)),
  );
};

const _alias = alias({
  entries: [
    {
      find: '@main',
      replacement: path.resolve(__dirname, './electron/main'),
    },
    {
      find: '@preload',
      replacement: path.resolve(__dirname, './electron/preload'),
    },
    {
      find: '@shared',
      replacement: path.resolve(__dirname, './electron/shared'),
    },
  ],
});

// https://vitejs.dev/config/
export default defineConfig(({ command }) => {
  rmSync('dist-electron', { recursive: true, force: true });

  const isServe = command === 'serve';
  const isBuild = command === 'build';
  const sourcemap = isServe || !!process.env.VSCODE_DEBUG;

  return {
    resolve: {
      // Aliases for the renderer process are defined here
      alias: {
        '@': path.resolve(__dirname, './src'),
        '@main': path.resolve(__dirname, './electron/main'),
        '@preload': path.resolve(__dirname, './electron/preload'),
        '@shared': path.resolve(__dirname, './electron/shared'),
      },
    },
    plugins: [
      vue(),
      electron({
        main: {
          entry: 'electron/main/index.ts',
          onstart({ startup }) {
            if (process.env.VSCODE_DEBUG) {
              console.log(
                /* For `.vscode/.debug.script.mjs` */ '[startup] Electron App',
              );
            } else {
              // For debugging
              // startup([
              //   '.',
              //   '--no-sandbox',
              //   '--trace-warnings',
              //   '--enable-logging',
              // ]);

              // Normal
              startup();
            }
          },
          vite: {
            build: {
              sourcemap,
              minify: isBuild,
              outDir: 'dist-electron/main',
              rollupOptions: {
                plugins: [
                  // Aliases for the main process are defined here
                  _alias,
                ],
                // Some third-party Node.js libraries may not be built correctly by Vite, especially `C/C++` addons,
                // we can use `external` to exclude them to ensure they work correctly.
                // Others need to put them in `dependencies` to ensure they are collected into `app.asar` after the app is built.
                // Of course, this is not absolute, just this way is relatively simple. :)
                external: Object.keys(
                  'dependencies' in pkg ? pkg.dependencies : {},
                ),
                input: [
                  'electron/main/index.ts',
                  ...globImport('electron/main/flash/workers/*.ts'),
                ],
              },
            },
          },
        },
        preload: {
          // Shortcut of `build.rollupOptions.input`.
          // Preload scripts may contain Web assets, so use the `build.rollupOptions.input` instead `build.lib.entry`.
          input: 'electron/preload/index.ts',
          vite: {
            build: {
              sourcemap: sourcemap ? 'inline' : undefined, // #332
              minify: isBuild,
              outDir: 'dist-electron/preload',
              rollupOptions: {
                external: Object.keys(
                  'dependencies' in pkg ? pkg.dependencies : {},
                ),
              },
            },
          },
        },
        renderer: {},
      }),
    ],
    server:
      process.env.VSCODE_DEBUG &&
      (() => {
        const url = new URL(pkg.debug.env.VITE_DEV_SERVER_URL);
        return {
          host: url.hostname,
          port: +url.port,
        };
      })(),
    clearScreen: false,
  };
});
