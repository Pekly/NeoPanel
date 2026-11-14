# File previews

Click a file or choose **Open / Preview** in its context menu. NeoPanel checks the file's contents before choosing a viewer; renaming a binary file to `.txt` does not bypass the editor guard.

- **Images:** PNG, JPEG, GIF, WebP, BMP and ICO are displayed in the panel.
- **Videos:** MP4, WebM and Theora Ogg video use the browser's player, with controls and seeking. Actual playback depends on browser codec support. Media stops when leaving the preview; ordinary UI refreshes retain playback state.
- **ZIP/JAR:** Browse folders and search the current archive folder without extracting files. Select a text entry to read it in a read-only preview. Download or explicitly extract the archive separately when an entry needs another application.
- **Text:** Valid UTF-8 text files continue to use the editor, draft recovery and save workflow. The existing 5 MiB editor limit remains.
- **Other binary files:** Show file details and a download button, never an editable text buffer. The editor API also refuses binary overwrites. HTML, SVG and XML are not rendered as active documents; valid text versions can be edited as source.

All preview requests require `files.view` on the selected server. They use the same node routing and authentication as the file manager. Media responses support ranges, use an allowlisted content type and are marked `nosniff` and `private, no-store`. Remote media streams do not have the ordinary API's total-body timeout; connection/header setup remains bounded and disconnects cancel the stream.

## Archive limits

To protect the panel from oversized or malicious archives:

- Archive file size: **256 MiB** maximum.
- Central directory: **8 MiB** maximum; **10,000 entries** maximum.
- Entry name: **4,096 bytes** maximum.
- Text entry preview: **1 MiB** maximum expanded size, with CRC verification.
- Store and Deflate compression are supported. ZIP64, encryption, unsupported compression, unsafe paths and symbolic links are rejected with an error.

Previews never extract archive entries or modify the source file. Update both panel and worker nodes to use the new preview endpoints.
