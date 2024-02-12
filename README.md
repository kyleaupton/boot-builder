# Boot Builder

[![CI](https://github.com/kyleaupton/os-install-maker/actions/workflows/ci.yml/badge.svg)](https://github.com/kyleaupton/os-install-maker/actions/workflows/ci.yml)


An app to easily create bootable USB installers for Windows, macOS, Linux, MS-DOS, and more

![](https://raw.githubusercontent.com/kyleaupton/os-install-maker/main/docs/screenshot.png)

## Todo

- [x] Support Windows creation
- [x] Ship with `wimlib` + it's deps
- [x] Support macOS install creation
- [ ] Support Linux/normal ISO flash
- [ ] Ability to download Windows `.iso` files
- [ ] Ability to download macOS installers (maybe not legal, we'll see)
- [ ] Checksum verification option for Win and Linux

## Quick Setup

```sh
# clone the project
git clone https://github.com/kyleaupton/os-install-maker.git

# enter the project directory
cd os-install-maker

# install dependencies
yarn

# develop
yarn dev
```
