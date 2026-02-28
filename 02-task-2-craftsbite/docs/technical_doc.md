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
