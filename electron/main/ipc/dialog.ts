import {
  ipcMain,
  dialog,
  BrowserWindow,
  OpenDialogOptions,
  MessageBoxOptions,
  IpcMainInvokeEvent,
} from 'electron';

export default function start() {
  ipcMain.handle(
    '/dialog/showOpenDialog',
    (event: IpcMainInvokeEvent, options: OpenDialogOptions) => {
      const sender = BrowserWindow.fromWebContents(event.sender);

      if (sender) {
        return dialog.showOpenDialog(sender, options);
      }
    },
  );

  ipcMain.handle(
    '/dialog/showMessageBox',
    (event: IpcMainInvokeEvent, options: MessageBoxOptions) => {
      const sender = BrowserWindow.fromWebContents(event.sender);

      if (sender) {
        return dialog.showMessageBox(sender, options);
      }
    },
  );
}
