# Steam Account Switcher

A Go CLI application for managing and switching between Steam accounts on Windows.

## CLI Commands

### List Accounts
List all Steam accounts found in your Steam installation.

```bash
steam-account-switcher.exe list
```

**Example output:**
```
Accounts:
 - account1
 - account2
```

### Show Current Account
Display the currently active Steam account (AutoLoginUser).

```bash
steam-account-switcher.exe current
```

**Example output:**
```
account1
```

### Switch Account
Switch to a specified Steam account. Optionally, start Steam automatically and/or run a custom command after switching.

```bash
steam-account-switcher.exe switch <account_name> [-y] [-c <custom_command>]
```
- `<account_name>`: The account name to switch to (must already exist in Steam).
- `-y`: (Optional) Automatically start Steam after switching.
- `-c <custom_command>`: (Optional) Run a custom command after switching (e.g., to launch a game or script). STEAM_PATH will be replaced with the actual Steam path.

**Examples:**
```
steam-account-switcher.exe switch myaccount
steam-account-switcher.exe switch myaccount -y
steam-account-switcher.exe switch myaccount -c "echo Switched!"
```

### Show Version
Display the current version of the application.

```bash
steam-account-switcher.exe version
```

**Example output:**
```
v1.0.0
```

## Usage Notes

- If no arguments are provided, you will see:
  ```
  No arguments provided. Please use 'list' or 'switch' or 'current'.
  ```
- The application is Windows-only and requires appropriate permissions to access the registry and file system.

## Prerequisites

- Go 1.21 or later
- Windows operating system

## Project Structure

```
steam-account-switcher-win/
├── main.go          # Main application entry point
├── go.mod           # Go module file
├── go.sum           # Go module checksums
├── .gitignore       # Git ignore rules
```

## Scripts

### Build

```bash
./make-build.sh
```

Simple script to build the application from source in MacOS.