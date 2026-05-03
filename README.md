# Typer Text Editor
### A simple and easy to use text editor written in Go

|                          Default Style                           |                          Classic Style                           |
|:----------------------------------------------------------------:|:----------------------------------------------------------------:|
| ![Example of the Typer's default style](media/default-style.png) | ![Example of the Typer's classic style](media/classic-style.png) |

### Installation
#### From a package manager:
|      Distribution      | Package name         |
|:----------------------:|:---------------------|
| Arch Linux/Artix Linux | `typer` from the AUR |
| Tide Linux | `typer` from the main repository |
#### From the releases section:
- Go to the [releases section](https://github.com/EnumeratedDev/typer/releases)
- Choose either the latest stable release (**Recommended**) or nightly pre-release
- Download the archive that corresponds to your operating system and architecture
- Optional: Add the extracted directory to your PATH so that typer can be launched from anywhere
- Optional: On Unix-based systems you can move the `typer` executable into `/usr/local/bin/typer` and `config` directory into `/usr/local/etc/typer` for a system-wide installation
#### From source:
- Download `go` from your package manager or from the go website
- Downlaod `which` from your package manager
- Download `make` from your package manager
- Run the following command to compile Typer
```shell
make
```
- Run the following command **with superuser privileges** to install Typer to your system
```shell
make install
```
