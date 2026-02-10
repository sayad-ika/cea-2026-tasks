## 12. Operations

### Deployment Architecture

**Environment:** Single server deployment (initial iteration)
- **Frontend:** Static files served via Nginx
- **Backend:** Go binary running as systemd service
- **Database:** PostgreSQL 17.x
- **Reverse Proxy:** Nginx (SSL termination, static file serving)

### Deployment Process

1. **Build:**
   - Frontend: `npm run build` → static files
   - Backend: `go build` → single binary

2. **Deploy:**
   - Upload static files to `/var/www/craftsbite`
   - Upload Go binary to `/opt/craftsbite/`
   - Restart systemd service: `systemctl restart craftsbite`
   - Nginx reload: `systemctl reload nginx`

3. **Database Migrations:**
   - Run migrations before deploying new binary
   - Use `golang-migrate/migrate` tool
   - Rollback capability for failed migrations

### Monitoring (future iteration)

**Application Metrics:**
- Request latency (p50, p95, p99)
- Error rates (4xx, 5xx)
- Active users
- API endpoint usage

**Infrastructure Metrics:**
- CPU/Memory usage
- Disk space
- Database connection pool
- Network I/O

**Logging:**
- Application logs: Structured JSON (Zap logger)
- Access logs: Nginx format
- Log retention: 30 days
- Log aggregation: Centralized logging system (future)

### Backup and Recovery

**Database Backups:**
- Frequency: Daily at 2 AM
- Retention: 30 days
- Method: pg_dump
- Storage: Off-server backup location

**Recovery Time Objective (RTO):** 4 hours
**Recovery Point Objective (RPO):** 24 hours

### Maintenance

**Scheduled Jobs:**
- **Participation history cleanup:** Daily at 3 AM
  - Delete records older than 3 months
- **Database vacuum:** Weekly on Sunday at 1 AM
  - Reclaim storage, update statistics

**Dependency Updates:**
- Security patches: Apply within 48 hours
- Minor updates: Monthly review
- Major updates: Quarterly review with testing

---


## 13. Risks, Assumptions, and Open Questions

### Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| **Low user adoption** | Medium | High | Comprehensive training, intuitive UI, stakeholder engagement |
| **Performance issues at scale** | Low | Medium | Performance testing before launch, database indexing, query optimization |
| **Cutoff time confusion** | Medium | Medium | Clear UI indicators, email reminders, grace period for first week |
| **Security breach** | Low | High | Security best practices, regular audits, HTTPS only, input validation |
| **Database corruption** | Low | High | Daily backups, transaction support, database replication (future) |
| **Team lead override abuse** | Low | Medium | Audit logging, monthly access reviews, change notifications |

### Assumptions

1. **Network Connectivity:** Users have stable internet access during work hours
2. **Browser Support:** Users have modern browsers (Chrome 100+, Firefox 100+, Safari 15+, Edge 100+)
3. **User Base:** Maximum 100 concurrent users for iteration 1
4. **Data Volume:** Average 200 meal participation changes per day
5. **Cutoff Times:** 9 AM for lunch/snacks
6. **Working Days:** Monday-Friday (weekends assumed office closed unless specified)
7. **Single Organization:** No multi-tenancy required
9. **Role Assignment:** User roles managed manually by admin (no self-registration)
10. **Language:** English only (no i18n required for iteration 1)

### Open Questions

1. **Q:** Should we send email notifications for participation reminders?
   **Status:** Deferred to iteration 2  
   **Decision needed by:** N/A

2. **Q:** Should we allow employees to see team-wide headcount?
   **Status:** Open – privacy concerns  
   **Decision needed by:** Before launch

3. **Q:** How should conflicting date-range participation rules (overlapping ranges) be handled?
   **Status:** Last-created rule wins (current implementation)  
   **Decision needed by:** Resolved

4. **Q:** Should Team Leads be able to apply date-range participation changes to their entire team?
   **Status:** No for iteration 1 – individual overrides only  
   **Decision needed by:** Resolved

5. **Q:** What happens to meal participation when a user’s role changes?
   **Status:** Permissions change immediately; historical data preserved  
   **Decision needed by:** Resolved

6. **Q:** Should we implement meal preferences or favorites?
   **Status:** Out of scope for iteration 1  
   **Decision needed by:** Out of scope

7. **Q:** How granular should audit logs be (field-level vs record-level)?
   **Status:** Record-level (sufficient for accountability)  
   **Decision needed by:** Resolved

8. **Q:** Should the system support recurring date-range participation rules (e.g., every Friday)?
   **Status:** No for iteration 1 – manual creation only  
   **Decision needed by:** N/A

---
