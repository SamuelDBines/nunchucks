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
