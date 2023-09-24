import { exec } from '../utils/child_process';

export interface Disks {
  AllDisksAndPartitions: AllDisksAndPartition[];
  VolumesFromDisks: string[];
  AllDisks: string[];
  WholeDisks: string[];
}

export interface AllDisksAndPartition {
  Partitions: Partition[];
  OSInternal: boolean;
  Content: string;
  Size: number;
  DeviceIdentifier: string;
  APFSPhysicalStores?: ApfsphysicalStore[];
  APFSVolumes?: Apfsvolume[];
}

export interface Partition {
  DiskUUID: string;
  Content: string;
  Size: number;
  DeviceIdentifier: string;
  MountPoint?: string;
  VolumeName?: string;
  VolumeUUID?: string;
}

export interface ApfsphysicalStore {
  DeviceIdentifier: string;
}

export interface Apfsvolume {
  DiskUUID: string;
  DeviceIdentifier: string;
  Size: number;
  OSInternal: boolean;
  MountPoint?: string;
  MountedSnapshots?: MountedSnapshot[];
  VolumeName: string;
  CapacityInUse: number;
  VolumeUUID: string;
}

export interface MountedSnapshot {
  SnapshotMountPoint: string;
  SnapshotUUID: string;
  SnapshotName: string;
  SnapshotBSD: string;
  Sealed: string;
}

export const getDisks = async (): Promise<Disks> => {
  return JSON.parse(
    (
      await exec(
        'diskutil list -plist external physical | plutil -convert json -o - -',
      )
    ).stdout,
  );
};
