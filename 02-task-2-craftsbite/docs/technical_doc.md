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
