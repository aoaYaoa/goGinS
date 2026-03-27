DELETE FROM casbin_rule
WHERE ptype = 'p' AND v0 = 'admin' AND v1 = 'audit_logs' AND v2 = 'read';

DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_action;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP INDEX IF EXISTS idx_audit_logs_tenant_id;

DROP TABLE IF EXISTS audit_logs;
