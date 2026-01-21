// Re-export Wails bindings types for convenience
export type { Drive } from '@bindings/boot-builder/internal/drives/models'
export type { Job } from '@bindings/boot-builder/internal/jobs/models'
export { Status } from '@bindings/boot-builder/internal/jobs/models'
export type { InstallerMeta, StartJobRequest } from '@bindings/boot-builder/internal/service/models'
export type { Plan, Target, OSFamily, Step } from '@bindings/boot-builder/internal/core/models'

// Import Target for local use in this file
import type { Target } from '@bindings/boot-builder/internal/core/models'

// Frontend-specific types

/** Event payload from backend job:event emissions */
export interface JobEvent {
  JobID: string
  Type: 'state' | 'step-start' | 'step-end' | 'progress' | 'log' | 'error'
  Message: string
  Step: string
  Percent: number
  Error: string
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
