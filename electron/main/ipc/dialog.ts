import {
  ipcMain,
  dialog,
  BrowserWindow,
  OpenDialogOptions,
  IpcMainInvokeEvent,
} from 'electron';

export default function start() {
  ipcMain.handle(
    '/dialog/showOpenDialog',
    (event: IpcMainInvokeEvent, options: OpenDialogOptions) => {
      const sender = BrowserWindow.fromWebContents(event.sender);
      return dialog.showOpenDialog(sender, options);
    },
  );
}
