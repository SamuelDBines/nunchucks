import express from "express";
import fs from "node:fs";
import path from "node:path";
import os from "node:os";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const rootDir = path.resolve(__dirname, "..");
const viewsDir = path.join(__dirname, "views");
const publicDir = path.join(__dirname, "public");
const PORT = Number(process.env.APP_PORT || 5177);
const NUNCHUCKS_BIN = process.env.NUNCHUCKS_BIN || path.join(rootDir, "bin", "nunchucks");

if (!fs.existsSync(viewsDir)) fs.mkdirSync(viewsDir, { recursive: true });

const clients = new Set();

function broadcastRefresh(reason) {
  for (const res of clients) {
    res.write(`event: refresh\ndata: ${JSON.stringify({ reason, at: Date.now() })}\n\n`);
  }
}

function listFilesRecursive(dir, baseDir = dir, out = {}) {
  for (const ent of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, ent.name);
    if (ent.isDirectory()) {
      listFilesRecursive(full, baseDir, out);
      continue;
    }
    const rel = path.relative(baseDir, full).replaceAll(path.sep, "/");
    out[rel] = fs.readFileSync(full, "utf-8");
  }
  return out;
}

function resolveGlobalArgs(files) {
  const normalize = (p) => String(p || "").replace(/\\/g, "/").replace(/^\.?\//, "");
  const byName = new Map();
  for (const key of Object.keys(files || {})) {
    byName.set(normalize(key), key);
  }

  const args = [];
  if (byName.has("_globals/links.njk")) {
    args.push("-global-head", "_globals/links.njk");
  }
  if (byName.has("_globals/scripts.njk")) {
    args.push("-global-foot", "_globals/scripts.njk");
  }

  // Any other _globals/*.njk file is loaded as a global prelude so
  // macros/functions/constants become available across all templates.
  for (const normalized of byName.keys()) {
    if (!normalized.startsWith("_globals/") || !normalized.endsWith(".njk")) continue;
    if (normalized === "_globals/links.njk" || normalized === "_globals/scripts.njk") continue;
    args.push("-global", normalized);
  }

  return args;
}

const app = express();
app.use(express.json({ limit: "2mb" }));

function renderFromViews(template, context = {}) {
  const files = listFilesRecursive(viewsDir);
  const globalArgs = resolveGlobalArgs(files);

  if (fs.existsSync(NUNCHUCKS_BIN)) {
    return execFileSync(
      NUNCHUCKS_BIN,
      [
        "render",
        "-views", viewsDir,
        "-template", template,
        "-data", JSON.stringify(context),
        ...globalArgs,
      ],
      { encoding: "utf-8" }
    );
  }

  return execFileSync(
    "go",
    [
      "run", "./cmd/nunchucks", "render",
      "-views", viewsDir,
      "-template", template,
      "-data", JSON.stringify(context),
      ...globalArgs,
    ],
    {
      cwd: path.join(rootDir, "go"),
      encoding: "utf-8",
    }
  );
}

app.get("/bareframe", (_req, res) => {
  try {
    const html = renderFromViews("bareframe/index.njk");
    res.type("html").send(html);
  } catch (err) {
    res.status(500).send(String(err));
  }
});

app.get("/bareframe/:page", (req, res, next) => {
  try {
    const rawPage = String(req.params.page || "").trim();
    if (!rawPage || rawPage.includes("/") || rawPage.includes("..")) {
      return next();
    }
    const templateName = rawPage.endsWith(".njk")
      ? rawPage
      : rawPage.endsWith(".html")
        ? rawPage.replace(/\.html$/i, ".njk")
        : `${rawPage}.njk`;

    const fullTemplate = path.join(viewsDir, "bareframe", templateName);
    if (!fs.existsSync(fullTemplate)) {
      return next();
    }

    const html = renderFromViews(`bareframe/${templateName}`);
    res.type("html").send(html);
  } catch (err) {
    res.status(500).send(String(err));
  }
});

app.use(express.static(publicDir));

app.get("/__events", (_req, res) => {
  res.setHeader("Content-Type", "text/event-stream");
  res.setHeader("Cache-Control", "no-cache");
  res.setHeader("Connection", "keep-alive");
  res.flushHeaders?.();
  clients.add(res);
  res.write(`event: ping\ndata: ${Date.now()}\n\n`);
  const keepAlive = setInterval(() => {
    res.write(`event: ping\ndata: ${Date.now()}\n\n`);
  }, 20000);
  res.on("close", () => {
    clearInterval(keepAlive);
    clients.delete(res);
  });
});

app.get("/api/files", (_req, res) => {
  try {
    const files = listFilesRecursive(viewsDir);
    res.json({ ok: true, files });
  } catch (err) {
    res.status(500).json({ ok: false, error: String(err) });
  }
});

app.post("/api/files", (req, res) => {
  const rel = String(req.body?.path || "").trim();
  const content = String(req.body?.content ?? "");
  if (!rel) return res.status(400).json({ ok: false, error: "path is required" });
  if (rel.includes("..")) return res.status(400).json({ ok: false, error: "invalid path" });

  const full = path.join(viewsDir, rel);
  fs.mkdirSync(path.dirname(full), { recursive: true });
  fs.writeFileSync(full, content);
  broadcastRefresh(`save:${rel}`);
  res.json({ ok: true });
});

app.delete("/api/files", (req, res) => {
  const rel = String(req.query.path || "").trim();
  if (!rel) return res.status(400).json({ ok: false, error: "path is required" });
  if (rel.includes("..")) return res.status(400).json({ ok: false, error: "invalid path" });
  const full = path.join(viewsDir, rel);
  if (fs.existsSync(full)) fs.rmSync(full);
  broadcastRefresh(`delete:${rel}`);
  res.json({ ok: true });
});

app.post("/api/render", async (req, res) => {
  try {
    const template = String(req.body?.template || "app.njk");
    const context = req.body?.context && typeof req.body.context === "object" ? req.body.context : {};
    const files = req.body?.files && typeof req.body.files === "object"
      ? req.body.files
      : listFilesRecursive(viewsDir);
    const tmpViews = fs.mkdtempSync(path.join(os.tmpdir(), "nunchucks-app-"));
    try {
      for (const [name, content] of Object.entries(files)) {
        if (name.includes("..")) continue;
        const full = path.join(tmpViews, name);
        fs.mkdirSync(path.dirname(full), { recursive: true });
        fs.writeFileSync(full, String(content));
      }
      const globalArgs = resolveGlobalArgs(files);
      let output = "";
      if (fs.existsSync(NUNCHUCKS_BIN)) {
        output = execFileSync(
          NUNCHUCKS_BIN,
          [
            "render",
            "-views", tmpViews,
            "-template", template,
            "-data", JSON.stringify(context),
            ...globalArgs,
          ],
          {
            encoding: "utf-8",
          }
        );
      } else {
        output = execFileSync(
          "go",
          [
            "run", "./cmd/nunchucks", "render",
            "-views", tmpViews,
            "-template", template,
            "-data", JSON.stringify(context),
            ...globalArgs,
          ],
          {
            cwd: path.join(rootDir, "go"),
            encoding: "utf-8",
          }
        );
      }
      const outputs = {};
      for (const [name, content] of Object.entries(files)) outputs[name] = String(content);
      outputs[template] = output;
      res.json({ ok: true, outputs });
    } finally {
      fs.rmSync(tmpViews, { recursive: true, force: true });
    }
  } catch (err) {
    res.status(500).json({ ok: false, error: String(err) });
  }
});

app.use((_req, res) => {
  res.sendFile(path.join(publicDir, "index.html"));
});

fs.watch(viewsDir, { recursive: true }, (_eventType, filename) => {
  if (!filename) return;
  broadcastRefresh(`fs:${filename}`);
});

const server = app.listen(PORT, () => {
  console.log(`app dev server listening on http://localhost:${PORT}`);
});

for (const sig of ["SIGINT", "SIGTERM"]) {
  process.on(sig, () => {
    server.close(() => {
      process.exit(0);
    });
  });
}
