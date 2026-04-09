package store

func (s *Store) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id          TEXT PRIMARY KEY,
			name        TEXT NOT NULL,
			git_url     TEXT DEFAULT '',
			branch      TEXT DEFAULT 'main',
			provider    TEXT DEFAULT '',
			framework   TEXT DEFAULT '',
			auto_deploy BOOLEAN DEFAULT 1,
			status      TEXT DEFAULT 'active',
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS environments (
			id         TEXT PRIMARY KEY,
			project_id TEXT REFERENCES projects(id),
			key        TEXT NOT NULL,
			value      TEXT NOT NULL,
			is_secret  BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(project_id, key)
		)`,
		`CREATE TABLE IF NOT EXISTS deployments (
			id          TEXT PRIMARY KEY,
			project_id  TEXT REFERENCES projects(id),
			commit_hash TEXT DEFAULT '',
			commit_msg  TEXT DEFAULT '',
			status      TEXT DEFAULT 'queued',
			image_tag   TEXT DEFAULT '',
			trigger     TEXT DEFAULT 'manual',
			started_at  DATETIME,
			finished_at DATETIME,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS deployment_logs (
			id            TEXT PRIMARY KEY,
			deployment_id TEXT REFERENCES deployments(id),
			step          TEXT NOT NULL,
			level         TEXT DEFAULT 'info',
			message       TEXT NOT NULL,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS services (
			id           TEXT PRIMARY KEY,
			name         TEXT NOT NULL,
			type         TEXT NOT NULL,
			image        TEXT NOT NULL,
			status       TEXT DEFAULT 'stopped',
			container_id TEXT DEFAULT '',
			config       TEXT DEFAULT '',
			created_at   DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS service_links (
			id         TEXT PRIMARY KEY,
			project_id TEXT REFERENCES projects(id),
			service_id TEXT REFERENCES services(id),
			env_prefix TEXT DEFAULT 'DATABASE_',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(project_id, service_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_deployments_project ON deployments(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_deployment_logs_deployment ON deployment_logs(deployment_id)`,
		`CREATE INDEX IF NOT EXISTS idx_environments_project ON environments(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_service_links_project ON service_links(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_service_links_service ON service_links(service_id)`,
		`CREATE TABLE IF NOT EXISTS domains (
			id         TEXT PRIMARY KEY,
			project_id TEXT REFERENCES projects(id),
			domain     TEXT UNIQUE NOT NULL,
			ssl_auto   BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id            TEXT PRIMARY KEY,
			username      TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id         TEXT PRIMARY KEY,
			user_id    TEXT REFERENCES users(id),
			token      TEXT UNIQUE NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_domains_project ON domains(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token)`,
		`CREATE TABLE IF NOT EXISTS backups (
			id         TEXT PRIMARY KEY,
			type       TEXT NOT NULL DEFAULT 'system',
			service_id TEXT DEFAULT '',
			filename   TEXT NOT NULL,
			size       INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id          TEXT PRIMARY KEY,
			user_id     TEXT DEFAULT '',
			username    TEXT DEFAULT '',
			action      TEXT NOT NULL,
			resource    TEXT NOT NULL,
			resource_id TEXT DEFAULT '',
			details     TEXT DEFAULT '',
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS invitations (
			id         TEXT PRIMARY KEY,
			email      TEXT NOT NULL,
			role       TEXT DEFAULT 'member',
			token      TEXT UNIQUE NOT NULL,
			used       BOOLEAN DEFAULT 0,
			created_by TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS volumes (
			id         TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			mount_path TEXT NOT NULL,
			project_id TEXT REFERENCES projects(id),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource, resource_id)`,
		`CREATE INDEX IF NOT EXISTS idx_invitations_token ON invitations(token)`,
		`CREATE INDEX IF NOT EXISTS idx_volumes_project ON volumes(project_id)`,
		// Enterprise tables
		`CREATE TABLE IF NOT EXISTS organizations (
			id              TEXT PRIMARY KEY,
			name            TEXT UNIQUE NOT NULL,
			slug            TEXT UNIQUE NOT NULL,
			max_projects    INTEGER DEFAULT 0,
			max_services    INTEGER DEFAULT 0,
			max_cpu         REAL DEFAULT 0,
			max_memory      INTEGER DEFAULT 0,
			max_deployments INTEGER DEFAULT 0,
			created_at      DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at      DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS org_members (
			id      TEXT PRIMARY KEY,
			org_id  TEXT REFERENCES organizations(id),
			user_id TEXT REFERENCES users(id),
			role    TEXT DEFAULT 'member',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(org_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS notification_channels (
			id         TEXT PRIMARY KEY,
			org_id     TEXT DEFAULT '',
			name       TEXT NOT NULL,
			type       TEXT NOT NULL,
			config     TEXT NOT NULL DEFAULT '{}',
			enabled    BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS notification_rules (
			id         TEXT PRIMARY KEY,
			channel_id TEXT REFERENCES notification_channels(id),
			event      TEXT NOT NULL,
			project_id TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(channel_id, event, project_id)
		)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id          TEXT PRIMARY KEY,
			user_id     TEXT REFERENCES users(id),
			name        TEXT NOT NULL,
			key_hash    TEXT UNIQUE NOT NULL,
			key_prefix  TEXT NOT NULL,
			scopes      TEXT DEFAULT '*',
			last_used   DATETIME,
			expires_at  DATETIME,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_org_members_org ON org_members(org_id)`,
		`CREATE INDEX IF NOT EXISTS idx_org_members_user ON org_members(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_hash ON api_keys(key_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_notification_rules_channel ON notification_rules(channel_id)`,
	}

	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}

	// Best-effort column additions for existing tables
	s.db.Exec(`ALTER TABLE projects ADD COLUMN cpu_limit REAL DEFAULT 0`)
	s.db.Exec(`ALTER TABLE projects ADD COLUMN mem_limit INTEGER DEFAULT 0`)
	s.db.Exec(`ALTER TABLE projects ADD COLUMN replicas INTEGER DEFAULT 0`)
	s.db.Exec(`ALTER TABLE projects ADD COLUMN created_by TEXT DEFAULT ''`)
	s.db.Exec(`ALTER TABLE projects ADD COLUMN org_id TEXT DEFAULT ''`)
	s.db.Exec(`ALTER TABLE users ADD COLUMN role TEXT DEFAULT 'admin'`)

	return nil
}
