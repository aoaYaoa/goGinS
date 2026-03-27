ALTER TABLE users
    ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE tasks
    ADD COLUMN IF NOT EXISTS tenant_id UUID;

CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users (tenant_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tenant_id ON tasks (tenant_id);
