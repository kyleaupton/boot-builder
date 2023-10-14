export type t_flashing_progress = {
  currentActivity: string;
  stdout: string;
  stderr: string;
  error: string;
  copy?: {
    transferred: number;
    speed: number;
    percentage: number;
    eta: number;
    etaHuman: string;
  };
};
