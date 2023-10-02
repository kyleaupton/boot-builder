import { stat } from 'fs/promises';
import { basename } from 'path';

export const getFileFromPath = async (
  path: string,
): Promise<{ name: string; path: string; size: number }> => {
  const stats = await stat(path);

  return {
    name: basename(path),
    path,
    size: stats.size,
  };
};
