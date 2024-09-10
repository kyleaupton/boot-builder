import path from 'node:path';
import { createIpcHandlers } from 'typed-electron-ipc';

export const pathIpc = () =>
  createIpcHandlers({
    '/path/basename': async (event, p: string, suffix?: string) =>
      path.basename(p, suffix),
  });
