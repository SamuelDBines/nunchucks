import path from 'path';
import { loadNunchucksWasm } from '../../go/wasm/loader.js';
import express from 'express';

const PORT = 13300;

const app = express();

const files = {
	'index.njk': `Hello {{ username }}`,
	'about.html': `<h1>About</h1><p>{{ user.name }}</p>`,
	'test.njk': `{% for i in items %}<li>{{ i }}</li>{% endfor %}`,
	'index.json': `{"user":"{{ user }}"}`,
	'index.yaml': `user: {{ user }}`,
};

const nc = await loadNunchucksWasm({
	wasmURL: path.resolve(__dirname, '../../go/wasm/nunchucks.wasm'),
	goURL: path.resolve(__dirname, '../../go/wasm/wasm_exec.js'),
});


// app

app.use(express.static(__dirname));

app.use(function (req, res, next) {
	res.locals.user = 'hello';
	next();
});

app.get('/', function (req, res) {
	const out = nc.renderFromMap({
		template: 'index.njk',
		files,
		context: { username: 'James Long <strong>copyright</strong>' },
	});
	res.type('html').send(out);
});
app.get('/index.json', function (req, res) {
	const out = nc.renderFromMap({ template: 'index.json', files, context: { user: 'Sam' } });
	res.type('application/json').send(out);
});
app.get('/index.yaml', function (req, res) {
	const out = nc.renderFromMap({ template: 'index.yaml', files, context: { user: 'Sam' } });
	res.type('text/yaml').send(out);
});

app.get('/about', function (req, res) {
	const out = nc.renderFromMap({ template: 'about.html', files, context: { items: [1, 2, 4], user: { name: 'tony' } } });
	res.type('html').send(out);
});

app.get('/test-one', function (req, res) {
	const out = nc.renderFromMap({ template: 'test.njk', files, context: { items: [1, 2, 4] } });
	res.type('html').send(out);
});

app.get('/test', function (req, res) {
	const out = nc.renderFromMap({ template: 'test.njk', files, context: { items: [1, 2, 4] } });
	res.type('html').send(out);
});

app.get('/{*splat}', function (req, res) {
	console.log(req.params.splat);
	const file = req.params.splat.join('/');
	if (!files[file]) {
		res.status(404).send('not found');
		return;
	}
	const out = nc.renderFromMap({ template: file, files, context: { user: 'Sam' } });
	res.type(file.endsWith('.json') ? 'application/json' : file.endsWith('.yaml') ? 'text/yaml' : 'html').send(out);
});

app.listen(PORT, function () {
	console.log('Express server running on http://localhost:' + PORT);
});
