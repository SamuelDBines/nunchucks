(function () {
  const tabs = Array.from(document.querySelectorAll('.lang-tab'));
  const panels = Array.from(document.querySelectorAll('.lang-panel'));

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

  tabs.forEach((tab) => {
    tab.addEventListener('click', function () {
      activate(tab.dataset.lang);
    });
  });
})();
