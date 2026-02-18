export async function loadNunchucksWasm({ wasmURL, goURL = './wasm_exec.js' }) {
  if (!wasmURL) throw new Error('wasmURL is required');

  await import(goURL);
  const go = new globalThis.Go();

  const result = await WebAssembly.instantiateStreaming(fetch(wasmURL), go.importObject);
  go.run(result.instance);

  if (!globalThis.NunchucksWasm) {
    throw new Error('NunchucksWasm API not found');
  }

  const parse = (raw) => {
    const out = JSON.parse(String(raw));
    if (!out.ok) throw new Error(out.error || 'wasm error');
    return out.output || '';
  };

  return {
    renderFromMap: (request) => parse(globalThis.NunchucksWasm.renderFromMap(JSON.stringify(request))),
    renderString: (request) => parse(globalThis.NunchucksWasm.renderString(JSON.stringify(request))),
  };
}
