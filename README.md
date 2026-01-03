# Config

My Personal configuration repository for development tools and system settings.

## Contents

- **Cursor IDE** - Settings and keybindings for Cursor editor
- **Zed** - Settings and keymap for Zed editor
- **AutoHotkey Scripts** - Keyboard remapping utilities:
  - `CapsCtrl-TapEscape.ahk` - CapsLock as Ctrl (hold) or Esc (tap)
  - `DoubleShift-CapsToggle.ahk` - Double Shift to toggle CapsLock
- **Fonts** - JetBrains Mono, PP Neue Montreal, SF Windows, Zed Fonts
- **Extensions** - VS Code/Cursor extension list and VSIX files
- **node-modules-cleaner** - Go utility to find and clean `node_modules` directories

## Usage

### AutoHotkey Scripts
Run the `.ahk` files directly or use `ahk.bat` to launch them.

### node-modules-cleaner
```bash
# Build
go build -o node-modules-cleaner.exe node-modules-cleaner.go

# Run
./node-modules-cleaner.exe [options]
```

## License

MIT License - see [LICENSE](LICENSE) for details.

