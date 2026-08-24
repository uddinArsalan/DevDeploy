# DevDeploy — Load Testing the Deployment Pipeline

DevDeploy is a self-hosted PaaS (think Vercel/Render/Railway, built from scratch) that takes a Git repo and turns it into a running, publicly reachable container — handling cloning, building, image creation, container orchestration, and live deployment status streaming.

This document covers the load test built for it: what it does, how to run it, and what the results say about the system.

---

## Why this test exists

Most load tests hit a single stateless endpoint and call it a day. That doesn't say much about a system like DevDeploy, where the interesting behavior — queueing, building, container lifecycle, streaming — happens *after* the initial HTTP response.

So this test drives the **real end-to-end pipeline**: it logs in, creates a project, kicks off a real deployment against a real Git repo, and stays connected via SSE until that deployment actually finishes — meaning a real container is built and running by the end of every successful iteration.

---

## Architecture under test

| Component | Role |
|---|---|
| Go backend | HTTP API, auth, orchestration |
| Docker SDK | Builds images, creates/manages deployment containers |
| RabbitMQ | Queues deployment jobs for workers |
| Redis Streams | Carries deployment status/log events |
| PostgreSQL | Source of truth for users, projects, deployments |
| Railpack | Detects and builds the application |
| Caddy | Reverse proxy inside deployment containers |
| SSE | Streams live deployment status/logs to the client |
| Cloudflare Tunnel | Exposes deployed apps publicly |

---

## The test script

Built with [k6](https://k6.io) + TypeScript, using the [`xk6-sse`](https://github.com/grafana/xk6-sse) extension to hold open a live event stream (k6 doesn't support SSE out of the box).

**Flow, per virtual user, per iteration:**

1. **Login** — `POST /auth/signin`
2. **Create project** — `POST /project`, pointing at a real test repo
3. **Trigger deployment** — `POST /projects/:id/deployments`
4. **Open SSE stream** to `/stream/:deployId` and block until the pipeline emits a `done` event — meaning clone, build, image build, container creation, and container start all actually happened

A `setup()` step pre-registers 10 test users once, up front, which the virtual users then log in as.

**Custom metrics tracked:**

- `deployment_duration` — time from "deployment created" to the SSE `done` event. This is the real pipeline time.
- `total_duration` — time from the start of the iteration (before login) to `done`. Adds login + project-creation overhead on top.

**Load shape:**

```
10 virtual users × 5 iterations each = 50 intended deployments
max duration: 10 minutes, 30s graceful stop
```

**Run it:**

```bash
xk6 build --with github.com/grafana/xk6-sse@latest
./k6 run load-test.ts
```

---

## Results

| Metric | Result |
|---|---|
| Checks passed | 171 / 181 (94.47%) |
| Iterations completed | 36 / 50 intended |
| Dropped iterations | 5 |
| Interrupted iterations | 9 |
| `http_req_failed` | 0% (0 / 145) |
| SSE events delivered | 6,452 (~10.2/sec) |

**Deployment duration** (create → running container):

| | |
|---|---|
| min | 32.1s |
| avg | 2m32s |
| median | 2m41s |
| p90 | 3m19s |
| p95 | 3m26s |
| max | 3m45s |

**API latency**, isolated from the long-lived SSE connection (`expected_response:true`):

| | |
|---|---|
| avg | 489.8ms |
| median | 503.0ms |
| p90 | 795.7ms |
| p95 | 812.7ms |

---

## What this shows

**The API layer is fast and reliable.** Login, project creation, and deployment creation respond in under a second on average (p95 ~813ms), with zero failed requests on the discrete calls. This is the number that matters for "is my API well-built," and it's solid.

**The pipeline works end-to-end, under concurrency, without crashing.** 36 real deployments completed successfully during the run — each one a genuine git clone → build → Docker image → running container, with live status streamed the whole way via Redis Streams → SSE.

**Aggregate `http_req_duration` (30.65s avg) is a red herring, not a defect.** It's averaging two very different things: sub-second API calls and multi-minute SSE connections that stay open for the life of a deployment. Once separated, the real story is "fast API, slow-but-working pipeline" — not "slow API."

**The 5 dropped + 9 interrupted iterations are a test-sizing artifact, not a backend failure.** With deployments taking up to 3m45s and only 10 VUs, some scheduled iterations couldn't start in time, and some that started ran past the test's time window. This is expected when a fixed-iteration test meets a multi-minute workload — the fix is a longer window or a duration-based executor, not a backend fix.

**Deployment duration itself (avg 2m32s) is not inherently a problem** — it's in line with what real platforms take for non-trivial builds. Whether it's *optimizable* is genuinely unknown from this test alone, because the 2m32s is currently a black box: clone time, build time, image build time, and container startup are all bundled into one number.

---

## What's next

The honest finding from this round: the system is fast where it should be fast (API) and functionally correct under load (pipeline). The next step isn't "make it faster" — it's **breaking the black box open**:

- Instrument the Go backend to emit per-phase timing (clone / build / image build / container start / app-ready) so the 2m32s can actually be attributed


This test's job was to prove the pipeline works and to establish a fast, trustworthy API baseline — both of which it did. The phase-level breakdown is what turns "here's a number" into "here's where the time goes and why."