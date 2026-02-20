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
    const editor = playground.querySelector('[data-pg-editor]');
    const output = playground.querySelector('[data-pg-output]');
    const addBtn = playground.querySelector('[data-add-file]');
    const jsonNode = document.getElementById('pg-initial-files');

    let files = {};
    let activeFile = '';

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

    function fileTypeClass(name) {
      if (name.endsWith('.njk')) return 'is-njk';
      if (name.endsWith('.html')) return 'is-html';
      return 'is-other';
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
          editor.value = files[name];
          renderTabs();
        });
        tabsHost.appendChild(btn);
      });
    }

    function simpleRender(template) {
      let out = template;
      out = out.replace(/\{\{\s*users\s*\|\s*length\s*\}\}/g, '2');
      out = out.replace(/\{\{\s*user\.name\s*\|\s*title\s*\}\}/g, 'Sam');
      out = out.replace(/\{\{\s*user\.email\s*\|\s*lower\s*\}\}/g, 'sam@example.com');
      out = out.replace(/\{\{\s*users\s*\|\s*length\s*\}\}/g, '2');
      out = out.replace(/\{%\s*extends[\s\S]*?%\}/g, '');
      out = out.replace(/\{%\s*import[\s\S]*?%\}/g, '');
      out = out.replace(/\{%\s*block[\s\S]*?%\}/g, '');
      out = out.replace(/\{%\s*endblock\s*%\}/g, '');
      out = out.replace(/\{%\s*for[\s\S]*?%\}/g, '');
      out = out.replace(/\{%\s*endfor\s*%\}/g, '');
      out = out.replace(/\{%\s*call[\s\S]*?%\}/g, '');
      out = out.replace(/\{%\s*endcall\s*%\}/g, '');
      out = out.replace(/\{%\s*include[\s\S]*?%\}/g, '<!-- include -->');
      out = out.replace(/\{\{[\s\S]*?\}\}/g, '');
      return out.trim();
    }

    function updateOutput() {
      const source = files['app.njk'] || files[activeFile] || '';
      output.textContent = simpleRender(source);
      if (window.hljs && typeof window.hljs.highlightElement === 'function') {
        window.hljs.highlightElement(output);
      }
    }

    editor.addEventListener('input', function () {
      files[activeFile] = editor.value;
      updateOutput();
    });

    if (addBtn) {
      addBtn.addEventListener('click', function () {
        const name = window.prompt('New file name (example: partials/new-card.njk)');
        if (!name) return;
        if (files[name]) {
          window.alert('File already exists');
          return;
        }
        files[name] = '{# new file #}\\n';
        activeFile = name;
        editor.value = files[name];
        renderTabs();
        updateOutput();
      });
    }

    editor.value = files[activeFile] || '';
    renderTabs();
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
