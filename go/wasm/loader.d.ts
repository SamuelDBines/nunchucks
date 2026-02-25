export type RenderFromMapRequest = {
  template: string;
  files: Record<string, string>;
  context?: Record<string, any>;
};

export type RenderStringRequest = {
  source: string;
  context?: Record<string, any>;
};

export type WasmRuntime = {
  renderFromMap(request: RenderFromMapRequest): string;
  renderString(request: RenderStringRequest): string;
  configure(opts?: { files?: Record<string, string> }): {
    render(template: string, context?: Record<string, any>): string;
    renderString(source: string, context?: Record<string, any>): string;
  };
};

export function loadNunchucksWasm(opts: {
  wasmURL: string;
  goURL?: string;
}): Promise<WasmRuntime>;

export default loadNunchucksWasm;
