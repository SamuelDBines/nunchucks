import path from 'path';
import fs from 'node:fs';
import { loadNunchucksWasm } from '../../go/wasm/loader.js';

const viewsDir = path.join(__dirname, 'views');
const outDir = path.join(__dirname, 'out');

const files: Record<string, string> = {};
const walk = (dir: string) => {
	for (const ent of fs.readdirSync(dir, { withFileTypes: true })) {
		const full = path.join(dir, ent.name);
		if (ent.isDirectory()) {
			walk(full);
			continue;
		}
		const rel = path.relative(viewsDir, full).replaceAll(path.sep, '/');
		files[rel] = fs.readFileSync(full, 'utf-8');
	}
};
walk(viewsDir);

const nc = await loadNunchucksWasm({
	wasmURL: path.resolve(__dirname, '../../go/wasm/nunchucks.wasm'),
	goURL: path.resolve(__dirname, '../../go/wasm/wasm_exec.js'),
});

fs.mkdirSync(outDir, { recursive: true });
for (const name of Object.keys(files)) {
	const out = nc.renderFromMap({
		template: name,
		files,
		context: { items: [1, 2, 4], user: 'Sam' },
	});
	fs.writeFileSync(path.join(outDir, name), out);
}
