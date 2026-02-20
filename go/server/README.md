# Playground Server (Stateless)

This server provides a stateless memory-render endpoint for the docs playground.

## Run

```bash
cd go
go run ./server
```

Server listens on `http://localhost:8090`.

## Endpoint

- `POST /api/playground/render`

Request body:

```json
{
  "template": "app.njk",
  "files": {
    "app.njk": "...",
    "layout.njk": "..."
  },
  "context": {
    "users": [{"name": "sam"}]
  }
}
```

Response body:

```json
{
  "ok": true,
  "outputs": {
    "app.njk": "...rendered...",
    "layout.njk": "...rendered..."
  }
}
```

## Playground wiring

Set this in browser console (or in your docs bootstrap) when running local docs:

```js
window.NUNCHUCKS_PLAYGROUND_API = "http://localhost:8090/api/playground/render";
```

If unreachable, playground falls back to client-side demo rendering.
