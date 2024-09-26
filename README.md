# Boot Builder

[![CI](https://github.com/kyleaupton/boot-builder/actions/workflows/ci.yml/badge.svg)](https://github.com/kyleaupton/boot-builder/actions/workflows/ci.yml)

> [!NOTE]
> This project is still under active development. Please check back soon for a v1 release!

Boot Builder is a lightweight, easy-to-use desktop app that helps you create bootable USB drives for a various operating systems.

![](https://raw.githubusercontent.com/kyleaupton/boot-builder/main/docs/screenshot.png)

## Support Matrix

> [!NOTE]
> We intend to support creating bootable USB installers on Windows, macOS, Linux. However, we are still in the early stages of development and have not yet implemented all features. The table below shows which features are currently supported.

Boot Builder supports creating bootable USB installers across various host operating systems. The table below shows which host OSes (your computer’s OS) can create installers for specific target OSes.

|                    | Creating Windows USB | Creating macOS USB | Creating Linux USB | Creating MS-DOS USB |
| ------------------ | -------------------- | ------------------ | ------------------ | ------------------- |
| Running on Windows | ❌                    | ❌                  | ❌                  | ❌                   |
| Running on macOS   | ✅                    | ✅                  | ❌                  | ❌                   |
| Running on Linux   | ❌                    | ❌                  | ❌                  | ❌                   |

## Development

Want to contribute? Here's how you can run Boot Builder in development mode:

```sh
# Clone the project
git clone https://github.com/kyleaupton/boot-builder.git

# Enter the project directory
cd boot-builder

# Install dependencies. We use classic yarn (v1) for this project.
yarn

# Develop
yarn dev
```

## Roadmap

- [ ] macOS
  - [x] Support Windows creation
  - [x] Support macOS creation (somewhat supported but needs further iteration)
  - [ ] Support Linux creation
  - [ ] Support MS-DOS creation
- [ ] Windows
  - [ ] Support Windows creation
  - [ ] Support macOS creation (if possible, needs RnD)
  - [ ] Support Linux creation
  - [ ] Support MS-DOS creation
- [ ] Linux
  - [ ] Support Windows creation
  - [ ] Support macOS creation
  - [ ] Support Linux creation
  - [ ] Support MS-DOS creation
- [ ] Ability to download Windows `.iso`s ([see win-iso](https://github.com/kyleaupton/win-iso))
- [ ] Ability to download macOS installers (maybe not legal, we'll see)
- [ ] Ability to download Linux `.iso`s
- [ ] Checksum verification
