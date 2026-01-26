# FlashIt Frontend

The Vue 3 frontend for FlashIt, a cross-platform desktop app for creating bootable USB OS installers.

## Tech Stack

- **Vue 3**
- **Pinia** for state management
- **Tailwind CSS v4** for styling
- **shadcn-vue** for UI components (built on Reka UI)
- **Wails v3** runtime for Go backend communication

## Key Concepts

### Stores

- **drives** - Polls the Go backend for available USB drives, handles drive selection
- **source** - Manages the selected ISO file and detected installer type
- **job** - Tracks flash job lifecycle (pending → running → complete/error)

### Backend Communication

The frontend communicates with Go services via auto-generated bindings in `bindings/`. These are regenerated when running `task dev` or `wails3 generate bindings`.

Events from the backend (job progress, drive updates) are received via the Wails event system.

### UI Components

Base UI components from shadcn-vue are in `components/ui/`. App-specific components are in the `components/` root.

