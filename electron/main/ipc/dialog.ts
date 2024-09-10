import {
  dialog,
  BrowserWindow,
  OpenDialogOptions,
  MessageBoxOptions,
} from 'electron';
import { createIpcHandlers } from 'typed-electron-ipc';

export const dialogIpc = () =>
  createIpcHandlers({
    '/dialog/showOpenDialog': async (event, options: OpenDialogOptions) => {
      const sender = BrowserWindow.fromWebContents(event.sender);

      if (!sender) {
        throw new Error('No sender found');
      }

      return dialog.showOpenDialog(sender, options);
    },

    '/dialog/showMessageBox': async (event, options: MessageBoxOptions) => {
      const sender = BrowserWindow.fromWebContents(event.sender);

      if (!sender) {
        throw new Error('No sender found');
      }

      return dialog.showMessageBox(sender, options);
    },
  });
