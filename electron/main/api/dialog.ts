import { dialog, OpenDialogOptions } from 'electron';

export const showOpenDialog = (opts: OpenDialogOptions) => {
  return dialog.showOpenDialog(null, opts);
};
