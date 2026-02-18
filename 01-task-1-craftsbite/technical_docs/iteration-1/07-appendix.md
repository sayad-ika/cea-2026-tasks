## 14. Appendix

### A. Glossary

| Term | Definition |
|------|------------|
| **Opt-In** | Default assumption that employee will participate in meal |
| **Opt-Out** | Employee explicitly indicates they will not participate |
| **Default Meal Preference** | User's preferred default behavior (opt-in or opt-out) for all meals unless explicitly changed |
| **Participation History** | Audit trail of all explicit meal participation changes |
| **Cutoff Time** | Deadline for changing meal participation (9 PM on the previous day for next-day meals) |
| **Headcount** | Total number of employees participating in a meal |
| **Override** | Admin/Team Lead action to modify employee's participation |
| **Participation Resolution** | Process of determining user's meal status by checking day schedule → specific records → bulk opt-outs → default preference in priority order |
| **Auto-Cleanup** | Scheduled job that deletes participation history records older than 3 months |
| **Bulk Opt-Out** | Date range where user is automatically opted-out of all meals |
| **Day Schedule** | Admin-defined status for a specific day (normal, holiday, celebration, office closed) |

### B. Participation Resolution Algorithm

Priority order for determining if a user is participating in a meal:

1. **Check day_schedules:** If day is marked "office_closed", return not participating
2. **Check meal_participations:** If explicit record exists, use its value
3. **Check bulk_opt_outs:** If date falls in range, return not participating
4. **Check default_meal_preference:** Use user's default (opt-in or opt-out)

---

### C. References

- [Go Official Documentation](https://golang.org/doc/)
- [Gin Web Framework](https://gin-gonic.com/docs/)
- [React Documentation](https://react.dev/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP Security Guidelines](https://owasp.org/)

---

**Document Revision History**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | Feb 6, 2026 | Sayad Ibn Khairul Alam | Initial draft |
| 2.0 | Feb 9, 2026 | Sayad Ibn Khairul Alam | Added teams/team_members tables, updated frontend architecture |
| 3.0 | Feb 9, 2026 | Sayad Ibn Khairul Alam | Restructured to follow standard tech doc format, condensed content |
| 4.0 | Feb 10, 2026 | Sayad Ibn Khairul Alam | Corrected participation terminology, clarified cutoff-time enforcement, refined headcount semantics, and aligned performance estimates with realistic internal usage |



---

*End of Document*


