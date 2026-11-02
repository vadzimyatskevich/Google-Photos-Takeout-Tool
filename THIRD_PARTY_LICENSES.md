# Third-party software

## ExifTool

TakeoutRestore embeds **ExifTool** by Phil Harvey (https://exiftool.org/).

- **Windows:** the official Windows package, including its bundled Perl runtime.
- **Linux and macOS:** the official Image-ExifTool source package, run with the system Perl.

ExifTool is free software; you can redistribute it and/or modify it under the same terms as Perl itself: either the GNU General Public License (version 1 or later) or the Artistic License. See https://exiftool.org/#license.

ExifTool is not modified. It is unpacked on first run (`%LOCALAPPDATA%\TakeoutRestore\` on Windows, `~/.local/share/TakeoutRestore/` on Linux, `~/Library/Application Support/TakeoutRestore/` on macOS) and called as a separate process. The original packages are available at https://exiftool.org/.
