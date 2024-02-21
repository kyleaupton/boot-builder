import builder, { Platform } from 'electron-builder';
import { notarize as doNotarize } from 'electron-notarize';
import { exec } from 'child_process';
import { promisify } from 'util';
import 'dotenv/config';

const execAsync = promisify(exec);

/**
 * build.mjs
 *
 * This script can be invoked via CLI or imported and used as an API
 *
 * CLI Usage:
 *
 *    build.mjs --targets [mac windows] --notarize=true
 *
 * API Usage:
 *
 *    await pack({
 *      mac: true,
 *      windows: true,
 *      notarize: true,
 *    });
 */

const ID = 'com.kyleupton.boot-builder';

export default async function pack({
  mac = true,
  windows = true,
  notarize = true,
}) {
  //
  // Env Validation
  //
  if (
    !process.env.S3_BUCKET ||
    !process.env.CSC_LINK ||
    !process.env.CSC_KEY_PASSWORD
    // !process.env.WIN_CSC_LINK ||
    // !process.env.WIN_CSC_KEY_PASSWORD
  ) {
    throw Error('Missing code signing env variable(s)');
  }

  if (notarize && (!process.env.APPLEID || !process.env.APPLEIDPASS)) {
    throw Error('Missing notarizing env variable(s)');
  }

  if (!mac && !windows) {
    throw Error('Must specify at least one target');
  }

  //
  // Build the app with Vite
  //
  try {
    const { stdout, stderr } = await execAsync(
      'npx vue-tsc --noEmit && vite build',
    );
    console.log(stdout, stderr);
  } catch (e) {
    console.error(e);
    throw e;
  }

  //
  // Pack the app
  //
  /**
   * @type {import('electron-builder').Configuration}
   * @see https://www.electron.build/configuration/configuration
   */
  const config = {
    productName: 'Boot Builder',
    appId: ID,
    asar: true,
    directories: {
      output: 'release/${version}',
    },
    files: ['dist', 'dist-electron'],

    // Mac Options
    mac: {
      mergeASARs: undefined,
      // singleArchFiles: 'electron/lib/*/arm64/**/*',
      x64ArchFiles: '**/*',
      artifactName: 'Boot_Builder_${version}_macOS.${ext}',
      target: {
        target: 'default',
        arch: ['universal'],
      },
    },

    // Windows Options
    win: {
      target: [
        {
          target: 'nsis',
          arch: ['x64'],
        },
      ],
      artifactName: 'Boot_Builder_${version}_Windows.${ext}',
    },
    nsis: {
      oneClick: false,
      perMachine: false,
      allowToChangeInstallationDirectory: true,
      deleteAppDataOnUninstall: false,
    },

    // Update Options
    // generateUpdatesFilesForAllChannels: true,
    // publish: {
    //   provider: 's3',
    //   bucket: process.env.S3_BUCKET,
    // },

    // Extra files
    extraFiles: ['electron/lib/**/*'],

    // Notarize
    afterSign: async (context) => {
      if (context.electronPlatformName === 'darwin' && notarize) {
        const { appOutDir } = context;
        const appName = context.packager.appInfo.productFilename;

        await doNotarize({
          appBundleId: ID,
          appPath: `${appOutDir}/${appName}.app`,
          appleId: process.env.APPLEID,
          appleIdPassword: process.env.APPLEIDPASS,
        });
      }
    },
  };

  const targets = new Map();
  if (mac) {
    targets.set(Platform.MAC, new Map());
  }
  if (windows) {
    targets.set(Platform.WINDOWS, new Map());
  }

  return builder.build({
    config,
    targets,
  });
}

if (process.argv.includes('--targets')) {
  await pack({
    mac: process.argv.includes('mac'),
    windows: process.argv.includes('windows'),
    notarize: process.argv.includes('--notarize=true'),
  });
}
