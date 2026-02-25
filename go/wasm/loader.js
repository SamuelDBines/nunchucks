/**
 * Go WASM-backed nunchucks runtime.
 * The JS package is a thin wrapper around Go + WASM, not a separate TS engine.
 */
export async function loadNunchucksWasm({ wasmURL, goURL = "./wasm_exec.js" }) {
  if (!wasmURL) throw new Error("wasmURL is required");

  await import(goURL);
  if (!globalThis.Go) throw new Error("Go WASM runtime (Go class) not found after loading wasm_exec.js");

  const go = new globalThis.Go();

  const response = await fetch(wasmURL);
  if (!response.ok) throw new Error(`failed to fetch wasm: ${response.status} ${response.statusText}`);

  let instance;
  if (WebAssembly.instantiateStreaming) {
    try {
      ({ instance } = await WebAssembly.instantiateStreaming(response, go.importObject));
    } catch (_err) {
      const bytes = await response.arrayBuffer();
      ({ instance } = await WebAssembly.instantiate(bytes, go.importObject));
    }
  } else {
    const bytes = await response.arrayBuffer();
    ({ instance } = await WebAssembly.instantiate(bytes, go.importObject));
  }

  go.run(instance);

  if (!globalThis.NunchucksWasm) {
    throw new Error("NunchucksWasm API not found");
  }

  const parse = (raw) => {
    const out = JSON.parse(String(raw));
    if (!out.ok) throw new Error(out.error || "wasm error");
    return out.output || "";
  };

  const api = {
    renderFromMap: (request) =>
      parse(globalThis.NunchucksWasm.renderFromMap(JSON.stringify(request))),
    renderString: (request) =>
      parse(globalThis.NunchucksWasm.renderString(JSON.stringify(request))),
  };

  return {
    ...api,
    configure: ({ files = {} } = {}) => ({
      render: (template, context = {}) =>
        api.renderFromMap({ template, files, context }),
      renderString: (source, context = {}) =>
        api.renderString({ source, context }),
    }),
  };
}

export default loadNunchucksWasm;
