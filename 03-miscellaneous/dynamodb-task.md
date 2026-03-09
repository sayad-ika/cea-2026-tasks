# Dynamodb task based on the session.

## Summary

All entities live in a single table. Entity type is stored as an attribute on each item — not encoded in the table structure itself.

The design uses **1 GSI** (`GSI1`) shared across all entities. Each entity writes a distinct `GSI1PK` prefix so there is no overlap between patterns — a date like `2026-03-10` can never be confused with a prefixed key like `AUDITEE#` or `WFHPERIOD`.

| Entity             | Items                             | GSI1 Used                           |
| ------------------ | --------------------------------- | ----------------------------------- |
| User               | `USER#`, `EMAIL#`, `DISCORD#`     | No                                  |
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
2.  `PK = EMAIL#<email>` + `SK = LOOKUP` → then fetch `USER#<id>`
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

All team-related items share the same partition key: `PK = TEAM`. The team metadata is stored under `SK = TEAM#<id>#METADATA`, and each team member is stored as a separate item under the same partition with sort key format `TEAM#<id>#MEMBER#<userID>`.

Getting all members of a team is done with a `Query` on `PK = TEAM` and `SK begins_with "TEAM#<id>#MEMBER#"` — this returns every member of that specific team in one call. Checking whether a specific user belongs to a team is a direct `GetItem` on `PK = TEAM`, `SK = TEAM#<id>#MEMBER#<userID>` — if the item exists, the user is a member of that team.

---
