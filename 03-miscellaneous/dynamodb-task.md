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

## Day & Meals

### Access Patterns

1. Get full day context
2. Get day type only
3. Get available meals only
4. Set day type
5. Set available meals
6. Get full context for a range of dates

### DB Schema

```
1. `PK = DAY`  +  `SK begins_with <date>`
2. `PK = DAY`  +  `SK = <date>#METADATA`
3. `PK = DAY`  +  `SK = <date>#MEALS`
4. `PUT PK = DAY`  +  `SK = <date>#METADATA`
5. `PUT PK = DAY`  +  `SK = <date>#MEALS`
6. `PK = DAY`  +  `SK BETWEEN <startDate> AND <endDate>#zzzz`
```

All day items share a single static partition key `DAY`. Moving the date into the sort key means every day's data lives in the same partition, which allows native range queries across dates.

A single `Query PK = DAY` with `SK begins_with <date>` returns both the `METADATA` and `MEALS` rows for that day in one round trip. Writing either row is a `PutItem` directly on its sort key.

The `METADATA` item stores the day status (e.g. `normal`, `office_closed`, `govt_holiday`) and an optional note. The `MEALS` item stores the list of available meal types for that day.

For pattern 6, a `BETWEEN` condition on the sort key covers the full date window in a single `Query` — no `BatchGetItem` needed.

---

## WFH Period

### Access Patterns

1. List all WFH periods
2. Is date in any WFH period?

### DB Schema

```
1. `PK = WFHPERIOD`
2. `PK = WFHPERIOD` + `SK begins_with`
```

All WFH periods share a single static partition key `WFHPERIOD`. There are only about 4–6 per year so a single partition is completely fine here.

The sort key is a composite of start and end date. Since dates are in `YYYY-MM-DD` format they sort lexicographically, so records naturally come back ordered by start date on every query.

For checking if a date falls within a period (pattern 2), we query:

```
PK = WFHPERIOD    SK <= "<date>#zzzz"
```

This returns all periods that started on or before the target date. We then check in the application whether any returned period has `end_date >= date`. There are at most a handful of records to evaluate.

---

## Audit Log

### Access Patterns

1. Write an audit entry on every mutation
2. Get all changes made by a specific user
3. Get all changes made to a specific user's records
4. Get changes of a specific entity type by a user

### DB Schema

```
1. PK = AUDIT#<actorUserID>    SK = <timestamp>#<entityType>#<entityKey>
```

Every write action in the system appends a record here — `PutItem` only, never updated or deleted.

Scoping by actor (`AUDIT#<actorUserID>`) keeps all of a user's actions co-located under one partition. The sort key starts with a timestamp so records naturally come back in chronological order when queried.

For getting all changes made by a user (pattern 2), it's just a `Query` on `PK = AUDIT#<actorUserID>`.

For getting all changes made _to_ a specific user's records regardless of who made them (pattern 3), audit items write to GSI1:

```
GSI1PK = AUDITEE#<targetUserID>
GSI1SK = <timestamp>
```

For filtering by entity type (pattern 4), because the sort key starts with a timestamp we can't use `begins_with "<entityType>"` as a key condition. Instead we query the full actor partition and apply `FilterExpression: entityType = :type` in the application to narrow it down.

---
