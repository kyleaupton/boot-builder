import { exec } from '../utils/child_process';

export const create = async ({
  isoFile,
  volume,
}: {
  isoFile: string;
  volume: string;
}) => {
  const diskName = 'WIN10';

  // Erase disk - GPT
  await exec(`diskutil eraseDisk MS-DOS "${diskName}" GPT ${volume}`);

  // Erase disk - MBR
  // await exec(`diskutil eraseDisk MS-DOS "${diskName}" MBR ${volume}`);

  // Mount ISO volume
  await exec(`hdiutil mount ${isoFile}`);

  // TODO: figure out what path the ISO volume is mounted to
  const isoVolume = '';

  // Copy over everything minus the big file
  await exec(
    `rsync -vha --exclude=sources/install.wim ${isoVolume}/* /Volumes/${diskName}`,
  );

  // Make directory for big file
  // But is this really needed...? Should already be there
  // await exec(`mkdir /Volumes/${diskName}/sources`);

  // Next copy the big file
  await exec(
    `wimlib-imagex split ${isoVolume}/sources/install.wim /Volumes/${diskName}/sources/install.swm 3800`,
  );
};
