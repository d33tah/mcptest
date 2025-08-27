# mcptest.xyz — Minimal MCP Flow Tester

Hosted, unauthorized MCP test server exposing a single tool, revealsecret. Use it to quickly verify your client can complete a real MCP round‑trip. Pure Go, no frameworks, easy to read and tweak the flow [1][2].

- Live site: https://mcptest.xyz
- Source: https://github.com/d33tah/mcptest

## What it does

- Exposes one tool, revealsecret, that returns a short confirmation so you know an MCP call really worked [2].  
- Built as a “pure” Go server using only the standard library (no frameworks) [1].  
- Supports both:
  - OpenAI Playground (Flows) via an SSE endpoint [2]
  - OpenAPI-based clients (e.g., Open WebUI) via an OpenAPI spec URL [2]
- Authorization: set to None or use any token — it’s ignored [2].

## Hosted endpoints

- MCP SSE (for OpenAI Playground Flows):  
  https://mcptest.xyz/sse [2]
- OpenAPI spec (for Open WebUI and other OpenAPI clients):  
  https://mcptest.xyz/openapi.json [2]
- Tool operation exposed by the spec:  
  POST https://mcptest.xyz/reveal_secret [2]

On success, revealsecret returns a small confirmation string; the current response body is:
{"result":"no siemanko"} [1].

## Quick start

Using OpenAI Playground (Flows)
- Add a tool that talks MCP over HTTP/S via SSE.
- Endpoint: https://mcptest.xyz/sse
- Authorization: None or any token (ignored).
- Invoke revealsecret and verify you get a short success string [2].

Using OpenAPI (Open WebUI)
- Import the OpenAPI spec URL: https://mcptest.xyz/openapi.json
- Run the operation that maps to revealsecret:
  - POST /reveal_secret
- Authorization: None or any token (ignored).
- You should receive a small “secret” confirming the call executed [2].

## API examples

cURL (hosted)
- POST https://mcptest.xyz/reveal_secret
- No auth required.

Example:
curl -X POST https://mcptest.xyz/reveal_secret

Expected response:
{"result":"no siemanko"} [1]

## Run locally

Prerequisites
- Go installed.

Build and run
- go run .  
- The server starts on http://localhost:8000 [1].

Local endpoints
- SSE (Playground Flows): http://localhost:8000/sse [1]
- OpenAPI spec: http://localhost:8000/openapi.json [1]
- Tool: POST http://localhost:8000/reveal_secret → {"result":"no siemanko"} [1]
- Basic health: http://localhost:8000/ping [1]

## OpenAI vs OpenAPI — don’t mix them up

- OpenAI: Company/products (e.g., Playground, Flows). Here, you configure a Flow tool that talks MCP over HTTP/S via SSE [2].
- OpenAPI: An API description format (formerly Swagger). Tools like Open WebUI can import an OpenAPI spec to call HTTP endpoints [2].

They differ by one letter, but they’re unrelated in this context. Use the SSE endpoint for OpenAI Playground Flows and the spec URL for OpenAPI-based clients [2].

## Feedback

Anything unexpected or have feedback? Please open a new issue:
https://github.com/d33tah/mcptest/issues/new [2]

---

References:  
[1] Server and handlers (pure Go, stdlib; port 8000; revealsecret returns "no siemanko")  
[2] Hosted endpoints, usage notes, and OpenAI vs OpenAPI guidance as presented on the site
