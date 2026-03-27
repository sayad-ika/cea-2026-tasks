# Craftsbite -- Technical Spec

- **Author:** Sayad Ibn Khairul Alam
- **Updated:** 2026-03-25
- **Status:** Draft

---

## 1. Overview

Craftsbite is a meal headcount planning system built for office teams. The system removes the manual overhead of tracking meal participation and work location through spreadsheets and chat messages. Discord is the primary user interface; Google Chat is the second supported platform. A web frontend may be introduced in future.

Employees interact through Discord or Google Chat slash commands to update their meal participation and work location for a given date. Team Leads can request a headcount summary scoped to their team. Admin and Logistics users have visibility across the entire organisation. The bot responds with the user's current status after every update.

---

## 2. Problem Statement

Meal participation and work location for office employees is currently tracked manually by Team Leads -- through spreadsheets and chat messages. This creates an unreliable process that depends entirely on Team Leads to collect, compile, and communicate headcounts to the logistics team each day.

Employees have no direct way to update their own participation. Logistics staff have no real-time visibility into headcount. The result is frequent miscommunication, over- or under-catering, and unnecessary interaction.

---

## 3. Goals and Non-Goals

**Goals**

- Employees can update meal participation and work location for a selected date via Discord or Google Chat.
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
Discord User                              Google Chat User
    |  slash command (HTTP POST)              |  slash command (HTTP POST)
    |  X-Signature-Ed25519 headers            |  Authorization: Bearer <Google-signed JWT>
    v                                         v
AWS API Gateway                           AWS API Gateway
    |  POST /interactions                     |  POST /gchat
    v                                         v
Discord Router Lambda                     GChat Router Lambda
(cmd/router)                              (cmd/gchat-router, internal/gchat/)
    |  verifies Ed25519 (DISCORD_PUBLIC_KEY)  |  verifies Bearer JWT (GCHAT_AUDIENCE)
    |  resolves caller: DISCORD#<id>          |  resolves caller: GCHAT#<email>
    |  checks ACL via discord.CheckPermission |  checks ACL via discord.CheckPermission
    |  async invoke (InvocationType=Event)    |  async invoke (InvocationType=Event)
    |  returns { "type": 5 }                  |  returns acknowledgement
    |                                         |
    +-------------------+---------------------+
                        |
                        v
             Command Lambda  (Go -- shared across both platforms)
                        |
                        |  self       -- /meal, /location, /status      (all roles)
                        |  `/status` fetches meal status + location + schedule: 3 parallel goroutines
                        |
                        |  management -- /override, /team-summary        (team_lead, admin, logistics)
                        |  `/team-summary` fan-out: one goroutine per member fetches meals + location
                        |
                        |  ops        -- /headcount, /schedule-day, /admin    (admin, logistics)
                        |  `/headcount` runs 3 parallel GSI1 queries, joins in memory
                        |
                        |  receives pre-verified event with caller identity + Source attached
                        |  Source=discord  →  Discord followup REST API
                        |  Source=gchat    →  Google Chat REST API (internal/gchat/reply.go)
                        v
             AWS DynamoDB  (on-demand -- single table)
                        +  craftsbite  (users, teams, meals, schedules, work locations, audit logs)
```

**Request flow (Discord):** Discord sends every slash command as an HTTP POST to the API Gateway URL, including two signature headers for request verification. API Gateway forwards the request to the Discord Router Lambda. This function first verifies the Ed25519 signature using `DISCORD_PUBLIC_KEY` — any request that fails verification is rejected immediately and no command Lambda is ever invoked. On success, it resolves the caller's identity by looking up their `discordId` in DynamoDB (`PK=DISCORD#<id>`), reads the command name from the request body, and asynchronously invokes the correct grouped command Lambda with the caller identity attached. It then immediately returns `{ "type": 5 }` to Discord within the 3-second deadline.

**Request flow (Google Chat):** Google Chat sends every slash command interaction as an HTTP POST to the `/gchat` API Gateway route, signed with a Google-issued Bearer JWT. The GChat Router Lambda verifies the JWT against Google's public keys using `GCHAT_AUDIENCE`, resolves the caller by email (`PK=GCHAT#<email>`), checks the ACL via the shared `discord.CheckPermission`, and asynchronously invokes the correct command Lambda. It then returns an immediate acknowledgement to Google Chat. The command Lambda receives a pre-verified, pre-routed event identical in structure to the Discord path (with `Source=gchat` attached), executes the business logic, and sends the result back via `internal/gchat/reply.go`.

**Boundaries:**

- API Gateway handles HTTPS termination — nothing else
- Discord Router Lambda owns Ed25519 verification, Discord identity resolution, and command dispatch — command Lambdas only ever receive verified, enriched events
- GChat Router Lambda owns JWT verification, Google Workspace identity resolution, and command dispatch — uses the same ACL and dispatch logic as the Discord router via `internal/discord/dispatch.go`
- Command Lambdas each own the business logic for their command group — no auth, no routing, no platform awareness beyond the `Source` field on the event
- DynamoDB owns persistence — all Lambdas are stateless and hold no data between invocations
- Discord and Google Chat are the two supported external callers in this iteration — a frontend application may be introduced later

---

## 5. Tech Stack & Rationale

|               | Choice                        | Why                                                                                                                                                                                                                                          |
| ------------- | ----------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Language      | Go                            | Compiles to a single static binary with minimal dependencies. Cold starts on Lambda are near-instant, which is critical for staying within Discord's hard 3-second response deadline.                                                        |
| Compute       | AWS Lambda                    | Serverless compute -- no servers to provision or maintain. Each function runs only when invoked and scales automatically. At current estimated usage (~6,000 requests/month) the cost sits within AWS's permanent free tier at $0.00/month.  |
| Handler model | Native Lambda handlers        | All Lambdas are written as native Go Lambda handlers using the AWS Lambda Go SDK. No HTTP framework or adapter is needed -- each function receives a structured event, processes it, and returns a structured response directly.             |
| Gateway       | AWS API Gateway (HTTP API)    | Exposes `POST /interactions` for Discord traffic and `POST /gchat` for Google Chat traffic. Handles HTTPS termination. Cost is negligible at current scale.                                                                                  |
| Database      | AWS DynamoDB (on-demand)      | Serverless NoSQL database -- no cluster to manage, no capacity to pre-provision. Scales with usage and costs nothing at idle. On-demand billing means we only pay for what we use.                                                           |
| Bot model     | Discord HTTP Interactions     | Slash commands are delivered as plain HTTP POST requests to our endpoint. This is the only Discord integration model compatible with serverless compute -- no persistent connection required.                                                |
| Bot model     | Google Chat HTTP Interactions | Slash commands are delivered as HTTP POST requests signed with a Google-issued Bearer JWT. Same serverless-compatible model as Discord — no persistent connection required. JWT verification uses Google's public keys via `GCHAT_AUDIENCE`. |

---

## 6. Requirements

**Functional**

- Employees can update meal participation and work location for a selected date via Discord or Google Chat slash commands.
- When opting out for a day without specifying a meal type, the service fans out and writes an opt-out record for every available meal on that date — no client-side enumeration required.
- The bot replies with the user's current status summary after each update.
- Team Leads can view a team-level participation summary for a selected date, and override participation for members of their own team.
- Admin can view org-wide summaries, manage day schedules and meal availability, and manage users and teams.
- Logistics can view org-wide headcount summaries and team summaries (read-only) — no write access.

**Role-based behavior**

| Role      | `/meal` `/location` `/status` | `/override`   | `/team-summary`       | `/headcount`  | `/schedule-day` | `/admin` |
| --------- | ----------------------------- | ------------- | --------------------- | ------------- | --------------- | -------- |
| Employee  | Own records only              | ✗             | ✗                     | ✗             | ✗               | ✗        |
| Team Lead | Own records only              | Own team only | Own team only         | ✗             | ✗               | ✗        |
| Logistics | Own records only              | ✗             | Read-only (all teams) | ✓ (read-only) | ✗               | ✗        |
| Admin     | Own records only              | Any user      | All teams             | ✓             | ✓               | ✓        |

**Validation rules**

- Updates for past dates are always rejected.
- Updates for a future date are rejected if the cutoff has passed (see Section 11).
- Overrides bypass the cutoff but not day availability — the meal must exist in `available_meals` for that date.
- When a day is marked `office_closed` or `govt_holiday`, available meals are forced to empty — no participation writes are accepted for that date.

**Participation resolution**

The effective status for any (user, date, meal) combination is determined in this order:

1. Meal not in `available_meals` for that date → **unavailable**
2. Explicit record exists in DynamoDB → **opted in** or **opted out**
3. No record, meal is available → **opted in** (system default)

When no day schedule exists for a date, the day is treated as normal and all meals resolve to opted in for users with no explicit record.

**Definition of Done**

- Discord and Google Chat bots respond to all slash commands within 3 seconds.
- Role-based access is enforced — no role can access data outside its scope.
- All participation and location writes are correctly validated against cutoff and date rules.

---

## 7. Key Decisions and Trade-offs

- **Single-table DynamoDB design** — all entities (users, teams, memberships, meal participations, work locations, day schedules, WFH periods, audit logs) live in one table (`craftsbite`). A single GSI (`GSI1`) is overloaded with clearly distinct `GSI1PK` prefixes to serve all secondary access patterns. This keeps billing, backups, and monitoring to one target, and is appropriate at the current scale of ~200 employees.
- **DynamoDB as primary data store** — chosen for its serverless model, zero idle cost, and natural fit with Lambda's stateless invocation pattern.
- **Fully serverless architecture** — Lambda, API Gateway, and DynamoDB together mean no persistent infrastructure to operate or scale manually. The entire system scales to zero when idle and scales up automatically under load.
- **Discord HTTP Interactions over Gateway (WebSocket) bot** — slash commands delivered as HTTP POST requests require no persistent connection, which is the only model compatible with Lambda. Signature verification via Ed25519 is handled by the Discord Router Lambda on every incoming request.
- **Google Chat HTTP Interactions** — slash commands are delivered as HTTP POST requests signed with a Google-issued Bearer JWT, consistent with the same serverless-compatible model as Discord. JWT verification is handled by the GChat Router Lambda on every incoming request.
- **Shared ACL and dispatch logic across platforms** — `discord.CheckPermission` and `discord.Dispatch` in `internal/discord/dispatch.go` are platform-agnostic. Both the Discord and GChat routers import them directly. The ACL table and Lambda dispatch map are defined once and enforced identically regardless of which platform the request originates from.
- **Router Lambda as sole entry point per platform** — each platform has its own router Lambda that owns verification, identity resolution, and dispatch. Keeping these separate avoids mixing auth schemes in a single function while still sharing all downstream logic.
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
| Google Chat lookup | `GCHAT#<email>`       | `LOOKUP`                               |
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
- `GCHAT#<email>` lookup row follows the same denormalized pattern — `userID` and `role` are resolved in a single `GetItem`. The `email` comes from the verified JWT claims.
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
| Resolve Google Chat user → internal user + role | `GetItem`        | `PK=GCHAT#<email>`, `SK=LOOKUP`                                                      |
| Get user profile by UUID                        | `GetItem`        | `PK=USER#<id>`, `SK=PROFILE`                                                         |
| List all active users                           | `Query` GSI1     | `GSI1PK=ENTITY#USER`, `GSI1SK begins_with "true#"`                                   |
| Get all team members                            | `Query`          | `PK=TEAM#<id>`, `SK begins_with "MEMBER#"`                                           |
| Find teams led by a user                        | `Query` GSI1     | `GSI1PK=TEAMLEAD#<leadUserID>`                                                       |
| Find which team a user belongs to               | `Query` GSI1     | `GSI1PK=USER_TEAMS#<userID>`                                                         |
| Get full day context (schedule + meals)         | `Query`          | `PK=DAY#<date>` — returns METADATA + MEALS in one round trip                         |
| Get specific meal participation                 | `GetItem`        | `PK=USER#<id>`, `SK=MEAL#<date>#<mealType>`                                          |
| Get all meal participation for a user on a date | `Query`          | `PK=USER#<id>`, `SK begins_with "MEAL#<date>#"`                                      |
| Get all participations for a date (headcount)   | `Query` GSI1     | `GSI1PK=<date>`, `GSI1SK begins_with "MEAL#"`                                        |
| Get WFH employees for a date                    | `Query` GSI1     | `GSI1PK=<date>`, filter `GSI1SK begins_with "wfh#"`                                  |
| Get Office employees for a date                 | `Query` GSI1     | `GSI1PK=<date>`, filter `GSI1SK begins_with "office#"`                               |
| Get all work locations for a date (headcount)   | `Query` GSI1     | `GSI1PK=<date>`, filter prefix `wfh#` OR `office#` — single query, filtered in DDB   |
| Get work location for (user, date)              | `GetItem`        | `PK=USER#<id>`, `SK=WORKLOCATION#<date>`                                             |
| Get monthly WFH count for a user                | `Query` + filter | `PK=USER#<id>`, `SK begins_with "WORKLOCATION#<YYYY-MM>"`, filter `location = "wfh"` |
| Check if date falls in a WFH period             | `Query`          | `PK=WFHPERIOD`, `SK <= "<date>#zzzz"` → check `end_date >= date` in app              |
| Write audit entry                               | `PutItem`        | `PK=AUDIT#<actorUserID>`, `SK=<timestamp>#<entityType>#<entityKey>`                  |
| Get all changes made to a user's records        | `Query` GSI1     | `GSI1PK=AUDITEE#<targetUserID>`                                                      |

---

## 10. Deployment

Each Lambda is compiled to a separate static binary named `bootstrap` (Lambda custom runtime requirement). The binaries are zipped and uploaded independently. No Docker image or layer is used. All functions use the `provided.al2` runtime.

- **Discord Router Lambda** — compiled from `cmd/router/main.go`, deployed as its own function, invoked directly by API Gateway (`POST /interactions`) on every Discord request
- **GChat Router Lambda** — compiled from `cmd/gchat-router/main.go`, deployed as its own function, invoked by API Gateway (`POST /gchat`) on every Google Chat interaction. Uses `internal/gchat/` (`event.go` — event types; `card.go` — Card v2 builder; `reply.go` — Chat REST API reply). Reuses `internal/discord/dispatch.go` for ACL checks and Lambda dispatch. Resolves callers by Google Workspace email (`PK=GCHAT#<email>`, `SK=LOOKUP`).
- **`self` Lambda** — compiled from `cmd/self/main.go`, handles `/meal`, `/location`, `/status` — available to all roles
- **`management` Lambda** — compiled from `cmd/management/main.go`, handles `/override`, `/team-summary` — available to `team_lead`, `admin`, and `logistics` (read-only)
- **`ops` Lambda** — compiled from `cmd/ops/main.go`, handles `/headcount`, `/schedule-day`, `/admin` — available to `admin` and `logistics` (headcount only)

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

## 13. Discord Router Lambda — Request Flow

For every `POST /interactions` call:

1. Verify Ed25519 signature — reject HTTP `401` on failure.
2. If `type=1` (PING), return `{ "type": 1 }` immediately.
3. Resolve caller identity: `GetItem PK=DISCORD#<discordId>`, `SK=LOOKUP` — return ephemeral error if not found.
4. Look up command name in dispatch table — return ephemeral error if unknown.
5. Invoke target Lambda asynchronously with enriched payload (`InvocationType=Event`).
6. Return `{ "type": 5 }` to Discord within the 3-second deadline.

Enriched payload: `userID`, `role`, `discordId`, `commandName`, `options`, `interactionToken`, `applicationId`, `Source=discord`.

---

## 14. GChat Router Lambda — Request Flow

For every `POST /gchat` call:

1. Verify Bearer JWT against Google's public keys using `GCHAT_AUDIENCE` — reject HTTP `401` on failure.
2. Parse the `ChatEvent` JSON body — reject HTTP `400` on failure.
3. Route on event type:
    - `ADDED_TO_SPACE` — return welcome text immediately; no command Lambda invoked.
    - `REMOVED_FROM_SPACE` / `CARD_CLICKED` — return empty acknowledgement; no command Lambda invoked.
    - `MESSAGE` — proceed to steps 4–7 below.
4. Resolve caller identity: `GetItem PK=GCHAT#<email>`, `SK=LOOKUP` — return error card if not found.
5. Check ACL via `discord.CheckPermission(commandName, role)` — return permission-denied card if access is denied.
6. Invoke target Lambda asynchronously with enriched payload (`InvocationType=Event`).
7. Return immediate acknowledgement to Google Chat.

Enriched payload: `userID`, `role`, `email`, `commandName`, `argumentText`, `replyName`, `Source=gchat`.

---

## 15. Router Dispatch Table

| Command        | Target Lambda | Environment Variable              |
| -------------- | ------------- | --------------------------------- |
| `meal`         | `self`        | `LAMBDA_SELF_FUNCTION_NAME`       |
| `location`     | `self`        | `LAMBDA_SELF_FUNCTION_NAME`       |
| `status`       | `self`        | `LAMBDA_SELF_FUNCTION_NAME`       |
| `override`     | `management`  | `LAMBDA_MANAGEMENT_FUNCTION_NAME` |
| `team-summary` | `management`  | `LAMBDA_MANAGEMENT_FUNCTION_NAME` |
| `headcount`    | `ops`         | `LAMBDA_OPS_FUNCTION_NAME`        |
| `schedule-day` | `ops`         | `LAMBDA_OPS_FUNCTION_NAME`        |
| `admin`        | `ops`         | `LAMBDA_OPS_FUNCTION_NAME`        |

---

## 16. Registered Slash Commands

### Discord

| Command         | Options                                                                                                                                                                     | Notes                                                                                                                                                                                                                                                                                                     |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/meal`         | `date` (req, supports range: `YYYY-MM-DD..YYYY-MM-DD`, `week`, `today..+N`), `status` (req: `in\|out`), `meal` (opt: `lunch\|snacks\|iftar\|event_dinner\|optional_dinner`) | If `meal` is omitted, the service reads available meals for that date and writes the status for every available meal (day-wide opt-in/out fan-out). Range operations apply the same status across all dates in the range; per-date successes and failures are tracked separately. Maximum range: 14 days. |
| `/location`     | `date` (req, supports range: `YYYY-MM-DD..YYYY-MM-DD`, `week`, `today..+N`), `location` (req: `office\|wfh`)                                                                | Range operations apply the same location across all dates in the range. Maximum range: 14 days.                                                                                                                                                                                                           |
| `/status`       | `date` (opt, default: tomorrow)                                                                                                                                             | Returns current meal participation, work location, and day status for the caller on that date. Read-only — no database writes. Fetches in 3 parallel goroutines. Changed fields are highlighted with an arrow prefix (→).                                                                                 |
| `/override`     | `date` (req), `user` (req), `meal` (req: `lunch\|snacks\|iftar\|event_dinner\|optional_dinner`), `status` (req: `in\|out`), `reason` (opt)                                  | Team Leads: own team only. Admin: any user. Bypasses cutoff; meal must be available for the date.                                                                                                                                                                                                         |
| `/team-summary` | `date` (req), `team_id` (opt)                                                                                                                                               | Team Leads: own team only (ignores `team_id`). Logistics: read-only, any team. Admin: any team.                                                                                                                                                                                                           |
| `/headcount`    | `date` (req)                                                                                                                                                                | Admin and Logistics only. Returns meal totals and Office vs WFH split for the date. Users with no work location record are counted as office.                                                                                                                                                             |
| `/schedule-day` | `date` (req), `status` (req: `normal\|office_closed\|govt_holiday\|celebration\|weekend\|event_day`), `meals` (opt: comma-separated meal types), `reason` (opt)             | Admin only. Setting `office_closed` or `govt_holiday` forces `meals` to empty.                                                                                                                                                                                                                            |
| `/admin`        | `action` (req: `create-user\|update-role\|deactivate-user\|create-team\|add-member\|remove-member`), plus action-specific options                                           | Admin only.                                                                                                                                                                                                                                                                                               |

### Google Chat

Google Chat uses free-text argument strings. Arguments are positional and parsed by the GChat Router Lambda before invoking the command Lambda. The six commands below are registered — `/override` and `/admin` are not exposed until their handlers are ready for the platform.

| Command         | Command ID | Argument format                    | Notes                                                                                                                                                                 |
| --------------- | ---------- | ---------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/meal`         | 1          | `<in\|out> <meal_type> [date]`     | `meal_type` required (no day-wide fan-out via omission on GChat). `date` defaults to today if omitted. `iftar` is not exposed — parity with what the handler accepts. |
| `/location`     | 2          | `<office\|wfh> [date]`             | `date` defaults to today if omitted.                                                                                                                                  |
| `/team-summary` | 3          | `[date]`                           | `date` defaults to today if omitted. `team_id` is not advertised — handler always uses the caller's first led team.                                                   |
| `/headcount`    | 4          | `<date>`                           | `date` is required. Router returns a usage hint card before invoking Lambda if omitted.                                                                               |
| `/status`       | 5          | `[date]`                           | `date` defaults to tomorrow if omitted. Read-only — no database writes.                                                                                               |
| `/schedule-day` | 6          | `<date> <status> [meals] [reason]` | Admin only. `date` in YYYY-MM-DD format. `meals` comma-separated. `reason` is free text (all remaining words).                                                        |

---

## 17. Command Usage

### `/meal`

```
/meal date:<YYYY-MM-DD|range> status:<in|out> [meal:<lunch|snacks|event_dinner|optional_dinner|all>]
```

Opts in or out of meals for a given date or date range. If `meal` is omitted or set to `all`, the status is applied to every available meal for that date. Date ranges use the format `YYYY-MM-DD..YYYY-MM-DD`; the `week` keyword targets the next 5 business days; shortcuts like `today..+4` are also supported. Maximum range: 14 days. For range operations, per-date successes and failures are reported separately.

| Scenario            | Reply                                                                                                                    |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| Success             | Updated status for all meals on that date (`✓` in, `✗` out, `—` unavailable). Changed meals highlighted with `→` prefix. |
| Range success       | Per-date summary with individual success/failure breakdown                                                               |
| Past date           | "Cannot update participation for a past date."                                                                           |
| Cutoff passed       | "Updates for \<date\> are closed. Cutoff was \<date−1\> at 9:00 PM."                                                     |
| Office closed       | "Office is closed on \<date\> — no meals are available."                                                                 |
| No meals configured | "No meals are configured for \<date\>."                                                                                  |
| Meal not available  | "That meal is not available on \<date\>."                                                                                |
| Range exceeds limit | "Date range cannot exceed 14 days."                                                                                      |

### `/location`

```
/location date:<YYYY-MM-DD|range> location:<office|wfh>
```

Sets work location for a given date or date range. On success, replies with the updated location and all meal statuses for that date. For single-date updates, the location change is highlighted with a `→` prefix; meal statuses are shown without prefix. Date range format follows the same rules as `/meal`.

| Scenario            | Reply                                                                |
| ------------------- | -------------------------------------------------------------------- |
| Success (office)    | `→ 🏢 Office` + meal statuses                                        |
| Success (WFH)       | `→ 🏠 WFH` + meal statuses                                           |
| Range success       | Per-date summary with individual success/failure breakdown           |
| Past date           | "Cannot set work location for a past date."                          |
| Cutoff passed       | "Updates for \<date\> are closed. Cutoff was \<date−1\> at 9:00 PM." |
| Range exceeds limit | "Date range cannot exceed 14 days."                                  |

### `/status`

```
/status [date:<YYYY-MM-DD>]
```

Returns the caller's current meal participation, work location, and day status for the given date. Read-only — no database writes occur. If `date` is omitted, defaults to tomorrow. Changed fields (relative to defaults) are highlighted with a `→` prefix.

| Scenario   | Reply                                                                                                          |
| ---------- | -------------------------------------------------------------------------------------------------------------- |
| Success    | Meal statuses (`✓` in, `✗` out, `—` unavailable) + location + day status. Changed fields highlighted with `→`. |
| Wrong role | Not applicable — available to all roles.                                                                       |

**Example reply:**

```
Status for 2026-03-26
📅 Normal Day
→ 🏠 WFH
Lunch      ✓
→ Snacks   ✗
Event Dinner  —
```

### `/headcount`

```
/headcount date:<YYYY-MM-DD>
```

Returns org-wide meal totals and Office vs WFH split for a date. Available to `admin` and `logistics` roles only.

**Implementation:** 3 parallel DynamoDB queries joined in memory:

1. `GSI1PK=<date>`, `GSI1SK begins_with "MEAL#"` — all meal participation records
2. `GSI1PK=<date>`, filter `GSI1SK begins_with "wfh#" OR begins_with "office#"` — all location records
3. `GSI1PK=ENTITY#USER`, `GSI1SK begins_with "true#"` — active user headcount
4. `GetItem PK=DAY#<date> SK=METADATA` — day schedule (for day status note)

Users with no location record are counted as `office`.

| Scenario   | Reply                                             |
| ---------- | ------------------------------------------------- |
| Success    | Meal totals + location split (see example below)  |
| Wrong role | "You do not have permission to use `/headcount`." |

**Example reply:**

```
Headcount for 2026-03-10 (8 employees)
🍽 Lunch:           7 opted in  │  1 opted out
🍴 Snacks:          4 opted in  │  4 opted out
📍 Office: 6  │  WFH: 2
```

---

### `/team-summary`

```
/team-summary date:<YYYY-MM-DD> [team_id:<uuid>]
```

Returns per-member participation and location for a team on a date. Available to `team_lead`, `admin`, and `logistics`.

**Role scoping:**

- Team Lead: auto-scoped to their own team (`team_id` option is ignored).
- Admin / Logistics: `team_id` option is required.

**Implementation:** per-member goroutine fan-out:

1. `Query PK=TEAM#<id>`, `SK begins_with "MEMBER#"` — all member edges.
2. For each member, one goroutine runs 3 calls in parallel:
    - `GetItem PK=USER#<id>`, `SK=PROFILE` — member name
    - `Query PK=USER#<id>`, `SK begins_with "MEAL#<date>#"` — meal statuses
    - `GetItem PK=USER#<id>`, `SK=WORKLOCATION#<date>` — location (absent = office)
3. Results sorted by member name.
4. Reply split into multiple messages if > 2000 chars.

| Scenario                        | Reply                                                |
| ------------------------------- | ---------------------------------------------------- |
| Success                         | Per-member table + totals footer                     |
| Wrong role                      | "You do not have permission to use `/team-summary`." |
| Team Lead with no team          | "You are not assigned to a team."                    |
| Admin/Logistics missing team_id | "Please provide a team_id."                          |

**Example reply:**

```
Team SAGA — 2026-03-10 (3 members)
Alice Johnson  │ Lunch ✓  │ Snacks ✗  │ Office
Bob Smith      │ Lunch ✓  │ Snacks ✓  │ WFH
Carol Lee      │ Lunch ✗  │ Snacks ✗  │ Office
──────────────────────────────────────
Totals: Lunch 2/3  │  Snacks 1/3  │  WFH 1/3
```

---

### `/schedule-day`

```
/schedule-day date:<YYYY-MM-DD|range> status:<normal|office_closed|govt_holiday|celebration|weekend|event_day> [meals:<comma-separated>] [reason:<text>]
```

Sets the day schedule and available meals for one or more dates. Available to `admin` role only.

#### Single-date usage

Passes a single `YYYY-MM-DD` date. The schedule is created or updated for that date immediately.

| Scenario                             | Reply                                                                  |
| ------------------------------------ | ---------------------------------------------------------------------- |
| Success                              | `✓ Day schedule set for <date>` with status, meals, and reason summary |
| Invalid date format                  | `Invalid date format. Use YYYY-MM-DD (e.g., 2026-03-25)`               |
| Invalid status value                 | `invalid day status: <value> (valid: [...])`                           |
| office_closed / govt_holiday + meals | `office_closed and govt_holiday days cannot have meals.`               |
| Wrong role                           | `You do not have permission to use /schedule-day.`                     |

**Example reply:**

```
✓ Day schedule set for 2026-03-27

**Status**: 📅 Normal Day
**Meals**: Lunch, Snacks, Iftar
```

#### Bulk scheduling

Passes a date range (`YYYY-MM-DD..YYYY-MM-DD`) or the `week` keyword (next 5 business days from tomorrow).

**Weekend skipping:** Saturdays and Sundays within the range are silently skipped — no record is created, no error is raised for those specific days. The reply note states that weekend dates were skipped.

**Atomicity:** All weekday writes in the batch succeed together or none do. If any single-date write fails (e.g. invalid meal type, DynamoDB error), every record already written in that batch is deleted before the error is returned to the user.

**Max range:** 14 days (shared with `/meal` and `/location`).

| Scenario                   | Reply                                                                                   |
| -------------------------- | --------------------------------------------------------------------------------------- |
| Success                    | `✓ Day schedule set to <status> for N weekday(s):` + bullet list + weekend-skipped note |
| All dates are weekends     | `All dates in the specified range fall on weekends. No schedule was set.`               |
| Start date after end date  | `Invalid date: end date X is before start date Y`                                       |
| Range exceeds 14 days      | `Invalid date: date range too large: N days (max 14 days)`                              |
| Any write fails (rollback) | `Bulk schedule failed (no changes saved): failed on <date>: <reason>`                   |

**Example reply (bulk success):**

```
✓ Day schedule set to 📅 Normal Day for 3 weekday(s):
  • 2026-04-07
  • 2026-04-08
  • 2026-04-09
_(Weekend dates in the range were automatically skipped.)_
```

---
