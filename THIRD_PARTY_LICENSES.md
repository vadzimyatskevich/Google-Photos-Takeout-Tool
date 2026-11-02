# Third-party software

## ExifTool

The Windows release of TakeoutRestore embeds **ExifTool** by Phil Harvey (https://exiftool.org/), including the bundled Perl runtime from the official Windows package.

ExifTool is free software; you can redistribute it and/or modify it under the same terms as Perl itself: either the GNU General Public License (version 1 or later) or the Artistic License. See https://exiftool.org/#license.

ExifTool is not modified. It is unpacked to `%LOCALAPPDATA%\TakeoutRestore\` on first run and called as a separate process. The original package is available at https://exiftool.org/.
