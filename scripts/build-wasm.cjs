const { execSync } = require("node:child_process");
const path = require("node:path");
const fs = require("node:fs");

const root = process.cwd();
const goDir = path.join(root, "go");

function run(cmd) {
  execSync(cmd, { stdio: "inherit", cwd: goDir });
}

run("GOOS=js GOARCH=wasm go build -o wasm/nunchucks.wasm ./cmd/nunchucks-wasm");
const goRoot = execSync("go env GOROOT", { cwd: goDir }).toString().trim();
const candidates = [
  path.join(goRoot, "lib", "wasm", "wasm_exec.js"),
  path.join(goRoot, "misc", "wasm", "wasm_exec.js"),
];
const source = candidates.find((p) => fs.existsSync(p));
if (!source) {
  throw new Error(`wasm_exec.js not found under ${goRoot}`);
}
fs.copyFileSync(source, path.join(goDir, "wasm", "wasm_exec.js"));
