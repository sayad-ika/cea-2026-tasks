# Craftsbite -- Technical Spec

- **Author:** Sayad Ibn Khairul Alam
- **Updated:** 2026-03-04
- **Status:** Draft

---

## 1. Overview

Craftsbite is a meal headcount planning system built for office teams. The system removes the manual overhead of tracking meal participation and work location through spreadsheets and chat messages. Discord is the primary user interface -- web frontend may be introduced in future.

Employees interact primarily through Discord slash commands to update their meal participation and work location for a given date. Team Leads can request a headcount summary scoped to their team. Admin and Logistics users have visibility across the entire organisation. The Discord bot responds with the user's current status after every update.

---

## 2. Problem Statement

Meal participation and work location for office employees is currently tracked manually by Team Leads -- through spreadsheets and chat messages. This creates an unreliable process that depends entirely on Team Leads to collect, compile, and communicate headcounts to the logistics team each day.

Employees have no direct way to update their own participation. Logistics staff have no real-time visibility into headcount. The result is frequent miscommunication, over- or under-catering, and unnecessary interaction.

---

## 3. Goals and Non-Goals

**Goals**

- Employees can update meal participation and work location for a selected date via Discord.
- The bot replies with a status summary after each successful update.
- Team Leads can view a team-level participation summary for a selected date.
- Admin/Logistics can view an org-wide headcount summary for a selected date.
- Business logic is structured to support a web dashboard and additional integrations in future iterations without rework.

**Non-Goals**

- No web dashboard or frontend in this iteration.
- No user registration, password reset, or profile management.
- No scheduled or automated report generation -- on-demand only.
- No audit log or change history tracking.
- No bulk actions or company-wide overrides.

---

## 4. Cloud Architecture

```
Discord User
    |  slash command (HTTP POST)
    |  X-Signature-Ed25519 + X-Signature-Timestamp headers
    v
AWS API Gateway  (HTTP API)
    |  single route: POST /interactions
    v
Authorizer + Router Lambda  (Go -- native Lambda handler)
    |  verifies Ed25519 signature using DISCORD_PUBLIC_KEY
    |  rejects invalid requests immediately -- command Lambdas never invoked
    |  resolves caller identity via discordId -> DynamoDB
    |  reads command name from request body
    |  invokes the correct command Lambda asynchronously
    |  returns ack to Discord within 3 seconds
    v
Per-command Lambda  (Go -- native Lambda handler, one per slash command)
    |  receives pre-verified, pre-routed event with caller identity attached
    |  executes business logic for its command
    |  sends result back to Discord via followup REST API call
    v
AWS DynamoDB  (on-demand -- 3 tables)
    |  craftsbite-users (Users, Teams and who belong to which team)
    |  craftsbite-meals (Meal Participations, and day schedules)
    +  craftsbite-work (where employee are working from)
```

**Request flow:** Discord sends every slash command as an HTTP POST to the API Gateway URL, including two signature headers for request verification. API Gateway forwards the request to the Authorizer + Router Lambda. This function first verifies the Ed25519 signature using `DISCORD_PUBLIC_KEY` -- any request that fails verification is rejected immediately and no command Lambda is ever invoked. On success, it resolves the caller's identity by looking up their `discordId` in DynamoDB, reads the command name from the request body, and asynchronously invokes the correct per-command Lambda with the caller identity attached. It then immediately returns an acknowledgement to Discord within the 3-second deadline. The per-command Lambda receives a pre-verified, pre-routed event, executes the business logic, and sends the result back to Discord via a followup REST API call.

**Boundaries:**

- API Gateway handles HTTPS termination -- nothing else
- Authorizer + Router Lambda owns signature verification, identity resolution, and command dispatch -- command Lambdas only ever receive verified, enriched events
- Per-command Lambdas each own the business logic for a single command -- no auth, no routing
- DynamoDB owns persistence -- all Lambdas are stateless and hold no data between invocations
- Discord is the primary external caller in this iteration -- may introduce frontend application later on

---

## 5. Tech Stack & Rationale

| | Choice | Why |
| --------- | -------------------------- | --- |
| Language | Go | Compiles to a single static binary with minimal dependencies. Cold starts on Lambda are near-instant, which is critical for staying within Discord's hard 3-second response deadline. |
| Compute | AWS Lambda | Serverless compute -- no servers to provision or maintain. Each function runs only when invoked and scales automatically. At current estimated usage (~6,000 requests/month) the cost sits within AWS's permanent free tier at $0.00/month. |
| Handler model | Native Lambda handlers | All Lambdas are written as native Go Lambda handlers using the AWS Lambda Go SDK. No HTTP framework or adapter is needed -- each function receives a structured event, processes it, and returns a structured response directly. |
| Gateway | AWS API Gateway (HTTP API) | Exposes a single `POST /interactions` route that forwards all Discord traffic to the Authorizer + Router Lambda. Handles HTTPS termination. Cost is negligible at current scale. |
| Database | AWS DynamoDB (on-demand) | Serverless NoSQL database -- no cluster to manage, no capacity to pre-provision. Scales with usage and costs nothing at idle. On-demand billing means we only pay for what we use. |
| Bot model | Discord HTTP Interactions | Slash commands are delivered as plain HTTP POST requests to our endpoint. This is the only Discord integration model compatible with serverless compute -- no persistent connection required. |

---

## 6. Requirements

**Functional**

- Employees can update meal participation and work location for a selected date via Discord slash commands.
- The bot replies with the user's current status summary after each update.
- Team Leads can request a team-level participation summary for a selected date.
- Admin/Logistics can request an org-wide headcount summary for a selected date.

**Role-based behavior**

| Role | Can Do |
| --------------- | --------------------------------------------------------------------------- |
| Employee | Read and update own meal participation and work location only |
| Team Lead | View team-level summaries and individual statuses within their team |
| Admin/Logistics | View all summaries, org-wide stats, and trigger on-demand report generation |

**Validation rules**

- Updates are blocked after the daily cutoff time.
- Updates are blocked for past dates.

**Definition of Done**

- Discord bot responds accurately to all slash commands within 3 seconds.
- Role-based access is enforced -- no role can access data outside its scope.
- All participation and location writes are correctly validated against cutoff and date rules.

---

## 7. Key Decisions and Trade-offs

- **DynamoDB as primary data store** -- stores all meal participation, work location, headcount, and user records. Chosen for its serverless model, zero idle cost, and natural fit with Lambda's stateless invocation pattern.
- **Fully serverless architecture** -- Lambda, API Gateway, and DynamoDB together mean no persistent infrastructure to operate or scale manually. The entire system scales to zero when idle and scales up automatically under load.
- **Discord HTTP Interactions over Gateway (WebSocket) bot** -- slash commands delivered as HTTP POST requests require no persistent connection, which is the only model compatible with Lambda. Signature verification via Ed25519 is handled by the Authorizer + Router Lambda on every incoming request.
- **Combined Authorizer + Router Lambda** -- signature verification and command dispatch are handled in a single Lambda rather than two separate functions. Merging them eliminates an extra invocation hop, reduces cold start risk within Discord's 3-second deadline, and keeps the entry point cohesive. The function has two clear responsibilities: verify the request is legitimate, then hand it off to the correct command Lambda.
- **Per-command Lambdas** -- each slash command is handled by its own dedicated Lambda function. This gives each command an independent deployment unit, isolated failure domain, and a single clear responsibility. Because Discord delivers all slash commands to one URL, API Gateway cannot route by command name -- the Authorizer + Router Lambda reads the command from the request body and invokes the correct function asynchronously before returning an ack to Discord.
- **Native Lambda handlers over HTTP framework** -- all Lambdas are written as native Go Lambda handlers. There is no HTTP server, no Gin, and no adapter layer. Each function receives a structured event and returns a structured response directly. This eliminates unnecessary dependencies and keeps cold starts minimal.

---

## 8. Data Model

DynamoDB tables follow a sparse key design. Each table uses a generic `PK` / `SK` string key pair. Entity type is stored as an attribute on each item, not encoded in the table structure.

**craftsbite-users** -- stores users, teams, and memberships using an adjacency list pattern under shared `TEAM#` partitions. A separate `DISCORD#<id>` item acts as a lookup index for resolving Discord identity to an internal user record.

**craftsbite-meals** -- stores meal participation records per user per date, and day schedules under a shared `SCHEDULE` partition.

**craftsbite-work** -- stores work location records per user per date.

Each table uses a single GSI (`GSI1`) for secondary access patterns. All GSIs use `PAY_PER_REQUEST` billing with full item projection.

---

## 9. Key Access Patterns

| Pattern | Table | Method |
| ------------------------------------- | ---------------- | ------------------------------- |
| Resolve Discord user -> internal user | craftsbite-users | GetItem on `DISCORD#<id>` |
| List all active users | craftsbite-users | Query GSI1 |
| Get team members | craftsbite-users | Query `TEAM#<id>` partition |
| Get user's team | craftsbite-users | Query GSI1 |
| Get day schedule | craftsbite-meals | GetItem on `SCHEDULE` partition |
| Get all meal participation for a date | craftsbite-meals | Query GSI1 |
| Get all work locations for a date | craftsbite-work | Query GSI1 |

---

## 10. Deployment

Each Lambda is compiled to a separate static binary named `bootstrap` (Lambda custom runtime requirement). The binaries are zipped and uploaded independently. No Docker image or layer is used. All functions use the `provided.al2` runtime.

- **Authorizer + Router Lambda** -- compiled from `cmd/router/main.go`, deployed as its own function, invoked directly by API Gateway on every request
- **Per-command Lambdas** -- each compiled from their own `cmd/<command>/main.go`, deployed as independent functions, invoked asynchronously by the Router Lambda

Local development runs each binary directly as a standalone executable -- no adapter or environment detection required.

---

## 11. Cutoff Time Logic

Meal participation and work location updates for a given date are only accepted before the cutoff time of the **previous day at 09:00 PM**. For example, to update participation for Tuesday, the cutoff is Monday at 09:00 PM.

Updates submitted after the cutoff are rejected. Updates for past dates are always rejected regardless of cutoff. The cutoff time is stored in config and applied at the handler level before any write is attempted.

---

## 12. Error Handling

- **Signature verification failure** -- Authorizer + Router Lambda rejects the request immediately; no command Lambda is invoked
- **Unknown command** -- Router Lambda returns a user-facing Discord message indicating the command is not recognised; no command Lambda is invoked
- **Cutoff or past date violation** -- command Lambda rejects the write and returns a user-facing Discord followup message explaining why the update was blocked
- **DynamoDB error** -- command Lambda returns a generic failure message to Discord via followup; the request is not retried
- **Missing or misconfigured env vars** -- the affected Lambda fails to start; no request is served