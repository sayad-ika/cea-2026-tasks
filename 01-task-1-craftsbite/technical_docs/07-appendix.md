## 14. Appendix

### A. Glossary

| Term                   | Definition                                                                             |
| ---------------------- | -------------------------------------------------------------------------------------- |
| **Opt-In**             | Default assumption that employee will participate in meal                              |
| **Opt-Out**            | Employee explicitly indicates they will not participate                                |
| **Cutoff Time**        | Deadline for changing meal participation (9 PM on the previous day for next-day meals) |
| **Headcount**          | Total number of employees participating in a meal                                      |
| **Override**           | Admin/Team Lead action to modify employee's participation                              |
| **Bulk Opt-Out**       | Date range where user is automatically opted-out of all meals                          |
| **Day Schedule**       | Admin-defined status for a specific day (normal, holiday, celebration, office closed, event day) |
| **Event Day**          | A special day (`day_status=event_day`) with custom available meals and an optional note |
| **Work Location**      | Employee's work location for a specific date: Office or WFH                            |
| **WFH Period**         | Admin-declared date range where all employees are treated as WFH                       |
| **Team View**          | Team Lead's view of team member participation                                          |
| **Live Update**        | Real-time UI update without page refresh via WebSocket                                 |
| **Announcement Draft** | Generated message with meal totals for distribution                                    |
| **Forward Window**     | Maximum number of future days (from today) an employee can set participation for; default 14 |
| **WFH Allowance**      | Soft limit of WFH days per month per employee; default 5; not a hard block             |
| **Over-Limit**         | Employee has used more WFH days than the monthly allowance                             |
| **Forecast Snapshot**  | Admin/Logistics view of projected headcount for upcoming dates                         |

### B. Participation Resolution Algorithm

Priority order for determining if a user is participating in a meal:

1. **Check day_schedules:** If day is marked "office_closed", return not participating
2. **Check meal_participations:** If explicit record exists, use its value
3. **Check bulk_opt_outs:** If date falls in range, return not participating
4. **Check team_bulk_action_entries:** If user affected by team bulk action for the date, return not participating
5. **Check default_meal_preference:** Use user's default (opt-in or opt-out)

> **Note (Iteration 3):** For `event_day` schedules, the available meals list may include `event_dinner`. The resolution algorithm is unchanged; employees can opt in/out for any meal that appears in the schedule's `available_meals` list.

### C. Work Location Resolution Algorithm

Priority order for determining user's work location for a date:

1. **Check wfh_periods:** If date falls in company-wide WFH period, return WFH
2. **Check work_locations:** If explicit record exists, use its value
3. **Use default:** Return Office as default

---

### D. References

- [Go Official Documentation](https://golang.org/doc/)
- [Gin Web Framework](https://gin-gonic.com/docs/)
- [React Documentation](https://react.dev/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)
- [OWASP Security Guidelines](https://owasp.org/)
- [MDN EventSource API](https://developer.mozilla.org/en-US/docs/Web/API/EventSource)
- [HTML Spec — Server-Sent Events](https://html.spec.whatwg.org/multipage/server-sent-events.html)

---

**Document Revision History**

| Version | Date         | Author                 | Changes                                                                                                                                                               |
| ------- | ------------ | ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1.0     | Feb 6, 2026  | Sayad Ibn Khairul Alam | Initial draft                                                                                                                                                         |
| 2.0     | Feb 9, 2026  | Sayad Ibn Khairul Alam | Added teams/team_members tables, updated frontend architecture                                                                                                        |
| 3.0     | Feb 9, 2026  | Sayad Ibn Khairul Alam | Restructured to follow standard tech doc format, condensed content                                                                                                    |
| 4.0     | Feb 10, 2026 | Sayad Ibn Khairul Alam | Corrected participation terminology, clarified cutoff-time enforcement, refined headcount semantics, and aligned performance estimates with realistic internal usage  |
| 5.0     | Feb 17, 2026 | Sayad Ibn Khairul Alam | Added Iteration 2 features: team-based visibility, bulk actions, work location tracking, WFH periods, live updates, announcement drafts, enhanced headcount reporting |
| 6.0     | Feb 21, 2026 | Sayad Ibn Khairul Alam | Added Iteration 3 features: future planning forward window, event_day scheduling, cross-user audit trail, headcount forecast, monthly WFH summary, over-limit indicators and filters |

---

_End of Document_
