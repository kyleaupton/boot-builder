import fs from 'node:fs/promises';
import child_process from 'node:child_process';
import path from 'node:path';
import url from 'node:url';
import util from 'node:util';
import inquirer from 'inquirer';
import ora from 'ora';
import { table } from 'table';

const TEST_DRIVE_DIR = 'test-drives';
const exec = util.promisify(child_process.exec);

const __dirname = path.dirname(url.fileURLToPath(import.meta.url));
const __project = path.join(__dirname, '..');
const __testDrives = path.join(__project, TEST_DRIVE_DIR);

/**
 * Represents a drive
 * @typedef {{
 *  name: string;
 *  mounted: boolean;
 *  size: string;
 * }} Drive
 *
 * Represents mount info
 * @typedef {{
 *  name: string;
 *  imagePath: string;
 *  mountPath: string;
 * }} MountInfo
 */

/**
 * getMountInfo
 * @returns {Promise<MountInfo[]>}
 */
const getMountInfo = async () => {
  /** @type {MountInfo[]} */
  const payload = [];
  const imagePathRegex = /image-path\s+:\s+(.+)/;
  const mountedPathRegex = /\/Volumes\/[^\s]+/;

  const { stdout: mountStdout } = await exec('hdiutil info');

  // Each section is separated by a line of equal signs
  // As to not hard-code the number of equal signs, we use a regex
  // I assume there will be at least 5 equal signs
  const sections = mountStdout.split(/={5,}/);

  for (const section of sections) {
    const imagePathMatch = section.match(imagePathRegex);
    const mountedPathMatch = section.match(mountedPathRegex);

    if (imagePathMatch) {
      const imagePath = imagePathMatch[1];

      payload.push({
        name: path.basename(imagePath),
        imagePath,
        mountedPath: mountedPathMatch ? mountedPathMatch[0] : null,
      });
    }
  }

  return payload;
};

/**
 * @returns {Promise<Drive[]>}
 */
const getDrives = async () => {
  /** @type {Drive[]} */
  const payload = [];

  const mountInfo = await getMountInfo();

  /** @type {string[]} */
  let existingDrives = [];
  try {
    existingDrives = await fs.readdir(__testDrives);
  } catch (e) {
    // Noop
  }

  for (const existingDrive of existingDrives) {
    // Get size of drive
    const { stdout: sizeStdout } = await exec(
      `du -sh ${path.join(__testDrives, existingDrive)}`,
    );

    payload.push({
      name: existingDrive,
      mounted: !!mountInfo.find((x) => x.name === existingDrive),
      size: sizeStdout.split('\t')[0],
    });
  }

  return payload;
};

/**
 * mountDrive
 * @param {string} driveName
 */
const mountDrive = async (driveName) => {
  return exec(`hdiutil attach ${path.join(__testDrives, driveName)}`);
};

/***
 * unmountDrive
 * @param {string} driveName
 */
const unmountDrive = async (driveName) => {
  const mountInfo = await getMountInfo();
  const mountedDrive = mountInfo.find((x) => x.name === driveName);

  if (!mountedDrive) {
    throw new Error(`Drive ${driveName} is not mounted`);
  }

  return exec(`hdiutil detach ${mountedDrive.mountPath}`);
};

//
// Main
//
const drives = await getDrives();

console.log(
  table([['Name', 'Mounted?', 'Size'], ...drives.map((x) => Object.values(x))]),
);

const { action } = await inquirer.prompt([
  {
    type: 'list',
    message: 'What would you like to do?',
    name: 'action',
    choices: [
      { name: 'Create drive', value: 'create' },
      { name: 'Delete drive', value: 'delete' },
      { name: 'Mount drive', value: 'mount' },
    ],
  },
]);

if (action === 'create') {
  //
  // Create drive
  //

  /**
   * @param {number} n
   * @returns {string}
   */
  const getDefault = (n = 1) => {
    const baseName = 'test-drive';
    const driveName = `${baseName}-${n}`;
    if (drives.find((x) => x.name.startsWith(driveName))) {
      return getDefault(n + 1);
    }
    return driveName;
  };

  const defaultName = getDefault();
  const { driveName, driveSize } = await inquirer.prompt([
    {
      type: 'input',
      message: 'What would you like to name the drive?',
      name: 'driveName',
      default: defaultName,
    },
    {
      type: 'input',
      message: 'How large would you like the drive to be?',
      name: 'driveSize',
      default: defaultName,
    },
  ]);

  const _name = `${driveName}.dmg`;

  // Create drive
  const createSpinner = ora(`Creating drive ${_name}`).start();
  await exec(
    `hdiutil create -size 1g -fs 'APFS' -volname ${driveName} ${path.join(__testDrives, _name)}`,
  );
  createSpinner.succeed('Drive created');

  // Mount drive
  const mountSpinner = ora('Mounting drive').start();
  // await mountDrive();
  await new Promise((resolve) => setTimeout(resolve, 2000));
  mountSpinner.succeed('Drive mounted');
} else if (action === 'delete') {
  //
  // Delete drive
  //
  const { driveName } = await inquirer.prompt([
    {
      type: 'list',
      message: 'What drive would you like to delete?',
      name: 'driveName',
      choices: drives.map((x) => x.name),
    },
  ]);

  // unount
} else if (action === 'mount') {
  //
  // Mount drive
  //
}
