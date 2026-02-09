## 14. Appendix

### A. Glossary

| Term | Definition |
|------|------------|
| **Opt-In** | Default assumption that employee will participate in meal |
| **Opt-Out** | Employee explicitly indicates they will not participate |
| **Default Meal Preference** | User's preferred default behavior (opt-in or opt-out) for all meals unless explicitly changed |
| **Participation History** | Audit trail of all explicit meal participation changes |
| **Cutoff Time** | Deadline for changing meal participation (9 AM lunch/snacks) |
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

### C. Database Indexes

```sql
-- Performance-critical indexes
CREATE INDEX idx_meal_participations_user_date ON meal_participations(user_id, date);
CREATE INDEX idx_meal_participations_date_meal ON meal_participations(date, meal_type);
CREATE INDEX idx_bulk_opt_outs_user_dates ON bulk_opt_outs(user_id, start_date, end_date);
CREATE INDEX idx_history_user_created ON meal_participation_history(user_id, created_at);
CREATE INDEX idx_history_cleanup ON meal_participation_history(created_at);
CREATE INDEX idx_day_schedules_date ON day_schedules(date);
```

### D. Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=craftsbite
DB_USER=craftsbite_user
DB_PASSWORD=<secure_password>

# Application
APP_PORT=8080
APP_ENV=production
JWT_SECRET=<secure_random_string>
JWT_EXPIRY=24h

# CORS
ALLOWED_ORIGINS=https://craftsbite.company.com

# Logging
LOG_LEVEL=info
```

### E. References

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

---

*End of Document*


