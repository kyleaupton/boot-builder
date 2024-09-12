export interface Progress {
  activity: string;
  transferred: number;
  speed: number;
  percentage: number;
  eta: number;
}

export interface SerializedFlash {
  id: string;
  status: string;
  done: boolean;
  canceled: boolean;
  progress: Progress;
}

export interface Flash extends SerializedFlash {
  flash: () => Promise<void>;
}
