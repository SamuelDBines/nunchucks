# Nunchucks Go WASM

Build:

```bash
cd go
GOOS=js GOARCH=wasm go build -o wasm/nunchucks.wasm ./cmd/nunchucks-wasm
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" wasm/wasm_exec.js
```

Browser/Node loader:

```js
import { loadNunchucksWasm } from './loader.js';

const nc = await loadNunchucksWasm({
  wasmURL: './nunchucks.wasm',
  goURL: './wasm_exec.js',
});

const html = nc.renderFromMap({
  template: 'child.njk',
  files: {
    'base.njk': '<h1>{% block title %}Base{% endblock %}</h1> {% block body %}x{% endblock %}',
    'child.njk': '{% extends "base.njk" %}{% block title %}Hello{% endblock %}{% block body %}{{ name }}{% endblock %}',
  },
  context: { name: 'world' },
});
```

Exposed WASM API on `globalThis.NunchucksWasm`:

- `renderFromMap(requestJson: string) => responseJson: string`
- `renderString(requestJson: string) => responseJson: string`
