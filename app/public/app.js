const fileTabs = document.getElementById("fileTabs");
const editorEl = document.getElementById("editor");
const templateName = document.getElementById("templateName");
const contextJson = document.getElementById("contextJson");
const previewFrame = document.getElementById("previewFrame");
const codeOut = document.getElementById("codeOut");
const addFileBtn = document.getElementById("addFile");
const saveFileBtn = document.getElementById("saveFile");
const deleteFileBtn = document.getElementById("deleteFile");
const resetWorkspaceBtn = document.getElementById("resetWorkspace");
const importWorkspaceBtn = document.getElementById("importWorkspace");
const exportWorkspaceBtn = document.getElementById("exportWorkspace");
const downloadOutputBtn = document.getElementById("downloadOutput");
const importWorkspaceInput = document.getElementById("importWorkspaceInput");
const renderBtn = document.getElementById("renderNow");
const viewButtons = Array.from(document.querySelectorAll(".view"));

const DB_NAME = "nunchucks-app-builder";
const DB_VERSION = 1;
const DB_STORE = "workspace";
const DB_KEY = "active-workspace";

let files = {};
let activeFile = "";
let editor = null;
let lastRenderedHtml = "";
let initialFiles = {};

function initEditor() {
  if (!window.ace || !editorEl) {
    throw new Error("Ace editor failed to load");
  }
  editor = window.ace.edit(editorEl);
  editor.setTheme("ace/theme/tomorrow_night");
  editor.session.setMode("ace/mode/django");
  editor.setShowPrintMargin(false);
  editor.setOption("fontSize", "13px");
  editor.setOption("wrap", false);
  editor.setOption("scrollPastEnd", 0.2);
}

function modeForFile(name) {
  if (name.endsWith(".html")) return "ace/mode/html";
  if (name.endsWith(".json")) return "ace/mode/json";
  if (name.endsWith(".js") || name.endsWith(".ts")) return "ace/mode/javascript";
  if (name.endsWith(".css")) return "ace/mode/css";
  if (name.endsWith(".yaml") || name.endsWith(".yml")) return "ace/mode/yaml";
  return "ace/mode/django";
}

function setEditorValue(value) {
  editor.session.setMode(modeForFile(activeFile));
  editor.setValue(String(value ?? ""), -1);
}

function getEditorValue() {
  return editor.getValue();
}

function setView(view) {
  viewButtons.forEach((btn) => btn.classList.toggle("is-active", btn.dataset.view === view));
  previewFrame.hidden = view !== "preview";
  codeOut.hidden = view !== "code";
}

function parseContext() {
  const raw = contextJson.value.trim();
  if (!raw) return {};
  return JSON.parse(raw);
}

function safeClone(value) {
  return JSON.parse(JSON.stringify(value));
}

function normalizeWorkspace(raw, defaults) {
  const next = {
    files: safeClone(defaults.files),
    activeFile: defaults.activeFile,
    templateName: defaults.templateName,
    contextJson: defaults.contextJson,
  };

  if (!raw || typeof raw !== "object") return next;

  if (raw.files && typeof raw.files === "object") {
    const normalizedFiles = {};
    for (const [name, content] of Object.entries(raw.files)) {
      if (!name || typeof name !== "string") continue;
      normalizedFiles[name] = String(content ?? "");
    }
    if (Object.keys(normalizedFiles).length > 0) {
      next.files = normalizedFiles;
    }
  }

  if (typeof raw.activeFile === "string" && next.files[raw.activeFile] != null) {
    next.activeFile = raw.activeFile;
  } else if (next.files[next.activeFile] == null) {
    next.activeFile = Object.keys(next.files)[0] || "app.njk";
  }

  if (typeof raw.templateName === "string" && raw.templateName.trim()) {
    next.templateName = raw.templateName.trim();
  } else {
    next.templateName = next.activeFile;
  }

  if (typeof raw.contextJson === "string") {
    next.contextJson = raw.contextJson;
  }

  return next;
}

function ensureDefaultGlobalFiles(workspace) {
  if (!workspace.files || typeof workspace.files !== "object") return workspace;
  if (workspace.files["_globals/links.njk"] == null) {
    workspace.files["_globals/links.njk"] = `<link id=\"themeStylesheet\" rel=\"stylesheet\" href=\"./bareframe/dist/themes/light.css\" />`;
  }
  if (workspace.files["_globals/scripts.njk"] == null) {
    workspace.files["_globals/scripts.njk"] = `<script type=\"module\" src=\"./bareframe/dist/index.js\"></script>`;
  }
  return workspace;
}

function openDB() {
  return new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onupgradeneeded = () => {
      const db = req.result;
      if (!db.objectStoreNames.contains(DB_STORE)) {
        db.createObjectStore(DB_STORE);
      }
    };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => reject(req.error || new Error("indexedDB open failed"));
  });
}

async function loadWorkspaceFromDB() {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(DB_STORE, "readonly");
    const store = tx.objectStore(DB_STORE);
    const req = store.get(DB_KEY);
    req.onsuccess = () => resolve(req.result || null);
    req.onerror = () => reject(req.error || new Error("indexedDB read failed"));
  });
}

async function saveWorkspaceToDB(payload) {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(DB_STORE, "readwrite");
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error || new Error("indexedDB write failed"));
    tx.objectStore(DB_STORE).put(payload, DB_KEY);
  });
}

async function clearWorkspaceDB() {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(DB_STORE, "readwrite");
    tx.oncomplete = () => resolve();
    tx.onerror = () => reject(tx.error || new Error("indexedDB clear failed"));
    tx.objectStore(DB_STORE).delete(DB_KEY);
  });
}

async function fetchInitialFiles() {
  const res = await fetch("/api/files");
  const body = await res.json();
  if (!body.ok) throw new Error(body.error || "failed to load initial files");
  const fromServer = body.files || {};
  if (Object.keys(fromServer).length === 0) {
    return { "app.njk": "" };
  }
  return fromServer;
}

function getWorkspaceSnapshot() {
  if (activeFile) {
    files[activeFile] = getEditorValue();
  }
  return {
    files,
    activeFile,
    templateName: templateName.value.trim() || activeFile,
    contextJson: contextJson.value,
  };
}

async function persistWorkspace() {
  await saveWorkspaceToDB(getWorkspaceSnapshot());
}

function renderFileTabs() {
  fileTabs.innerHTML = "";
  Object.keys(files).sort().forEach((name) => {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "tab" + (name === activeFile ? " is-active" : "");
    btn.textContent = name;
    btn.addEventListener("click", () => {
      files[activeFile] = getEditorValue();
      activeFile = name;
      setEditorValue(files[name] ?? "");
      templateName.value = activeFile;
      renderFileTabs();
      renderNow().catch((err) => window.alert(String(err)));
    });
    fileTabs.appendChild(btn);
  });
}

async function renderNow() {
  files[activeFile] = getEditorValue();
  const template = templateName.value.trim() || activeFile;
  const context = parseContext();

  const res = await fetch("/api/render", {
    method: "POST",
    headers: { "content-type": "application/json" },
    body: JSON.stringify({ template, files, context }),
  });
  const body = await res.json();
  if (!body.ok) throw new Error(body.error || "render failed");

  const out = body.outputs?.[template] || "";
  lastRenderedHtml = out;
  previewFrame.srcdoc = out;
  codeOut.textContent = out;
}

function downloadTextFile(filename, text, type = "text/plain") {
  const blob = new Blob([text], { type });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(url);
}

async function applyWorkspace(rawWorkspace) {
  const defaults = {
    files: safeClone(initialFiles),
    activeFile: Object.keys(initialFiles)[0] || "app.njk",
    templateName: Object.keys(initialFiles)[0] || "app.njk",
    contextJson: '{ "users": [{ "name": "sam" }, { "name": "kai" }] }',
  };

  const normalized = normalizeWorkspace(rawWorkspace, defaults);
  ensureDefaultGlobalFiles(normalized);
  files = normalized.files;
  activeFile = normalized.activeFile;
  templateName.value = normalized.templateName;
  contextJson.value = normalized.contextJson;
  setEditorValue(files[activeFile] ?? "");
  renderFileTabs();
  await persistWorkspace();
  await renderNow();
}

addFileBtn.addEventListener("click", async () => {
  const name = window.prompt("New file path (example: partials/card.njk):", "new-file.njk");
  if (!name) return;
  const trimmed = name.trim();
  if (!trimmed) return;
  if (files[trimmed] != null) {
    window.alert("File already exists");
    return;
  }

  files[activeFile] = getEditorValue();
  files[trimmed] = "";
  activeFile = trimmed;
  setEditorValue("");
  templateName.value = activeFile;
  renderFileTabs();

  try {
    await persistWorkspace();
    await renderNow();
  } catch (err) {
    window.alert(String(err));
  }
});

saveFileBtn.addEventListener("click", async () => {
  try {
    await persistWorkspace();
    await renderNow();
  } catch (err) {
    window.alert(String(err));
  }
});

deleteFileBtn.addEventListener("click", async () => {
  if (!activeFile) return;
  if (!window.confirm(`Delete "${activeFile}"?`)) return;

  try {
    delete files[activeFile];
    const names = Object.keys(files).sort();
    if (names.length === 0) {
      files = { "app.njk": "" };
      activeFile = "app.njk";
    } else {
      activeFile = names[0];
    }
    setEditorValue(files[activeFile] ?? "");
    templateName.value = activeFile;
    renderFileTabs();
    await persistWorkspace();
    await renderNow();
  } catch (err) {
    window.alert(String(err));
  }
});

resetWorkspaceBtn.addEventListener("click", async () => {
  if (!window.confirm("Reset workspace to server defaults?")) return;
  try {
    initialFiles = await fetchInitialFiles();
    await clearWorkspaceDB();
    await applyWorkspace(null);
  } catch (err) {
    window.alert(String(err));
  }
});

importWorkspaceBtn.addEventListener("click", () => {
  importWorkspaceInput.click();
});

importWorkspaceInput.addEventListener("change", async () => {
  const file = importWorkspaceInput.files?.[0];
  importWorkspaceInput.value = "";
  if (!file) return;

  try {
    const text = await file.text();
    const parsed = JSON.parse(text);
    await applyWorkspace(parsed);
  } catch (err) {
    window.alert(`Import failed: ${String(err)}`);
  }
});

exportWorkspaceBtn.addEventListener("click", async () => {
  try {
    const snapshot = getWorkspaceSnapshot();
    await persistWorkspace();
    downloadTextFile("nunchucks-workspace.json", JSON.stringify(snapshot, null, 2), "application/json");
  } catch (err) {
    window.alert(String(err));
  }
});

downloadOutputBtn.addEventListener("click", async () => {
  try {
    if (!lastRenderedHtml) {
      await renderNow();
    }
    const template = templateName.value.trim() || activeFile || "app";
    const base = template.replace(/\.[^.]+$/, "") || "output";
    downloadTextFile(`${base}.html`, lastRenderedHtml || "", "text/html");
  } catch (err) {
    window.alert(String(err));
  }
});

renderBtn.addEventListener("click", async () => {
  try {
    await persistWorkspace();
    await renderNow();
  } catch (err) {
    window.alert(String(err));
  }
});

viewButtons.forEach((btn) => btn.addEventListener("click", () => setView(btn.dataset.view)));

async function init() {
  try {
    initEditor();
    initialFiles = await fetchInitialFiles();

    const saved = await loadWorkspaceFromDB();
    await applyWorkspace(saved);

    setView("preview");
    editor.on("change", () => {
      files[activeFile] = getEditorValue();
    });

    const saveShortcut = async (evt) => {
      const isSaveKey = (evt.metaKey || evt.ctrlKey) && String(evt.key).toLowerCase() === "s";
      if (!isSaveKey) return;
      evt.preventDefault();
      try {
        await persistWorkspace();
        await renderNow();
      } catch (err) {
        window.alert(String(err));
      }
    };

    window.addEventListener("keydown", saveShortcut);
    editor.commands.addCommand({
      name: "nunchucksSaveActiveFile",
      bindKey: { win: "Ctrl-S", mac: "Command-S" },
      exec: async () => {
        try {
          await persistWorkspace();
          await renderNow();
        } catch (err) {
          window.alert(String(err));
        }
      },
    });
  } catch (err) {
    window.alert(String(err));
  }
}

init();
