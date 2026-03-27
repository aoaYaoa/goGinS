CREATE TABLE IF NOT EXISTS casbin_rule (
    id    BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL,
    v0    VARCHAR(100) NOT NULL DEFAULT '',
    v1    VARCHAR(100) NOT NULL DEFAULT '',
    v2    VARCHAR(100) NOT NULL DEFAULT '',
    v3    VARCHAR(100) NOT NULL DEFAULT '',
    v4    VARCHAR(100) NOT NULL DEFAULT '',
    v5    VARCHAR(100) NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_casbin_rule_unique
    ON casbin_rule (ptype, v0, v1, v2, v3, v4, v5);

INSERT INTO casbin_rule (ptype, v0, v1, v2)
VALUES
    ('p', 'admin', 'users', 'read'),
    ('p', 'admin', 'users', 'write'),
    ('p', 'admin', 'tasks', 'read'),
    ('p', 'admin', 'tasks', 'write')
ON CONFLICT DO NOTHING;
