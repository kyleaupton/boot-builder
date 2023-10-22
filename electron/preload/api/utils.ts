import ipc from './ipc';

export const getFileIcon = async (path: string): Promise<string> => {
  return ipc.invoke('/utils/app/fileIcon', { path });
};
