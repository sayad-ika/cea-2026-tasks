# Craftsbite — Technical Spec

- **Author:** Sayad Ibn Khairul Alam
- **Updated:** 28 February 2026
- **Status:** Draft

---

## 1. Overview

Craftsbite is a meal headcount planning system built for office teams. The system removes the manual overhead of tracking meal participation and work location through spreadsheets and chat messages. Discord is the primary user interface — web frontend may be introduced in future.

Employees interact primarily through Discord slash commands to update their meal participation and work location for a given date. Team Leads can request a headcount summary scoped to their team. Admin and Logistics users have visibility across the entire organisation. The Discord bot responds with the user's current status after every update.

---

## 2. Problem Statement

Meal participation and work location for office employees is currently tracked manually by Team Leads — through spreadsheets and chat messages. This creates an unreliable process that depends entirely on Team Leads to collect, compile, and communicate headcounts to the logistics team each day.

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
- No audit log or change history tracking.
- No bulk actions or company-wide overrides.

---

## 4. Cloud Architecture

```
Discord User
    │  slash command (HTTP POST)
    │  X-Signature-Ed25519 + X-Signature-Timestamp headers
    ▼
AWS API Gateway  (HTTP API)
    │  single route: ANY /{proxy+}
    │  proxies full request — no transformation
    ▼
AWS Lambda  (Go binary — provided.al2 runtime)
    │  verifies Ed25519 signature
    │  resolves caller identity via discordId → DynamoDB
    │  executes business logic
    │  responds within 3 seconds
    ▼
AWS DynamoDB  (on-demand — 3 tables)
    │  craftsbite-users (Users, Teams and who belong to which team)
    │  craftsbite-meals (Meal Participations, and day schedules)
    └  craftsbite-work (where employee are working from)
```

**Request flow:** Discord sends every slash command as an HTTP POST to the API Gateway URL, including two signature headers for request verification. API Gateway proxies the request to Lambda without modification. Lambda verifies the Ed25519 signature using `DISCORD_PUBLIC_KEY` — any request that fails verification is rejected with `401` before any business logic runs. On success, the handler resolves the caller's identity by looking up their `discordId` in DynamoDB, determines their role and team, executes the appropriate logic, and returns a response to Discord.

**Boundaries:**

- API Gateway handles HTTPS termination and request routing — nothing else
- Lambda owns all application logic — routing, validation, business rules, and data access
- DynamoDB owns persistence — Lambda is stateless and holds no data between invocations
- Discord is the primary external caller in this iteration — may introduce frontend application later on

---

## 5. Tech Stack & Rationale

|           | Choice                     | Why                                                                                                                                                                                                                                                |
| --------- | -------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Language  | Go                         | Compiles to a single static binary with minimal dependencies. Cold starts on Lambda are near-instant, which is critical for staying within Discord's hard 3-second response deadline.                                                              |
| Framework | Gin                        | Lightweight HTTP router that structures the service cleanly. Works identically in local development and on Lambda via the Gin adapter — no code changes between environments.                                                                      |
| Compute   | AWS Lambda                 | Serverless compute — no servers to provision or maintain. The function runs only when a request arrives and scales automatically. At current estimate usage (~6,000 requests/month) the cost sits within AWS's permanent free tier at $0.00/month. |
| Gateway   | AWS API Gateway (HTTP API) | Exposes a single `ANY /{proxy+}` route that forwards all traffic to Lambda. Handles HTTPS termination and request routing without additional configuration. Cost is negligible at current scale.                                                   |
| Database  | AWS DynamoDB (on-demand)   | Serverless NoSQL database — no cluster to manage, no capacity to pre-provision. Scales with usage and costs nothing at idle. On-demand billing means we only pay for what we use.                                                                  |
| Bot model | Discord HTTP Interactions  | Slash commands are delivered as plain HTTP POST requests to our endpoint. This is the only Discord integration model compatible with serverless compute — no persistent connection required.                                                       |

---

## 6. Requirements

**Functional**

- Employees can update meal participation and work location for a selected date via Discord slash commands.
- The bot replies with the user's current status summary after each update.
- Team Leads can request a team-level participation summary for a selected date.
- Admin/Logistics can request an org-wide headcount summary for a selected date.

**Role-based behavior**

| Role            | Can Do                                                                      |
| --------------- | --------------------------------------------------------------------------- |
| Employee        | Read and update own meal participation and work location only               |
| Team Lead       | View team-level summaries and individual statuses within their team         |
| Admin/Logistics | View all summaries, org-wide stats, and trigger on-demand report generation |

**Validation rules**

- Updates are blocked after the daily cutoff time.
- Updates are blocked for past dates.

**Definition of Done**

- Discord bot responds accurately to all slash commands within 3 seconds.
- Role-based access is enforced — no role can access data outside its scope.
- All participation and location writes are correctly validated against cutoff and date rules.

---

## 7. Key Decisions and Trade-offs

- **DynamoDB as primary data store** — stores all meal participation, work location, headcount, and user records. Chosen for its serverless model, zero idle cost, and natural fit with Lambda's stateless invocation pattern.
- **Fully serverless architecture** — Lambda, API Gateway, and DynamoDB together mean no persistent infrastructure to operate or scale manually. The entire system scales to zero when idle and scales up automatically under load.
- **Discord HTTP Interactions over Gateway (WebSocket) bot** — slash commands delivered as HTTP POST requests require no persistent connection, which is the only model compatible with Lambda. Signature verification via Ed25519 is handled on every incoming request.
- **One Lambda function, one binary** — all routes and Discord command handlers live in a single Go binary behind one API Gateway route. At current scale (~100 Users) there is no operational or cost benefit to splitting into per-function Lambdas, and it would significantly increase infrastructure and deployment complexity.

---
