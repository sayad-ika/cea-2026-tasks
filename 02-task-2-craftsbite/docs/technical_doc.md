# Craftsbite -- Technical Spec

- **Author:** Sayad Ibn Khairul Alam
- **Updated:** 2026-03-08
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
- No scheduled or automated report generation — on-demand only.

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
Router Lambda  (Go -- native Lambda handler)
    |  verifies Ed25519 signature using DISCORD_PUBLIC_KEY
    |  rejects invalid requests immediately -- command Lambdas never invoked
    |  resolves caller identity via discordId -> DynamoDB (GetItem DISCORD#<id>)
    |  reads command name from request body
    |  invokes the correct command Lambda asynchronously (InvocationType=Event)
    |  returns { "type": 5 } to Discord within 3 seconds
    v
Command Lambda  (Go -- native Lambda handler, one per command group)
    |  self       -- /meal, /location, /status      (all roles)
    |  management -- /override, /team-summary        (team_lead, admin, logistics read-only)
    |  ops        -- /headcount, /set-day, /admin    (admin, logistics)
    |  receives pre-verified, pre-routed event with caller identity attached
    |  executes business logic for its command group
    |  sends result back to Discord via followup REST API call
    v
AWS DynamoDB  (on-demand -- single table)
    +  craftsbite  (all entities: users, teams, meals, schedules, work locations, audit logs)
```

**Request flow:** Discord sends every slash command as an HTTP POST to the API Gateway URL, including two signature headers for request verification. API Gateway forwards the request to the Router Lambda. This function first verifies the Ed25519 signature using `DISCORD_PUBLIC_KEY` — any request that fails verification is rejected immediately and no command Lambda is ever invoked. On success, it resolves the caller's identity by looking up their `discordId` in DynamoDB, reads the command name from the request body, and asynchronously invokes the correct grouped command Lambda with the caller identity attached. It then immediately returns `{ "type": 5 }` to Discord within the 3-second deadline. The command Lambda receives a pre-verified, pre-routed event, executes the business logic, and sends the result back to Discord via a followup REST API call.

**Boundaries:**

- API Gateway handles HTTPS termination — nothing else
- Router Lambda owns signature verification, identity resolution, and command dispatch — command Lambdas only ever receive verified, enriched events
- Command Lambdas each own the business logic for their command group — no auth, no routing
- DynamoDB owns persistence — all Lambdas are stateless and hold no data between invocations
- Discord is the primary external caller in this iteration — may introduce frontend application later on

---

## 5. Tech Stack & Rationale

|               | Choice                     | Why                                                                                                                                                                                                                                         |
| ------------- | -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Language      | Go                         | Compiles to a single static binary with minimal dependencies. Cold starts on Lambda are near-instant, which is critical for staying within Discord's hard 3-second response deadline.                                                       |
| Compute       | AWS Lambda                 | Serverless compute -- no servers to provision or maintain. Each function runs only when invoked and scales automatically. At current estimated usage (~6,000 requests/month) the cost sits within AWS's permanent free tier at $0.00/month. |
| Handler model | Native Lambda handlers     | All Lambdas are written as native Go Lambda handlers using the AWS Lambda Go SDK. No HTTP framework or adapter is needed -- each function receives a structured event, processes it, and returns a structured response directly.            |
| Gateway       | AWS API Gateway (HTTP API) | Exposes a single `POST /interactions` route that forwards all Discord traffic to the Authorizer + Router Lambda. Handles HTTPS termination. Cost is negligible at current scale.                                                            |
| Database      | AWS DynamoDB (on-demand)   | Serverless NoSQL database -- no cluster to manage, no capacity to pre-provision. Scales with usage and costs nothing at idle. On-demand billing means we only pay for what we use.                                                          |
| Bot model     | Discord HTTP Interactions  | Slash commands are delivered as plain HTTP POST requests to our endpoint. This is the only Discord integration model compatible with serverless compute -- no persistent connection required.                                               |

---

## 6. Requirements

**Functional**

- Employees can update meal participation and work location for a selected date via Discord slash commands.
- When opting out for a day without specifying a meal type, the service fans out and writes an opt-out record for every available meal on that date — no client-side enumeration required.
- The bot replies with the user's current status summary after each update.
- Team Leads can view a team-level participation summary for a selected date, and override participation for members of their own team.
- Admin can view org-wide summaries, manage day schedules and meal availability, and manage users and teams.
- Logistics can view org-wide headcount summaries and team summaries (read-only) — no write access.

**Role-based behavior**

| Role      | `/meal` `/location` `/status` | `/override`   | `/team-summary`       | `/headcount` | `/set-day` | `/admin` |
| --------- | ----------------------------- | ------------- | --------------------- | ------------ | ---------- | -------- |
| Employee  | Own records only              | ✗             | ✗                     | ✗            | ✗          | ✗        |
| Team Lead | Own records only              | Own team only | Own team only         | ✗            | ✗          | ✗        |
| Logistics | Own records only              | ✗             | Read-only (all teams) | ✓            | ✗          | ✗        |
| Admin     | Own records only              | Any user      | All teams             | ✓            | ✓          | ✓        |

**Validation rules**

- Updates for past dates are always rejected.
- Updates for a future date are rejected if the cutoff has passed (see Section 11).
- Overrides bypass the cutoff but not day availability — the meal must exist in `available_meals` for that date.
- When a day is marked `office_closed` or `govt_holiday`, available meals are forced to empty — no participation writes are accepted for that date.

**Definition of Done**

- Discord bot responds to all slash commands within 3 seconds.
- Role-based access is enforced — no role can access data outside its scope.
- All participation and location writes are correctly validated against cutoff and date rules.

---

## 7. Key Decisions and Trade-offs

- **Single-table DynamoDB design** — all entities (users, teams, memberships, meal participations, work locations, day schedules, WFH periods, audit logs) live in one table (`craftsbite`). A single GSI (`GSI1`) is overloaded with clearly distinct `GSI1PK` prefixes to serve all secondary access patterns. This keeps billing, backups, and monitoring to one target, and is appropriate at the current scale of ~200 employees.
- **DynamoDB as primary data store** — chosen for its serverless model, zero idle cost, and natural fit with Lambda's stateless invocation pattern.
- **Fully serverless architecture** — Lambda, API Gateway, and DynamoDB together mean no persistent infrastructure to operate or scale manually. The entire system scales to zero when idle and scales up automatically under load.
- **Discord HTTP Interactions over Gateway (WebSocket) bot** — slash commands delivered as HTTP POST requests require no persistent connection, which is the only model compatible with Lambda. Signature verification via Ed25519 is handled by the Router Lambda on every incoming request.
- **Router Lambda as sole entry point** — signature verification and command dispatch are handled in one Lambda. Merging them eliminates an extra invocation hop, reduces cold start risk within Discord's 3-second deadline, and keeps the entry point cohesive. The function has two clear responsibilities: verify the request is legitimate, then hand it off to the correct command Lambda.
- **Grouped command Lambdas (not per-command)** — commands are grouped by role scope into three Lambdas: `self` (employee self-service), `management` (team lead and admin oversight), and `ops` (admin and logistics operations). Each group gets an independent deployment unit and failure domain. Grouping by role scope is more natural than one Lambda per command and avoids unnecessary proliferation of functions for closely related operations.
- **Day-wide opt-out handled by service fan-out** — when a user opts out for a full day without specifying a meal type, the `self` Lambda reads the available meals for that date and issues one `PutItem` per meal via `BatchWriteItem`. The schema stays uniform — headcount queries always see individual per-meal records regardless of whether the opt-out was issued one meal at a time or for the whole day.
- **Native Lambda handlers over HTTP framework** — all Lambdas are written as native Go Lambda handlers. There is no HTTP server, no Gin, and no adapter layer. Each function receives a structured event and returns a structured response directly. This eliminates unnecessary dependencies and keeps cold starts minimal.
- **Audit log as append-only per-actor** — every mutation (participation, location, override, day schedule, admin write) appends a record to `AUDIT#<actorUserID>`. A reverse GSI entry (`AUDITEE#<targetUserID>`) enables admins to query all changes made to any user's records regardless of who made them.

---

## 8. Data Model

All entities live in a single DynamoDB table named `craftsbite` (`PAY_PER_REQUEST`). Every item carries a generic `PK` / `SK` string key pair and an `entityType` attribute as a string discriminator. A single GSI (`GSI1`) is overloaded across all entities using distinct `GSI1PK` prefixes.

**Entity → Key mapping**

| Entity             | PK                    | SK                                     |
| ------------------ | --------------------- | -------------------------------------- |
| User profile       | `USER#<id>`           | `PROFILE`                              |
| Email lookup       | `EMAIL#<email>`       | `LOOKUP`                               |
| Discord lookup     | `DISCORD#<discordId>` | `LOOKUP`                               |
| Team metadata      | `TEAM#<id>`           | `METADATA`                             |
| Team listing       | `TEAM#<id>`           | `LISTING`                              |
| Team member        | `TEAM#<id>`           | `MEMBER#<userID>`                      |
| Day schedule       | `DAY#<date>`          | `METADATA`                             |
| Available meals    | `DAY#<date>`          | `MEALS`                                |
| Meal participation | `USER#<id>`           | `MEAL#<date>#<mealType>`               |
| Work location      | `USER#<id>`           | `WORKLOCATION#<date>`                  |
| WFH period         | `WFHPERIOD`           | `<start_date>#<end_date>`              |
| Audit log          | `AUDIT#<actorUserID>` | `<timestamp>#<entityType>#<entityKey>` |

**GSI1 — overloaded index**

| GSI1PK                   | GSI1SK                              | Serves                                        |
| ------------------------ | ----------------------------------- | --------------------------------------------- |
| `<date>`                 | `MEAL#<userID>`                     | All participations for a date (headcount)     |
| `<date>`                 | `WFH#<userID>` or `OFFICE#<userID>` | Office vs WFH split for a date                |
| `USER_TEAMS#<userID>`    | `TEAM#<teamID>`                     | Reverse team lookup per user                  |
| `TEAMLEAD#<leadID>`      | `TEAM#<id>`                         | Teams led by a specific user                  |
| `ENTITY#TEAM`            | `<active>#TEAM#<id>`                | Active team listing (written on LISTING item) |
| `ENTITY#USER`            | `<active>#USER#<id>`                | Active user listing                           |
| `AUDITEE#<targetUserID>` | `<timestamp>`                       | All changes made to a user's records          |

**Key design notes:**

- `DISCORD#<discordId>` lookup row carries `role` denormalized — a single `GetItem` resolves both `userID` and `role` on every Lambda invocation with no second read.
- `DAY#<date>` is a shared partition for both `METADATA` and `MEALS`. A single `Query` returns the full day context in one round trip.
- Team metadata needs two distinct GSI1 patterns (by lead and by listing). Because a DynamoDB item can only carry one GSI1PK/GSI1SK pair, the team uses two items: `METADATA` carries `TEAMLEAD#`, and a separate `LISTING` item carries `ENTITY#TEAM`.
- User profile and Discord lookup are always kept in sync via `TransactWriteItems` — role changes update both atomically.
- When a user or team's `active` flag changes, `GSI1SK` must be updated in the same write operation (`true#` ↔ `false#` prefix) to keep the active listing queries correct.
- Audit log is write-only (`PutItem` exclusively). The reverse GSI entry `AUDITEE#<targetUserID>` is written on every audit item to support admin accountability queries.

---

## 9. Key Access Patterns

| Pattern                                         | Operation        | Key Expression                                                                       |
| ----------------------------------------------- | ---------------- | ------------------------------------------------------------------------------------ |
| Resolve Discord user → internal user + role     | `GetItem`        | `PK=DISCORD#<discordId>`, `SK=LOOKUP`                                                |
| Get user profile                                | `GetItem`        | `PK=USER#<id>`, `SK=PROFILE`                                                         |
| List all active users                           | `Query` GSI1     | `GSI1PK=ENTITY#USER`, `GSI1SK begins_with "true#"`                                   |
| Get all team members                            | `Query`          | `PK=TEAM#<id>`, `SK begins_with "MEMBER#"`                                           |
| Find which team a user belongs to               | `Query` GSI1     | `GSI1PK=USER_TEAMS#<userID>`                                                         |
| Get full day context (schedule + meals)         | `Query`          | `PK=DAY#<date>` — returns METADATA + MEALS in one round trip                         |
| Get specific meal participation                 | `GetItem`        | `PK=USER#<id>`, `SK=MEAL#<date>#<mealType>`                                          |
| Get all meal participation for a user on a date | `Query`          | `PK=USER#<id>`, `SK begins_with "MEAL#<date>#"`                                      |
| Get all participations for a date (headcount)   | `Query` GSI1     | `GSI1PK=<date>`, `GSI1SK begins_with "MEAL#"`                                        |
| Get WFH employees for a date                    | `Query` GSI1     | `GSI1PK=<date>`, `GSI1SK begins_with "WFH#"`                                         |
| Get Office employees for a date                 | `Query` GSI1     | `GSI1PK=<date>`, `GSI1SK begins_with "OFFICE#"`                                      |
| Get work location for (user, date)              | `GetItem`        | `PK=USER#<id>`, `SK=WORKLOCATION#<date>`                                             |
| Get monthly WFH count for a user                | `Query` + filter | `PK=USER#<id>`, `SK begins_with "WORKLOCATION#<YYYY-MM>"`, filter `location = "wfh"` |
| Check if date falls in a WFH period             | `Query`          | `PK=WFHPERIOD`, `SK <= "<date>#zzzz"` → check `end_date >= date` in app              |
| Write audit entry                               | `PutItem`        | `PK=AUDIT#<actorUserID>`, `SK=<timestamp>#<entityType>#<entityKey>`                  |
| Get all changes made to a user's records        | `Query` GSI1     | `GSI1PK=AUDITEE#<targetUserID>`                                                      |

---

## 10. Deployment

Each Lambda is compiled to a separate static binary named `bootstrap` (Lambda custom runtime requirement). The binaries are zipped and uploaded independently. No Docker image or layer is used. All functions use the `provided.al2` runtime.

- **Router Lambda** — compiled from `cmd/router/main.go`, deployed as its own function, invoked directly by API Gateway on every request
- **`self` Lambda** — compiled from `cmd/self/main.go`, handles `/meal`, `/location`, `/status` — available to all roles
- **`management` Lambda** — compiled from `cmd/management/main.go`, handles `/override`, `/team-summary` — available to `team_lead`, `admin`, and `logistics` (read-only)
- **`ops` Lambda** — compiled from `cmd/ops/main.go`, handles `/headcount`, `/set-day`, `/admin` — available to `admin` and `logistics` (headcount only)

Local development runs each binary directly as a standalone executable — no adapter or environment detection required.

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

---

---

## 13. Router Lambda Request Flow

For every `POST /interactions` call:

1. Verify Ed25519 signature — reject HTTP `401` on failure.
2. If `type=1` (PING), return `{ "type": 1 }` immediately.
3. Resolve caller identity: `GetItem PK=DISCORD#<discordId>`, `SK=LOOKUP` — return ephemeral error if not found.
4. Look up command name in dispatch table — return ephemeral error if unknown.
5. Invoke target Lambda asynchronously with enriched payload (`InvocationType=Event`).
6. Return `{ "type": 5 }` to Discord within the 3-second deadline.

Enriched payload: `userID`, `role`, `discordId`, `commandName`, `options`, `interactionToken`, `applicationId`.

---

## 14. Router Dispatch Table

| Command        | Target Lambda | Environment Variable              |
| -------------- | ------------- | --------------------------------- |
| `meal`         | `self`        | `LAMBDA_SELF_FUNCTION_NAME`       |
| `location`     | `self`        | `LAMBDA_SELF_FUNCTION_NAME`       |
| `status`       | `self`        | `LAMBDA_SELF_FUNCTION_NAME`       |
| `override`     | `management`  | `LAMBDA_MANAGEMENT_FUNCTION_NAME` |
| `team-summary` | `management`  | `LAMBDA_MANAGEMENT_FUNCTION_NAME` |
| `headcount`    | `ops`         | `LAMBDA_OPS_FUNCTION_NAME`        |
| `set-day`      | `ops`         | `LAMBDA_OPS_FUNCTION_NAME`        |
| `admin`        | `ops`         | `LAMBDA_OPS_FUNCTION_NAME`        |

---
