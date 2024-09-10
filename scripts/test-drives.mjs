import fs from 'node:fs/promises';
import child_process from 'node:child_process';
import path from 'node:path';
import url from 'node:url';
import util from 'node:util';
import { program } from 'commander';
import inquirer from 'inquirer';
import ora from 'ora';

const TEST_DRIVE_DIR = 'test-drives';
const exec = util.promisify(child_process.exec);

const __dirname = path.dirname(url.fileURLToPath(import.meta.url));
const __project = path.join(__dirname, '..');
const __testDrives = path.join(__project, TEST_DRIVE_DIR);

const getDrives = async () => {
  const drives = await fs.readdir(__testDrives);
  return drives;
};

program
  .command('list')
  .description('List all test drives')
  .action(async () => {
    const drives = await getDrives();

    if (drives.length === 0) {
      console.log('No test drives found');
      return;
    }

    console.log('Test Drives:');
    drives.forEach((drive) => {
      console.log(`- ${drive}`);
    });
  });

program
  .command('create')
  .option('-s, --size <size>', 'size of the test drive')
  .option('-n, --name <name>', 'name of the test drive')
  .description('Create a test drive')
  .action(async ({ size = '8', name } = {}) => {
    if (!name) {
      const drives = await getDrives();
      const baseName = 'test-drive';

      const getName = (n = 1) => {
        const driveName = `${baseName}-${n}`;
        if (drives.find((x) => x.startsWith(driveName))) {
          return getName(n + 1);
        }
        return driveName;
      };

      name = getName();
    }

    const _name = `${name}.dmg`;

    const size_GB = Math.min(Math.max(parseInt(size), 1), 12);
    const size_MB = size_GB * 1024;

    const spinner = ora(
      `Creating a test drive called ${_name} with ${size_GB}G of storage`,
    ).start();

    await exec(
      `hdiutil create -size ${size_MB}m -fs MS-DOS -volname "name" ${path.join(__testDrives, _name)}`,
    );

    spinner.succeed('Test drive created');
  });

program
  .command('mount')
  .description('Mount a test drive')
  .action(async () => {
    const drives = await getDrives();

    const { drive } = await inquirer.prompt([
      {
        name: 'drive',
        type: 'list',
        message: 'Select a drive to mount',
        choices: drives,
      },
    ]);

    const spinner = ora(`Mounting ${drive}`).start();

    await exec(`hdiutil attach ${path.join(__testDrives, drive)}`);

    spinner.succeed('Drive mounted');
  });

program.parse();
