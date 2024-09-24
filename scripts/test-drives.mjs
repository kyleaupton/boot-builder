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
 *  mountPath: string | null;
 *  size: string;
 * }} Drive
 *
 * Represents mount info
 * @typedef {{
 *  name: string;
 *  imagePath: string;
 *  mountPath: string | null;
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
  const mountPathRegex = /\/Volumes\/.+/;

  const { stdout: mountStdout } = await exec('hdiutil info');

  // Each section is separated by a line of equal signs
  // As to not hard-code the number of equal signs, we use a regex
  // I assume there will be at least 5 equal signs
  const sections = mountStdout.split(/={5,}/);

  for (const section of sections) {
    const imagePathMatch = section.match(imagePathRegex);

    if (imagePathMatch && imagePathMatch[1].startsWith(__testDrives)) {
      const imagePath = imagePathMatch[1];

      // Find last line of section, split on space, get last element, and trim
      const sectionLines = section.trim().split('\n');
      const mountPathMatch = sectionLines.pop()?.match(mountPathRegex);

      if (mountPathMatch && mountPathMatch[0]) {
        payload.push({
          name: path.basename(imagePath),
          imagePath,
          mountPath: mountPathMatch[0],
        });
      }
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

    const mountedDrive = mountInfo.find((x) => x.name === existingDrive);

    payload.push({
      name: existingDrive,
      mountPath: mountedDrive ? mountedDrive.mountPath : null,
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
 * @param {Drive} drive
 */
const unmountDrive = async (drive) => {
  if (drive.mountPath) {
    return exec(`hdiutil detach "${drive.mountPath}"`);
  }
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
      type: 'list',
      message: 'How large would you like the drive to be?',
      name: 'driveSize',
      default: '8g',
      choices: ['1g', '2g', '4g', '8g', '16g', '32g'],
    },
  ]);

  const _name = `${driveName}.dmg`;

  // Create drive
  const createSpinner = ora(`Creating drive ${_name}`).start();
  await exec(
    `hdiutil create -size ${driveSize} -fs MS-DOS -volname "${driveName}" ${path.join(
      __testDrives,
      _name,
    )}`,
  );
  createSpinner.succeed('Drive created');

  // Mount drive
  const mountSpinner = ora('Mounting drive').start();
  await mountDrive(_name);
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

  const unmountSpinner = ora('Unmounting drive').start();
  const drive = drives.find((x) => x.name === driveName);
  if (drive) await unmountDrive(drive);
  unmountSpinner.succeed('Drive unmounted');

  const deleteSpinner = ora('Deleting drive').start();
  await fs.rm(path.join(__testDrives, driveName));
  deleteSpinner.succeed('Drive deleted');
} else if (action === 'mount') {
  //
  // Mount drive
  //
  const unMountedDrives = drives.filter((x) => !x.mountPath);

  if (unMountedDrives.length === 0) {
    console.log('\nNo drives to mount');
    process.exit(0);
  }

  const { driveName } = await inquirer.prompt([
    {
      type: 'list',
      message: 'What drive would you like to mount?',
      name: 'driveName',
      choices: unMountedDrives.map((x) => x.name),
    },
  ]);

  const mountSpinner = ora('Mounting drive').start();
  await mountDrive(driveName);
  mountSpinner.succeed('Drive mounted');
}
