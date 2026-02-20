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
    const preview = playground.querySelector('[data-pg-preview]');
    const outputTabsHost = playground.querySelector('[data-pg-output-tabs]');
    const viewTabs = Array.from(playground.querySelectorAll('[data-pg-view-tab]'));
    const viewPanels = Array.from(playground.querySelectorAll('[data-pg-view]'));
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
    let activeOutputView = 'preview';
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

    function activateOutputView(view) {
      activeOutputView = view;
      viewTabs.forEach((tab) => {
        const on = tab.dataset.pgViewTab === view;
        tab.classList.toggle('is-active', on);
        tab.setAttribute('aria-selected', on ? 'true' : 'false');
      });
      viewPanels.forEach((panel) => {
        const on = panel.dataset.pgView === view;
        panel.classList.toggle('is-active', on);
        panel.hidden = !on;
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
      if (preview) {
        const asHtml = /layout\.njk|app\.njk|user-card\.njk/.test(name);
        preview.srcdoc = asHtml ? text : '<pre style="font-family:monospace;white-space:pre-wrap;padding:12px;">' + text.replace(/[&<>]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;' }[c])) + '</pre>';
      }
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

    viewTabs.forEach((tab) => {
      tab.addEventListener('click', function () {
        activateOutputView(tab.dataset.pgViewTab);
      });
    });

    setEditorText(files[activeFile] || '');
    renderTabs();
    renderOutputTabs();
    activateOutputView(activeOutputView);
    updateOutput();
  }

  function escapeHtml(s) {
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function highlightHtmlSegment(seg) {
    const escaped = escapeHtml(seg);
    return escaped.replace(/(&lt;\/?)([a-zA-Z][\w:-]*)([\s\S]*?)(&gt;)/g, function (_m, open, tagName, attrs, close) {
      const attrsHighlighted = attrs.replace(/([a-zA-Z_:][-a-zA-Z0-9_:.]*)(=)/g, '<span class="nc-html-attr">$1</span>$2');
      return open + '<span class="nc-html-tag">' + tagName + '</span>' + attrsHighlighted + close;
    });
  }

  function highlightTemplateExpr(expr) {
    let out = escapeHtml(expr);
    const stash = [];
    out = out.replace(/&quot;[^&]*?&quot;|'[^']*?'/g, function (m) {
      stash.push('<span class="nc-string">' + m + '</span>');
      return '@@NCSTR' + (stash.length - 1) + '@@';
    });

    out = out.replace(/\|\s*([a-zA-Z_][\w]*)/g, function (_m, name) {
      return '<span class="nc-operator">|</span><span class="nc-filter">' + name + '</span>';
    });
    out = out.replace(/\b([a-zA-Z_][\w]*)\s*(?=\()/g, '<span class="nc-func">$1</span>');
    out = out.replace(/\b\d+(?:\.\d+)?\b/g, '<span class="nc-number">$&</span>');
    out = out.replace(/\b(and|or|not|in|is)\b/g, '<span class="nc-operator">$1</span>');
    out = out.replace(/\b(if|else|true|false|none|null)\b/g, '<span class="nc-keyword">$1</span>');
    out = out.replace(/\b(extends|include|import|from|as|with|context|ignore|missing|elif|endif|for|endfor|set|endset|block|endblock|macro|endmacro|call|endcall|filter|endfilter|raw|endraw|verbatim|endverbatim|super)\b/g, '<span class="nc-keyword">$1</span>');

    out = out.replace(/@@NCSTR(\d+)@@/g, function (_m, i) {
      return stash[Number(i)] || '';
    });
    return out;
  }

  function highlightNunchucksSource(src) {
    const tokenRe = /(\{#[\s\S]*?#\}|\{%-?[\s\S]*?-?%\}|\{\{-?[\s\S]*?-?\}\})/g;
    let out = '';
    let pos = 0;
    let m;
    while ((m = tokenRe.exec(src)) !== null) {
      out += highlightHtmlSegment(src.slice(pos, m.index));
      const token = m[0];
      if (token.startsWith('{#')) {
        out += '<span class="nc-comment"><span class="nc-delim">{#</span>' +
          escapeHtml(token.slice(2, -2)) +
          '<span class="nc-delim">#}</span></span>';
      } else if (token.startsWith('{{')) {
        const open = token.startsWith('{{-') ? '{{-' : '{{';
        const close = token.endsWith('-}}') ? '-}}' : '}}';
        const inner = token.slice(open.length, token.length - close.length);
        out += '<span class="nc-delim">' + escapeHtml(open) + '</span>' +
          highlightTemplateExpr(inner) +
          '<span class="nc-delim">' + escapeHtml(close) + '</span>';
      } else {
        const open = token.startsWith('{%-') ? '{%-' : '{%';
        const close = token.endsWith('-%}') ? '-%}' : '%}';
        const inner = token.slice(open.length, token.length - close.length);
        out += '<span class="nc-delim">' + escapeHtml(open) + '</span>' +
          highlightTemplateExpr(inner) +
          '<span class="nc-delim">' + escapeHtml(close) + '</span>';
      }
      pos = tokenRe.lastIndex;
    }
    out += highlightHtmlSegment(src.slice(pos));
    return out;
  }

  function renderNunchucksBlock(codeEl, focusNeedle) {
    const raw = codeEl.dataset.ncRaw || codeEl.textContent || '';
    codeEl.dataset.ncRaw = raw;
    codeEl.classList.add('nunchucks-code');
    const rendered = highlightNunchucksSource(raw).split('\n');
    const lines = raw.split('\n');
    let focusIndex = -1;
    if (focusNeedle) {
      focusIndex = lines.findIndex(function (line) {
        return line.indexOf(focusNeedle) >= 0;
      });
    }
    codeEl.innerHTML = rendered.map(function (line, i) {
      const cls = 'nc-line' + (i === focusIndex ? ' is-focus' : '');
      return '<span class="' + cls + '" data-line="' + (i + 1) + '">' + (line || '&nbsp;') + '</span>';
    }).join('\n');
  }

  function applyNunchucksHighlighting() {
    const blocks = Array.from(document.querySelectorAll('code.language-django, code.language-nunchucks-template'));
    blocks.forEach(function (codeEl) {
      renderNunchucksBlock(codeEl);
    });
  }

  function initCoreTopicSync() {
    const core = document.getElementById('core-topics');
    if (!core || document.body.dataset.coreSyncInit === '1') return;
    document.body.dataset.coreSyncInit = '1';

    const topicMap = {
      'core-user-defined-templates-warning': { file: 'app', needle: '{# core dashboard section #}' },
      'core-file-extensions': { file: 'layout', needle: '<!doctype html>' },
      'core-syntax-highlighting': { file: 'app', needle: '{% import "macros/ui.njk" as ui %}' },
      'core-variables': { file: 'card', needle: '{{ user.name | title }}' },
      'core-filters': { file: 'app', needle: 'users | groupby("role", "none")' },
      'core-template-inheritance': { file: 'app', needle: '{% extends "layout.njk" %}' },
      'core-super': { file: 'app', needle: '{{ super() }}' },
      'core-keyword-arguments': { file: 'macros', needle: 'macro panel(kind="info")' },
      'core-comments': { file: 'app', needle: '{# core dashboard section #}' },
      'core-whitespace-control': { file: 'layout', needle: '{% block body %}{% endblock %}' },
      'core-autoescaping': { file: 'card', needle: 'user.website | urlize | safe' }
    };

    const ids = Object.keys(topicMap).filter(function (id) {
      return !!document.getElementById(id);
    });
    if (ids.length === 0) return;

    function syncTopic(id) {
      ids.forEach(function (topicId) {
        const el = document.getElementById(topicId);
        if (!el) return;
        el.classList.toggle('core-topic-active', topicId === id);
      });

      const cfg = topicMap[id];
      if (!cfg) return;
      activateFile(cfg.file);
      const activeTab = document.querySelector('.file-tab[data-file="' + cfg.file + '"]');
      if (activeTab && activeTab.scrollIntoView) {
        activeTab.scrollIntoView({ inline: 'center', block: 'nearest' });
      }

      const fileCodes = Array.from(document.querySelectorAll('.file-panel code.language-django'));
      fileCodes.forEach(function (codeEl) {
        renderNunchucksBlock(codeEl);
      });
      const target = document.querySelector('.file-panel[data-file="' + cfg.file + '"] code.language-django');
      if (target) {
        renderNunchucksBlock(target, cfg.needle);
      }
    }

    function activeIdFromScroll() {
      const marker = 180;
      let bestId = ids[0];
      let bestDistance = Number.POSITIVE_INFINITY;
      ids.forEach(function (id) {
        const el = document.getElementById(id);
        if (!el) return;
        const rect = el.getBoundingClientRect();
        const distance = Math.abs(rect.top - marker);
        if (distance < bestDistance) {
          bestDistance = distance;
          bestId = id;
        }
      });
      return bestId;
    }

    let ticking = false;
    function onScroll() {
      if (ticking) return;
      ticking = true;
      window.requestAnimationFrame(function () {
        syncTopic(activeIdFromScroll());
        ticking = false;
      });
    }

    window.addEventListener('scroll', onScroll, { passive: true });
    window.addEventListener('resize', onScroll);
    syncTopic(activeIdFromScroll());
  }

  function applyHighlighting() {
    if (window.hljs && typeof window.hljs.highlightAll === 'function') {
      window.hljs.highlightAll();
    }
    applyNunchucksHighlighting();
    initCoreTopicSync();
    return true;
  }

  if (!applyHighlighting()) {
    window.addEventListener('load', applyHighlighting, { once: true });
  }
})();
