# Burro

Burro is a modular security and traffic inspection tool that allows extending its behavior through a plugin-based architecture.

## License

Burro core is licensed under the GNU General Public License v3.0 (GPLv3). This applies to all source code included in the main repository, including bundled plugins located in the `plugins/` directory, which are considered part of the Burro codebase and distributed under GPLv3.

External plugins that are developed and distributed independently of this repository are not considered part of Burro itself if they interact only through the documented Plugin API (including hook events and IPC-based communication such as stdin/stdout, HTTP, or similar mechanisms). Such external plugins are independent works and may be licensed under any terms chosen by their authors, including permissive licenses (e.g. MIT, Apache-2.0) or proprietary licenses.

More details here:

- [PLUGIN_EXCEPTION.md](./PLUGIN_EXCEPTION.md)
- [PLUGINS](./PLUGINS.md)
