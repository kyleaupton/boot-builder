// Re-export Wails bindings types for convenience
export type { Drive } from '@bindings/boot-builder/internal/drives/models'
export type { Job } from '@bindings/boot-builder/internal/jobs/models'
export { Status } from '@bindings/boot-builder/internal/jobs/models'
export type { InstallerMeta, StartJobRequest, StartJobResponse } from '@bindings/boot-builder/internal/service/models'
export type { Plan, Target, OSFamily, StepInfo } from '@bindings/boot-builder/internal/core/models'

// Import Target for local use in this file
import type { Target } from '@bindings/boot-builder/internal/core/models'

// Frontend-specific types

/** Event payload from backend job:event emissions */
export interface JobEvent {
  jobId: string
  type: 'state' | 'step-start' | 'step-end' | 'progress' | 'log' | 'error'
  message: string
  step: string
  percent: number
  error: string
}

/** Step status for UI display */
export type StepStatus = 'pending' | 'running' | 'completed' | 'failed'

/** Step state for tracking in the frontend */
export interface StepState {
  key: string
  name: string
  hasProgress: boolean
  status: StepStatus
  progress: number
  message: string | null
}

/** Application UI states */
export type AppState =
  | 'empty'       // Initial: no source selected
  | 'source-only' // ISO selected, no drive selected
  | 'ready'       // Both selected, can flash
  | 'in-progress' // Flashing in progress
  | 'complete'    // Successfully finished
  | 'error'       // Error occurred

/** Source file info after selection/detection */
export interface SourceInfo {
  path: string
  filename: string
  sizeBytes: number
  installerID: string | null
  installerName: string | null
  detectedTargets: Target[]
}
