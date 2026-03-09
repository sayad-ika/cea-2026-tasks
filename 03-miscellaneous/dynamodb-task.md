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

## Meal Participation

### Access Patterns

1. Get user's all meals for a date
2. Get user's specific meal
3. Opt in/out of a meal
4. All participation for a date

### DB Schema

```
1. `PK = USER#<id>` + `SK begins_with MEAL#<date>`
2. `PK = USER#<id>` + `SK = MEAL#<date>#<meal_type>`
3. `PK = USER#<id>`+`SK = MEAL#<date>#<meal_type>`
4. `GSI1_PK = <date>`
```

Each participation record is keyed by user, date, and meal type together in the sort key. Getting all meals for a user on a date is a `Query` with `SK begins_with "MEAL#<date>#"`. Getting one specific meal is a direct `GetItem`.

Writing participation is always a `PutItem` — it either creates or replaces the record, so calling it multiple times is safe.

For headcount (pattern 4), participation items write to GSI1:

```
GSI1PK = <date>
GSI1SK = MEAL#<userID>
```

Querying `GSI1PK = <date>` with `GSI1SK begins_with "MEAL#"` returns every participation record for that date across all users in one call.

---

## Work Location

### Access Patterns

1. Get user's location for a date
2. Set user's location
3. All WFH employees on a date
4. Monthly WFH count for a user

### DB Schema

```
1. `PK = USER#<id>` + `SK = WORKLOCATION#<date>`
2. `PUT PK = USER#<id>` + `SK = WORKLOCATION#<date>`
3. `GSI1_PK = <date>` + `SK begins_with WFH#`
4. `PK = USER#<id>` + `SK begins_with WORKLOCATION#<year-month>`
```

One item per user per date. Reading is a direct `GetItem`. Writing is a `PutItem` — same upsert pattern as meal participation.

For getting all WFH employees on a date (pattern 3), work location items write to GSI1:

```
GSI1PK = <date>
GSI1SK = WFH#<userID>     or     OFFICE#<userID>
```

The location value is embedded at the start of the sort key. This lets us filter to WFH-only employees using `GSI1SK begins_with "WFH#"` as a key condition — no filter expression needed.

For monthly WFH count (pattern 4), we query the user's own partition:

```
PK = USER#<id>    SK begins_with "WORKLOCATION#<YYYY-MM>"
```

This returns all location records for that user in a given month. We then count the ones where `location = "wfh"` in the application.

---
