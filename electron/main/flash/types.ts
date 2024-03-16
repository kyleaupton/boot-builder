export interface Progress {
  id: string;
  activity: string;
  done: boolean;
  canceled?: boolean;
  transferred: number;
  speed: number;
  percentage: number;
  eta: number;
}
