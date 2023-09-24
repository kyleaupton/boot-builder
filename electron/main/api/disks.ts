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

export interface SystemProfiler {
  SPUSBDataType: SpusbdataType[];
}

export interface SpusbdataType {
  _items?: Item[];
  _name: string;
  host_controller: string;
  pci_device?: string;
  pci_revision?: string;
  pci_vendor?: string;
}

export interface Item {
  _name: string;
  bcd_device: string;
  bus_power: string;
  bus_power_used: string;
  device_speed: string;
  extra_current_used: string;
  location_id: string;
  manufacturer: string;
  product_id: string;
  serial_num: string;
  sleep_current?: string;
  vendor_id: string;
  _items?: Item2[];
  'Built-in_Device'?: string;
}

export interface Item2 {
  _name: string;
  bcd_device: string;
  bus_power: string;
  bus_power_used: string;
  device_speed: string;
  extra_current_used: string;
  location_id: string;
  manufacturer: string;
  Media?: Medum[];
  product_id: string;
  serial_num: string;
  vendor_id: string;
  _items?: Item3[];
}

export interface Medum {
  _name: string;
  bsd_name: string;
  'Logical Unit': number;
  partition_map_type: string;
  removable_media: string;
  size: string;
  size_in_bytes: number;
  smart_status: string;
  'USB Interface': number;
  volumes: Volume[];
}

export interface Volume {
  _name: string;
  bsd_name: string;
  file_system: string;
  iocontent: string;
  size: string;
  size_in_bytes: number;
  volume_uuid: string;
  free_space?: string;
  free_space_in_bytes?: number;
  mount_point?: string;
  writable?: string;
}

export interface Item3 {
  _items?: Item4[];
  _name: string;
  bcd_device: string;
  bus_power: string;
  bus_power_used?: string;
  device_speed: string;
  extra_current_used: string;
  location_id: string;
  manufacturer: string;
  product_id: string;
  vendor_id: string;
  serial_num?: string;
}

export interface Item4 {
  _name: string;
  bcd_device: string;
  bus_power: string;
  bus_power_used: string;
  device_speed: string;
  extra_current_used: string;
  location_id: string;
  manufacturer: string;
  product_id: string;
  vendor_id: string;
  _items?: Item5[];
}

export interface Item5 {
  _name: string;
  bcd_device: string;
  bus_power: string;
  bus_power_used: string;
  device_speed: string;
  extra_current_used: string;
  location_id: string;
  manufacturer: string;
  product_id: string;
  vendor_id: string;
  serial_num?: string;
}

export const getUsbData = async (): Promise<SystemProfiler> => {
  return JSON.parse((await exec('system_profiler SPUSBDataType -json')).stdout);
};
