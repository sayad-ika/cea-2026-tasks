# CraftsBite – Technical Design Documentation

This directory contains the complete technical design documentation for the **CraftsBite** system.

The documentation has been intentionally split into focused, review-friendly files while preserving **all original content**. No technical or business details have been omitted.

---

## Document Metadata

- **Project:** CraftsBite
- **Document Type:** Technical Design Document
- **Version:** 5.0
- **Status:** Iteration 2
- **Last Updated:** February 17, 2026
- **Author:** Sayad Ibn Khairul Alam

---

## Reading Guide (Recommended Order)

1. **00-header-summary.md**  
   High-level context, purpose, and system overview.

2. **01-problem-goals.md**  
   Problem statement, goals, and explicit non-goals.

3. **02-scope-requirements.md**  
   Scope definition, functional requirements, and non-functional requirements.

4. **03-user-flows.md**  
   End-to-end user and admin workflows describing system behavior.

5. **04-architecture-design.md**  
   Technical architecture, tech stack rationale, and key design decisions.

6. **05-security-testing.md**  
   Authentication, authorization, RBAC, security controls, and complete testing strategy.

7. **06-operations-risks.md**  
   Deployment, monitoring, maintenance, risks, assumptions, and open questions.

8. **07-appendix.md** _(Reference)_  
   Glossary, algorithms, database indexes, environment variables, and revision history.

---

## Iteration 2 Features

The following features have been added in Iteration 2:

- **Team-based visibility** - Employees see their team, Team Leads view team participation
- **Bulk actions** - Team Leads/Admin can apply bulk opt-out for their scope
- **Enhanced special day controls** - Special Celebration Day with notes
- **Improved headcount reporting** - By meal type, team, Office/WFH split
- **Live updates** - Real-time headcount updates for Admin/Logistics views
- **Daily announcement draft** - Generate copy/paste-friendly messages
- **Work location tracking** - Employees can set Office/WFH per date
- **Company-wide WFH period** - Admin can declare WFH date ranges

---

## Review Notes

- Each file is **self-contained** and can be reviewed independently.
- Section numbering inside files reflects the original document for traceability.
- Changes should be made in the **most relevant file** rather than duplicating content.
- The appendix is reference-only and not required for first-pass review.

---

## Change Policy

- Structural changes to this documentation should preserve section numbering.
- Technical changes must be reflected in both:
  - Design sections
  - Relevant operational or security sections (if applicable)

---

## Questions or Feedback

Please leave comments directly on the relevant file in GitHub.  
For cross-cutting concerns, reference the file name and section number in your feedback.

---
