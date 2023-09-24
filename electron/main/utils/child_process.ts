import { exec as oldExec } from 'node:child_process';
import { promisify } from 'node:util';

export const exec = promisify(oldExec);
