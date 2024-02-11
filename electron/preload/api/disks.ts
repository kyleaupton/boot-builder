import drives from 'drivelist';

export const getDisks = async (): Promise<drives.Drive[]> => {
  return drives.list();
};
