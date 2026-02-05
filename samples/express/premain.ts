import path from 'path';
import * as nunjucks from '../../src';

const viewsDir = path.join(__dirname, 'views');
const outDir = path.join(__dirname, 'out');

const nunev = nunjucks.configure({
	path: viewsDir,
	dev: true,
	watch: true,
	devRefresh: true,
	detectExtensions: true
});

nunev.precompile(outDir)
