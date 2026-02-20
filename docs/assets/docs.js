(function () {
  const tabs = Array.from(document.querySelectorAll('.lang-tab'));
  const panels = Array.from(document.querySelectorAll('.lang-panel'));
  const fileTabs = Array.from(document.querySelectorAll('.file-tab'));
  const filePanels = Array.from(document.querySelectorAll('.file-panel'));
  const projectPane = document.querySelector('[data-project-pane]');
  const projectOpen = document.querySelector('[data-open-project]');
  const projectClose = document.querySelector('[data-close-project]');
  const projectBackdrop = document.querySelector('[data-project-backdrop]');
  const paneTabs = Array.from(document.querySelectorAll('[data-pane-tab]'));
  const panes = Array.from(document.querySelectorAll('[data-pane]'));
  const playground = document.querySelector('[data-playground]');

  function activate(lang) {
    tabs.forEach((tab) => {
      const on = tab.dataset.lang === lang;
      tab.classList.toggle('is-active', on);
      tab.setAttribute('aria-selected', on ? 'true' : 'false');
    });

    panels.forEach((panel) => {
      const on = panel.dataset.lang === lang;
      panel.classList.toggle('is-active', on);
      panel.hidden = !on;
    });
  }

  function activateFile(file) {
    fileTabs.forEach((tab) => {
      const on = tab.dataset.file === file;
      tab.classList.toggle('is-active', on);
      tab.setAttribute('aria-selected', on ? 'true' : 'false');
    });

    filePanels.forEach((panel) => {
      const on = panel.dataset.file === file;
      panel.classList.toggle('is-active', on);
      panel.hidden = !on;
    });
  }

  function closeProjectPane() {
    if (!projectPane || !projectBackdrop) return;
    projectPane.classList.remove('is-open');
    projectBackdrop.hidden = true;
  }

  function openProjectPane() {
    if (!projectPane || !projectBackdrop) return;
    projectPane.classList.add('is-open');
    projectBackdrop.hidden = false;
  }

  function activatePane(pane) {
    paneTabs.forEach((tab) => {
      const on = tab.dataset.paneTab === pane;
      tab.classList.toggle('is-active', on);
      tab.setAttribute('aria-selected', on ? 'true' : 'false');
    });
    panes.forEach((section) => {
      const on = section.dataset.pane === pane;
      section.classList.toggle('is-active', on);
    });
  }

  tabs.forEach((tab) => {
    tab.addEventListener('click', function () {
      activate(tab.dataset.lang);
    });
  });

  fileTabs.forEach((tab) => {
    tab.addEventListener('click', function () {
      activateFile(tab.dataset.file);
    });
  });

  if (projectOpen) {
    projectOpen.addEventListener('click', openProjectPane);
  }
  if (projectClose) {
    projectClose.addEventListener('click', closeProjectPane);
  }
  if (projectBackdrop) {
    projectBackdrop.addEventListener('click', closeProjectPane);
  }

  paneTabs.forEach((tab) => {
    tab.addEventListener('click', function () {
      activatePane(tab.dataset.paneTab);
    });
  });

  if (playground) {
    const tabsHost = playground.querySelector('[data-pg-tabs]');
    const editorHost = playground.querySelector('[data-pg-editor]');
    const output = playground.querySelector('[data-pg-output]');
    const outputTabsHost = playground.querySelector('[data-pg-output-tabs]');
    const addBtn = playground.querySelector('[data-add-file]');
    const exportBtn = playground.querySelector('[data-export-output]');
    const modal = document.querySelector('[data-pg-modal]');
    const modalBackdrop = document.querySelector('[data-pg-modal-backdrop]');
    const modalClose = document.querySelector('[data-pg-modal-close]');
    const modalCancel = document.querySelector('[data-pg-modal-cancel]');
    const modalCreate = document.querySelector('[data-pg-modal-create]');
    const fileTypeInput = document.querySelector('[data-pg-file-type]');
    const fileNameInput = document.querySelector('[data-pg-file-name]');
    const jsonNode = document.getElementById('pg-initial-files');

    let files = {};
    let activeFile = '';
    let activeOutputFile = '';
    let editor = null;
    let renderSeq = 0;

    try {
      files = JSON.parse(jsonNode ? jsonNode.textContent || '{}' : '{}');
    } catch (_err) {
      files = {};
    }

    const names = Object.keys(files);
    if (names.length === 0) {
      files['app.njk'] = '';
    }
    activeFile = Object.keys(files)[0];
    activeOutputFile = activeFile;

    if (window.ace && editorHost) {
      editor = window.ace.edit(editorHost);
      editor.setTheme('ace/theme/tomorrow_night');
      editor.session.setMode('ace/mode/django');
      editor.setShowPrintMargin(false);
      editor.setOption('fontSize', '13px');
      editor.setOption('wrap', true);
    }

    function fileTypeClass(name) {
      if (name.endsWith('.njk')) return 'is-njk';
      if (name.endsWith('.html')) return 'is-html';
      return 'is-other';
    }

    function modeForFile(name) {
      if (name.endsWith('.html')) return 'ace/mode/html';
      if (name.endsWith('.json')) return 'ace/mode/json';
      if (name.endsWith('.js') || name.endsWith('.ts')) return 'ace/mode/javascript';
      return 'ace/mode/django';
    }

    function setEditorText(text) {
      if (editor) {
        editor.session.setMode(modeForFile(activeFile));
        editor.setValue(text, -1);
        return;
      }
      if (editorHost) {
        editorHost.textContent = text;
      }
    }

    function getEditorText() {
      if (editor) return editor.getValue();
      return editorHost ? editorHost.textContent || '' : '';
    }

    function renderTabs() {
      tabsHost.innerHTML = '';
      Object.keys(files).forEach((name) => {
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'pg-tab' + (name === activeFile ? ' is-active' : '') + ' ' + fileTypeClass(name);
        btn.setAttribute('role', 'tab');
        btn.setAttribute('aria-selected', name === activeFile ? 'true' : 'false');
        btn.dataset.file = name;

        const base = document.createElement('span');
        base.className = 'pg-tab-name';
        base.textContent = name;
        btn.appendChild(base);

        btn.addEventListener('click', function () {
          activeFile = name;
          setEditorText(files[name]);
          renderTabs();
          activeOutputFile = name;
          renderOutputTabs();
          updateOutput();
        });
        tabsHost.appendChild(btn);
      });
    }

    function renderOutputTabs() {
      outputTabsHost.innerHTML = '';
      const names = Object.keys(files);
      names.forEach((name) => {
        const btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'pg-out-tab' + (name === activeOutputFile ? ' is-active' : '');
        btn.setAttribute('role', 'tab');
        btn.setAttribute('aria-selected', name === activeOutputFile ? 'true' : 'false');
        btn.textContent = name;
        btn.addEventListener('click', function () {
          activeOutputFile = name;
          renderOutputTabs();
          updateOutput();
        });
        outputTabsHost.appendChild(btn);
      });
    }

    function titleCase(s) {
      return String(s).replace(/\w\S*/g, (w) => w[0].toUpperCase() + w.slice(1).toLowerCase());
    }

    function renderUserCard(tpl, user) {
      return tpl
        .replace(/\{\{\s*user\.name\s*\|\s*title\s*\}\}/g, titleCase(user.name))
        .replace(/\{\{\s*user\.email\s*\|\s*lower\s*\}\}/g, String(user.email).toLowerCase());
    }

    function renderPanel(tpl, kind, body) {
      return tpl
        .replace(/\{\{\s*kind\s*\}\}/g, kind)
        .replace(/\{\{\s*caller\(\)\s*\}\}/g, body);
    }

    function extractBlock(src, name) {
      const re = new RegExp('\\{%-?\\s*block\\s+' + name + '\\s*-?%\\}([\\s\\S]*?)\\{%-?\\s*endblock\\s*-?%\\}');
      const m = src.match(re);
      return m ? m[1].trim() : '';
    }

    function renderAppLikeOutputs() {
      const users = [
        { name: 'sam', email: 'Sam@Example.com' },
        { name: 'kai', email: 'Kai@Example.com' }
      ];
      const app = files['app.njk'] || '';
      const layout = files['layout.njk'] || '<html><body>{% block body %}{% endblock %}</body></html>';
      const card = files['partials/user-card.njk'] || '<article>{{ user.name }}</article>';
      const macros = files['macros/ui.njk'] || '<section>{{ caller() }}</section>';

      let body = extractBlock(app, 'body');
      const title = extractBlock(app, 'title') || 'Demo';

      const cards = users.map((u) => renderUserCard(card, u)).join('\n');
      body = body.replace(/\{%\s*for\s+user\s+in\s+users\s*%\}[\s\S]*?\{%\s*endfor\s*%\}/g, cards);

      body = body.replace(/\{%\s*call\s+ui\.panel\(\"([^\"]+)\"\)\s*%\}([\s\S]*?)\{%\s*endcall\s*%\}/g, (_m, kind, panelBody) => {
        const panelTpl = macros.match(/\{%\s*macro\s+panel\([^\)]*\)\s*%\}([\s\S]*?)\{%\s*endmacro\s*%\}/);
        const tpl = panelTpl ? panelTpl[1].trim() : '<section class="panel panel-{{ kind }}">{{ caller() }}</section>';
        return renderPanel(tpl, kind, panelBody.replace(/\{\{\s*users\s*\|\s*length\s*\}\}/g, String(users.length)).trim());
      });

      body = body
        .replace(/\{%\s*include[\s\S]*?%\}/g, '')
        .replace(/\{%\s*set[\s\S]*?%\}/g, '')
        .replace(/\{%\s*import[\s\S]*?%\}/g, '')
        .replace(/\{\{\s*users\s*\|\s*length\s*\}\}/g, String(users.length))
        .trim();

      const html = layout
        .replace(/\{%\s*block\s+title\s*%\}[\s\S]*?\{%\s*endblock\s*%\}/g, title)
        .replace(/\{%\s*block\s+body\s*%\}[\s\S]*?\{%\s*endblock\s*%\}/g, body)
        .trim();

      const out = {};
      Object.keys(files).forEach((name) => {
        out[name] = files[name];
      });
      out['app.njk'] = html;
      out['layout.njk'] = html;
      if (files['partials/user-card.njk']) {
        out['partials/user-card.njk'] = renderUserCard(card, users[0]);
      }
      if (files['macros/ui.njk']) {
        out['macros/ui.njk'] = renderPanel('<section class=\"panel panel-{{ kind }}\">{{ caller() }}</section>', 'info', 'sample');
      }
      return out;
    }

    function applyOutput(rendered) {
      const keys = Object.keys(rendered);
      const name = activeOutputFile in rendered ? activeOutputFile : keys[0];
      const text = rendered[name] || '';
      output.className = /layout\.njk|app\.njk|user-card\.njk/.test(name) ? 'language-html' : 'language-django';
      output.textContent = text;
      if (window.hljs && typeof window.hljs.highlightElement === 'function') {
        window.hljs.highlightElement(output);
      }
    }

    async function fetchServerOutputs() {
      const api = window.NUNCHUCKS_PLAYGROUND_API || '/api/playground/render';
      try {
        const res = await fetch(api, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            template: 'app.njk',
            files,
            context: {
              users: [
                { name: 'sam', email: 'Sam@Example.com' },
                { name: 'kai', email: 'Kai@Example.com' }
              ]
            }
          })
        });
        if (!res.ok) return null;
        const body = await res.json();
        if (!body || typeof body !== 'object' || !body.outputs) return null;
        return body.outputs;
      } catch (_err) {
        return null;
      }
    }

    function updateOutput() {
      const seq = ++renderSeq;
      const local = renderAppLikeOutputs();
      applyOutput(local);
      fetchServerOutputs().then((remote) => {
        if (!remote || seq !== renderSeq) return;
        applyOutput(remote);
      });
    }

    if (editor) {
      editor.session.on('change', function () {
        files[activeFile] = getEditorText();
        updateOutput();
      });
    }

    function openModal() {
      if (!modal || !modalBackdrop) return;
      modal.hidden = false;
      modalBackdrop.hidden = false;
      if (fileNameInput) fileNameInput.focus();
    }

    function closeModal() {
      if (!modal || !modalBackdrop) return;
      modal.hidden = true;
      modalBackdrop.hidden = true;
      if (fileNameInput) fileNameInput.value = '';
      if (fileTypeInput) fileTypeInput.value = 'page';
    }

    function normalizeNewFileName(raw, t) {
      let name = (raw || '').trim();
      if (name === '') return '';
      if (!name.endsWith('.njk')) name += '.njk';
      if (name.includes('/')) return name;
      if (t === 'layout') return 'layouts/' + name;
      if (t === 'partial') return 'partials/' + name;
      if (t === 'macro') return 'macros/' + name;
      if (t === 'page') return 'pages/' + name;
      return name;
    }

    function createTemplateByType(t, file) {
      switch (t) {
      case 'layout':
        return '<!doctype html>\\n<html>\\n  <head><title>{% block title %}' + file + '{% endblock %}</title></head>\\n  <body>\\n    {% block body %}{% endblock %}\\n  </body>\\n</html>\\n';
      case 'partial':
        return '<section class=\"partial\">\\n  {{ value | default(\"partial\", true) }}\\n</section>\\n';
      case 'macro':
        return '{% macro sample(text) %}\\n<span>{{ text }}</span>\\n{% endmacro %}\\n';
      case 'page':
        return '{% extends \"layout.njk\" %}\\n{% block title %}Page{% endblock %}\\n{% block body %}\\n  New page\\n{% endblock %}\\n';
      default:
        return '{# new file #}\\n';
      }
    }

    if (addBtn) {
      addBtn.addEventListener('click', function () {
        openModal();
      });
    }

    if (modalBackdrop) {
      modalBackdrop.addEventListener('click', closeModal);
    }
    if (modalClose) {
      modalClose.addEventListener('click', closeModal);
    }
    if (modalCancel) {
      modalCancel.addEventListener('click', closeModal);
    }
    if (modalCreate) {
      modalCreate.addEventListener('click', function () {
        const t = fileTypeInput ? fileTypeInput.value : 'page';
        const name = normalizeNewFileName(fileNameInput ? fileNameInput.value : '', t);
        if (!name) return;
        if (files[name]) return;
        files[name] = createTemplateByType(t, name);
        activeFile = name;
        activeOutputFile = name;
        setEditorText(files[name]);
        renderTabs();
        renderOutputTabs();
        updateOutput();
        closeModal();
      });
    }

    if (exportBtn) {
      exportBtn.addEventListener('click', function () {
        const outName = activeOutputFile || 'output.html';
        const blob = new Blob([output.textContent || ''], { type: 'text/plain;charset=utf-8' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = outName.replace(/[\\/]/g, '_') + '.out';
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      });
    }

    setEditorText(files[activeFile] || '');
    renderTabs();
    renderOutputTabs();
    updateOutput();
  }

  function applyHighlighting() {
    if (window.hljs && typeof window.hljs.highlightAll === 'function') {
      window.hljs.highlightAll();
      return true;
    }
    return false;
  }

  if (!applyHighlighting()) {
    window.addEventListener('load', applyHighlighting, { once: true });
  }
})();
