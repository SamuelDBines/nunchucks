# Nunchucks VS Code Extension

Syntax highlighting extension for Nunchucks templates (`.njk`, `.nunchucks`) with custom tag/filter-friendly tokenization.

## Links

- Source repository: <https://github.com/SamuelDBines/nunchucks>
- Extension folder: <https://github.com/SamuelDBines/nunchucks/tree/main/vscode-nunchucks>
- Issues: <https://github.com/SamuelDBines/nunchucks/issues>

## What This Extension Does

- Registers a dedicated language id: `nunchucks-template`
- Associates file extensions:
  - `.njk`
  - `.nunchucks`
- Applies a TextMate grammar for Nunchucks syntax:
  - tag blocks: `{% ... %}`
  - output blocks: `{{ ... }}`
  - comments: `{# ... #}`
  - filters, operators, and common expression forms
- Uses `language-configuration.json` for editor behaviors:
  - bracket/quote pairs
  - auto-closing
  - surrounding pairs
  - comments configuration

## How It Works (Internals)

- `package.json`
  - declares the language contribution and grammar contribution
- `syntaxes/nunchucks.tmLanguage.json`
  - TextMate grammar rules that drive syntax scopes and colors
- `language-configuration.json`
  - editing experience rules (pairs, comments, etc.)

Coloring itself is controlled by the active VS Code theme. This extension provides scopes; themes choose the final colors.

## Install

### Option 1: Install from local folder

1. Open VS Code.
2. Run: `Developer: Install Extension from Location...`
3. Select `vscode-nunchucks`.

### Option 2: Install from VSIX

1. Build package:
   - `npx vsce package`
2. In VS Code run:
   - `Extensions: Install from VSIX...`
3. Select the generated `.vsix` file.

## Usage

1. Open an `.njk` file.
2. Confirm language mode is `Nunchucks` (`nunchucks-template`).
3. If needed, force association in workspace settings:

```json
{
  "files.associations": {
    "*.njk": "nunchucks-template",
    "*.nunchucks": "nunchucks-template"
  }
}
```

## Recommended Workspace Styling

If you want HTML tags/attributes colored differently from Nunchucks tags, keep `.njk` on `nunchucks-template` and tune `editor.tokenColorCustomizations` in your workspace.

## Publish / Release

1. Bump version in `vscode-nunchucks/package.json`
2. Package:
   - `npx vsce package`
3. Publish:
   - `npx vsce login <publisher>`
   - `npx vsce publish`

If publish fails with PAT permission errors, create a new Marketplace token with publisher-scoped publish permissions and re-run `vsce login`.

## License

BSD-2-Clause. See `vscode-nunchucks/LICENSE`.
