const state = {
	components: [],
	filtered: [],
	selected: '',
	search: '',
	theme: 'system',
	snippet: '',
	testHtml: '',
	activeTab: 'preview',
};

const groupOrder = [
	'Base Components',
	'Form Components',
	'Date & Time',
	'Feedback & Overlay',
	'Navigation',
	'Collections & Data',
	'Media & Visual',
	'Editor',
	'Canvas',
	'Charts',
	'Other',
];

const el = {
	componentSearch: document.getElementById('componentSearch'),
	componentGroups: document.getElementById('componentGroups'),
	themeSelect: document.getElementById('themeSelect'),
	themeStylesheet: document.getElementById('themeStylesheet'),
	selectedTag: document.getElementById('selectedTag'),
	editorTitle: document.getElementById('editorTitle'),
	snippetEditor: document.getElementById('snippetEditor'),
	testHtmlEditor: document.getElementById('testHtmlEditor'),
	previewFrame: document.getElementById('previewFrame'),
	openTestPage: document.getElementById('openTestPage'),
	previewPane: document.getElementById('previewPane'),
	testPane: document.getElementById('testPane'),
	refreshPreview: document.getElementById('refreshPreview'),
	tabs: Array.from(document.querySelectorAll('.tab')),
};

function titleize(value) {
	return value
		.split('-')
		.map((part) => part.charAt(0).toUpperCase() + part.slice(1))
		.join(' ');
}

function groupFor(component) {
	const slug = component.toLowerCase();
	const has = (...tokens) => tokens.some((token) => slug.includes(token));

	if (has('editor')) return 'Editor';
	if (has('canvas')) return 'Canvas';
	if (has('chart')) return 'Charts';
	if (has('date', 'calendar', 'timeline', 'countdown')) return 'Date & Time';
	if (has('input', 'field', 'select', 'checkbox', 'radio', 'toggle', 'search', 'file', 'otp', 'combobox', 'autocomplete', 'range', 'number')) return 'Form Components';
	if (has('modal', 'dialog', 'drawer', 'toast', 'message', 'error', 'progress', 'loading', 'skeleton')) return 'Feedback & Overlay';
	if (has('nav', 'menu', 'breadcrumb', 'tab', 'pagination', 'toolbar', 'link', 'anchor', 'action')) return 'Navigation';
	if (has('table', 'list', 'grid', 'tree', 'filter', 'key-value', 'data')) return 'Collections & Data';
	if (has('avatar', 'image', 'icon', 'video', 'carousel', 'map')) return 'Media & Visual';
	if (has('button', 'card', 'divider', 'dropdown', 'edge', 'fab', 'hero', 'tag', 'split', 'wizard', 'accordion', 'empty-state', 'metric')) return 'Base Components';
	return 'Other';
}

function defaultSnippet(component) {
	const tag = `bf-${component}`;
	const samples = {
		button: `<bf-button>Primary</bf-button>`,
		card: `<bf-card>\n  <div stack="sm">\n    <h3 typography="h3">Card Title</h3>\n    <p typography="body">Card content goes here.</p>\n    <bf-button variant="secondary">Action</bf-button>\n  </div>\n</bf-card>`,
		dropdown: `<bf-dropdown>\n  <bf-button slot="trigger">Actions</bf-button>\n  <button slot="content" item value="edit">Edit</button>\n  <button slot="content" item value="share">Share</button>\n  <button slot="content" item value="delete">Delete</button>\n</bf-dropdown>`,
		editor: `<bf-editor code value="function hello(name) {\\n  return 'Hello ' + name;\\n}\\n\\nconsole.log(hello('Bareframe'))"></bf-editor>`,
		menu: `<bf-menu>\n  <div item value="open">Open</div>\n  <div item value="rename">Rename</div>\n  <div item value="archive">Archive</div>\n</bf-menu>`,
		modal: `<div stack="sm">\n  <bf-button bf-open="demo-modal">Open Modal</bf-button>\n  <bf-modal id="demo-modal">\n    <h3>Project settings</h3>\n    <p>Centered modal content.</p>\n    <bf-button bf-close="demo-modal">Close</bf-button>\n  </bf-modal>\n</div>`,
		dialog: `<div stack="sm">\n  <bf-button bf-open="demo-dialog">Open Dialog</bf-button>\n  <bf-dialog id="demo-dialog" variant="panel" position="right">\n    <h3>Panel dialog</h3>\n    <bf-button bf-close="demo-dialog">Close</bf-button>\n  </bf-dialog>\n</div>`,
		drawer: `<div stack="sm">\n  <bf-button bf-open="demo-drawer">Open Drawer</bf-button>\n  <bf-drawer id="demo-drawer" side="right">\n    <h3>Filters</h3>\n    <bf-button bf-close="demo-drawer">Close</bf-button>\n  </bf-drawer>\n</div>`,
		progress: `<div stack="md">\n  <bf-progress value="62" max="100" striped primary>Upload <value /></bf-progress>\n  <bf-progress value="35" max="100" secondary>Processing <value /></bf-progress>\n</div>`,
		chart: `<div row>\n  <div col="6"><bf-chart variant="bar">Revenue</bf-chart></div>\n  <div col="6"><bf-chart variant="line">Trend</bf-chart></div>\n</div>`,
	};
	return samples[component] || `<${tag}></${tag}>`;
}

function setEditorValue(editor, value) {
	editor.setAttribute('value', value);
}

function tabTo(view) {
	state.activeTab = view;
	for (const tab of el.tabs) {
		tab.classList.toggle('is-active', tab.dataset.tab === view);
	}
	el.previewPane.hidden = view !== 'preview';
	el.testPane.hidden = view !== 'test';
}

function renderPreview() {
	const html = `<!doctype html>
<html lang="en">
	<head>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1" />
		<link rel="stylesheet" href="./bareframe/themes/${state.theme}.css" />
		<style>
			body {
				margin: 0;
				padding: 1rem;
				font-family: 'Avenir Next', 'Segoe UI', sans-serif;
				background: var(--bf-theme-surface-2, #f8fafc);
				color: var(--bf-theme-text-1, #111827);
			}
		</style>
	</head>
	<body>
${state.snippet}
		<script type="module" src="./bareframe/dist/index.js"></script>
	</body>
</html>`;
	el.previewFrame.srcdoc = html;
}

function debounce(fn, wait = 220) {
	let timer = null;
	return (...args) => {
		window.clearTimeout(timer);
		timer = window.setTimeout(() => fn(...args), wait);
	};
}

const debouncedPreview = debounce(renderPreview, 180);

async function loadTestHtml(component) {
	const path = `./bareframe/public/${component}-test.html`;
	el.openTestPage.href = path;
	try {
		const res = await fetch(path);
		if (!res.ok) {
			throw new Error(`failed to load ${path}`);
		}
		const text = await res.text();
		state.testHtml = text;
		setEditorValue(el.testHtmlEditor, text);
	} catch (err) {
		const message = `<!-- ${String(err)} -->`;
		state.testHtml = message;
		setEditorValue(el.testHtmlEditor, message);
	}
}

function renderGroups() {
	const grouped = new Map();
	for (const group of groupOrder) grouped.set(group, []);
	for (const component of state.filtered) {
		const group = groupFor(component);
		if (!grouped.has(group)) grouped.set(group, []);
		grouped.get(group).push(component);
	}

	el.componentGroups.innerHTML = '';
	for (const [groupName, items] of grouped.entries()) {
		if (!items.length) continue;
		const details = document.createElement('details');
		details.className = 'group';
		details.open = true;

		const summary = document.createElement('summary');
		summary.textContent = `${groupName} (${items.length})`;
		details.append(summary);

		const list = document.createElement('div');
		list.className = 'group-list';
		for (const component of items.sort()) {
			const btn = document.createElement('button');
			btn.type = 'button';
			btn.className = 'component-btn';
			btn.textContent = `bf-${component}`;
			btn.classList.toggle('is-active', state.selected === component);
			btn.addEventListener('click', async () => {
				state.selected = component;
				state.snippet = defaultSnippet(component);
				el.editorTitle.textContent = `${titleize(component)} Editor`;
				el.selectedTag.textContent = `bf-${component}`;
				setEditorValue(el.snippetEditor, state.snippet);
				renderGroups();
				await loadTestHtml(component);
				renderPreview();
			});
			list.append(btn);
		}
		details.append(list);
		el.componentGroups.append(details);
	}
}

function applySearch() {
	const q = state.search.trim().toLowerCase();
	if (!q) {
		state.filtered = [...state.components];
		renderGroups();
		return;
	}
	state.filtered = state.components.filter((component) => {
		return component.includes(q) || `bf-${component}`.includes(q);
	});
	renderGroups();
}

async function boot() {
	const manifestRes = await fetch('./bareframe/dist/manifest.json');
	const manifest = await manifestRes.json();
	state.components = (manifest.components || []).slice().sort();
	state.filtered = [...state.components];
	state.selected = state.components.includes('button') ? 'button' : state.components[0] || '';
	state.snippet = defaultSnippet(state.selected);

	el.editorTitle.textContent = `${titleize(state.selected)} Editor`;
	el.selectedTag.textContent = `bf-${state.selected}`;
	setEditorValue(el.snippetEditor, state.snippet);
	renderGroups();
	await loadTestHtml(state.selected);
	renderPreview();
}

el.componentSearch.addEventListener('input', (event) => {
	state.search = event.target.value || '';
	applySearch();
});

el.themeSelect.addEventListener('change', (event) => {
	state.theme = event.target.value;
	el.themeStylesheet.href = `./bareframe/themes/${state.theme}.css`;
	renderPreview();
});

el.refreshPreview.addEventListener('click', () => renderPreview());

el.tabs.forEach((tab) => {
	tab.addEventListener('click', () => tabTo(tab.dataset.tab));
});

el.snippetEditor.addEventListener('bf-change', (event) => {
	state.snippet = event.detail.value || '';
	debouncedPreview();
});

boot().catch((err) => {
	const message = `Failed to initialize workbench: ${String(err)}`;
	el.componentGroups.textContent = message;
	setEditorValue(el.testHtmlEditor, `<!-- ${message} -->`);
});
