# Dynamodb task based on the session.

## Summary

All entities live in a single table. Entity type is stored as an attribute on each item — not encoded in the table structure itself.

The design uses **1 GSI** (`GSI1`) shared across all entities. Each entity writes a distinct `GSI1PK` prefix so there is no overlap between patterns — a date like `2026-03-10` can never be confused with a prefixed key like `AUDITEE#` or `WFHPERIOD`.

| Entity             | Items                             | GSI1 Used                           |
| ------------------ | --------------------------------- | ----------------------------------- |
| User               | `USER#`, `EMAIL#`, `DISCORD#`     | Yes — active user listing           |
| Team               | `TEAM# METADATA`, `TEAM# MEMBER#` | No                                  |
| Meal Participation | `USER# MEAL#<date>#<mealType>`    | Yes — headcount by date             |
| Work Location      | `USER# WORKLOCATION#<date>`       | Yes — WFH/Office split by date      |
| Day & Meals        | `DAY# METADATA`, `DAY# MEALS`     | No                                  |
| WFH Period         | `WFHPERIOD`                       | No                                  |
| Audit Log          | `AUDIT#<actorUserID>`             | Yes — reverse lookup by target user |

---

## User

### Access Patterns

1. Get user profile
2. Login by email
3. Bot auth by Discord ID

### DB Schema

```
1.  `PK = USER#<id>` -- `SK = PROFILE`
2.  `PK = EMAIL#<email>` + `SK = LOOKUP` → then fetch `USER#<id>` `PK = EMAIL#<email>` + `SK = LOOKUP` → then fetch `USER#<id>`
3.  `PK = DISCORD#<discordId>` -- `SK = LOOKUP`
```

Each user gets three items in the table. The `USER#` item is the main profile and holds everything about the user — name, email, role, and so on. The `EMAIL#` and `DISCORD#` items are lightweight lookup rows that only store the user's internal ID, so we can find a user by email or Discord ID without needing a GSI.

The `DISCORD#` lookup also stores the user's role directly on it. This means when the bot receives a request, a single `GetItem` on `DISCORD#<id>` gives us both the user ID and the role — no second read on the profile needed.

---

## Team

### Access Patterns

1. Get team details
2. Get all team members
3. Check if user is in team
4. Get all users grouped by all teams

### DB Schema

```
1. `PK = TEAM` + `SK = TEAM#<id>#METADATA`
2. `PK = TEAM` + `SK begins_with TEAM#<id>#MEMBER#`
3. `PK = TEAM` + `SK = TEAM#<id>#MEMBER#<userID>`
4. `PK = TEAM` + `SK begins_with TEAM#`
```

A team shares one partition key for all its items. The `METADATA` item holds the team details. Each member gets their own `MEMBER#<userID>` item under the same partition.

Getting all members is a `Query` on `PK = TEAM#<id>` with `SK begins_with "MEMBER#"` — it returns every member in one call. Checking if a specific user is in the team is a direct `GetItem` on `PK = TEAM#<id>`, `SK = MEMBER#<userID>` — if the item exists, they're a member.

---
