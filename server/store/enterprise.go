package store

import (
	"database/sql"
	"time"

	"github.com/my-paas/server/model"
)

// --- Organizations ---

func (s *Store) CreateOrganization(input model.CreateOrgInput) (*model.Organization, error) {
	org := &model.Organization{
		ID:             newID(),
		Name:           input.Name,
		Slug:           input.Slug,
		MaxProjects:    input.MaxProjects,
		MaxServices:    input.MaxServices,
		MaxCPU:         input.MaxCPU,
		MaxMemory:      input.MaxMemory,
		MaxDeployments: input.MaxDeployments,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err := s.db.Exec(`
		INSERT INTO organizations (id, name, slug, max_projects, max_services, max_cpu, max_memory, max_deployments, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		org.ID, org.Name, org.Slug, org.MaxProjects, org.MaxServices, org.MaxCPU, org.MaxMemory, org.MaxDeployments, org.CreatedAt, org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return org, nil
}

func (s *Store) GetOrganization(id string) (*model.Organization, error) {
	org := &model.Organization{}
	err := s.db.QueryRow(`
		SELECT id, name, slug, max_projects, max_services, max_cpu, max_memory, max_deployments, created_at, updated_at
		FROM organizations WHERE id = ?`, id,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.MaxProjects, &org.MaxServices, &org.MaxCPU, &org.MaxMemory, &org.MaxDeployments, &org.CreatedAt, &org.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return org, err
}

func (s *Store) GetOrganizationBySlug(slug string) (*model.Organization, error) {
	org := &model.Organization{}
	err := s.db.QueryRow(`
		SELECT id, name, slug, max_projects, max_services, max_cpu, max_memory, max_deployments, created_at, updated_at
		FROM organizations WHERE slug = ?`, slug,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.MaxProjects, &org.MaxServices, &org.MaxCPU, &org.MaxMemory, &org.MaxDeployments, &org.CreatedAt, &org.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return org, err
}

func (s *Store) ListOrganizations() ([]model.Organization, error) {
	rows, err := s.db.Query(`
		SELECT id, name, slug, max_projects, max_services, max_cpu, max_memory, max_deployments, created_at, updated_at
		FROM organizations ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []model.Organization
	for rows.Next() {
		var org model.Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.MaxProjects, &org.MaxServices, &org.MaxCPU, &org.MaxMemory, &org.MaxDeployments, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (s *Store) UpdateOrganization(id string, input model.UpdateOrgInput) (*model.Organization, error) {
	org, err := s.GetOrganization(id)
	if err != nil || org == nil {
		return nil, err
	}
	if input.Name != nil {
		org.Name = *input.Name
	}
	if input.MaxProjects != nil {
		org.MaxProjects = *input.MaxProjects
	}
	if input.MaxServices != nil {
		org.MaxServices = *input.MaxServices
	}
	if input.MaxCPU != nil {
		org.MaxCPU = *input.MaxCPU
	}
	if input.MaxMemory != nil {
		org.MaxMemory = *input.MaxMemory
	}
	if input.MaxDeployments != nil {
		org.MaxDeployments = *input.MaxDeployments
	}
	org.UpdatedAt = time.Now()

	_, err = s.db.Exec(`
		UPDATE organizations SET name=?, max_projects=?, max_services=?, max_cpu=?, max_memory=?, max_deployments=?, updated_at=? WHERE id=?`,
		org.Name, org.MaxProjects, org.MaxServices, org.MaxCPU, org.MaxMemory, org.MaxDeployments, org.UpdatedAt, org.ID,
	)
	return org, err
}

func (s *Store) DeleteOrganization(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec(`DELETE FROM org_members WHERE org_id=?`, id)
	tx.Exec(`UPDATE projects SET org_id='' WHERE org_id=?`, id)
	tx.Exec(`DELETE FROM notification_channels WHERE org_id=?`, id)
	tx.Exec(`DELETE FROM organizations WHERE id=?`, id)

	return tx.Commit()
}

// --- Org Members ---

func (s *Store) AddOrgMember(orgID, userID, role string) error {
	_, err := s.db.Exec(`
		INSERT INTO org_members (id, org_id, user_id, role, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(org_id, user_id) DO UPDATE SET role=excluded.role`,
		newID(), orgID, userID, role, time.Now(),
	)
	return err
}

func (s *Store) RemoveOrgMember(orgID, userID string) error {
	_, err := s.db.Exec(`DELETE FROM org_members WHERE org_id=? AND user_id=?`, orgID, userID)
	return err
}

func (s *Store) ListOrgMembers(orgID string) ([]model.OrgMember, error) {
	rows, err := s.db.Query(`
		SELECT om.id, om.org_id, om.user_id, om.role, om.created_at, u.username
		FROM org_members om
		INNER JOIN users u ON om.user_id = u.id
		WHERE om.org_id = ? ORDER BY om.created_at ASC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.OrgMember
	for rows.Next() {
		var m model.OrgMember
		if err := rows.Scan(&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.CreatedAt, &m.Username); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (s *Store) GetUserOrganizations(userID string) ([]model.Organization, error) {
	rows, err := s.db.Query(`
		SELECT o.id, o.name, o.slug, o.max_projects, o.max_services, o.max_cpu, o.max_memory, o.max_deployments, o.created_at, o.updated_at
		FROM organizations o
		INNER JOIN org_members om ON o.id = om.org_id
		WHERE om.user_id = ? ORDER BY o.name ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []model.Organization
	for rows.Next() {
		var org model.Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.MaxProjects, &org.MaxServices, &org.MaxCPU, &org.MaxMemory, &org.MaxDeployments, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (s *Store) GetOrgProjectCount(orgID string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM projects WHERE org_id = ?`, orgID).Scan(&count)
	return count, err
}

func (s *Store) GetOrgServiceCount(orgID string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM services s
		INNER JOIN service_links sl ON s.id = sl.service_id
		INNER JOIN projects p ON sl.project_id = p.id
		WHERE p.org_id = ?`, orgID).Scan(&count)
	return count, err
}

// --- API Keys ---

func (s *Store) CreateAPIKey(key *model.APIKey) error {
	_, err := s.db.Exec(`
		INSERT INTO api_keys (id, user_id, name, key_hash, key_prefix, scopes, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		key.ID, key.UserID, key.Name, key.KeyHash, key.KeyPrefix, key.Scopes, key.ExpiresAt, key.CreatedAt,
	)
	return err
}

func (s *Store) GetAPIKeyByHash(keyHash string) (*model.APIKey, error) {
	key := &model.APIKey{}
	err := s.db.QueryRow(`
		SELECT id, user_id, name, key_hash, key_prefix, scopes, last_used, expires_at, created_at
		FROM api_keys WHERE key_hash = ?`, keyHash,
	).Scan(&key.ID, &key.UserID, &key.Name, &key.KeyHash, &key.KeyPrefix, &key.Scopes, &key.LastUsed, &key.ExpiresAt, &key.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return key, err
}

func (s *Store) ListAPIKeys(userID string) ([]model.APIKey, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, name, key_hash, key_prefix, scopes, last_used, expires_at, created_at
		FROM api_keys WHERE user_id = ? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []model.APIKey
	for rows.Next() {
		var k model.APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.Name, &k.KeyHash, &k.KeyPrefix, &k.Scopes, &k.LastUsed, &k.ExpiresAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *Store) UpdateAPIKeyLastUsed(id string) error {
	_, err := s.db.Exec(`UPDATE api_keys SET last_used=? WHERE id=?`, time.Now(), id)
	return err
}

func (s *Store) DeleteAPIKey(id string) error {
	_, err := s.db.Exec(`DELETE FROM api_keys WHERE id=?`, id)
	return err
}

// --- Notification Channels ---

func (s *Store) CreateNotificationChannel(ch *model.NotificationChannel) error {
	_, err := s.db.Exec(`
		INSERT INTO notification_channels (id, org_id, name, type, config, enabled, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		ch.ID, ch.OrgID, ch.Name, ch.Type, ch.Config, ch.Enabled, ch.CreatedAt,
	)
	return err
}

func (s *Store) GetNotificationChannel(id string) (*model.NotificationChannel, error) {
	ch := &model.NotificationChannel{}
	err := s.db.QueryRow(`
		SELECT id, org_id, name, type, config, enabled, created_at
		FROM notification_channels WHERE id = ?`, id,
	).Scan(&ch.ID, &ch.OrgID, &ch.Name, &ch.Type, &ch.Config, &ch.Enabled, &ch.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return ch, err
}

func (s *Store) ListNotificationChannels(orgID string) ([]model.NotificationChannel, error) {
	query := `SELECT id, org_id, name, type, config, enabled, created_at FROM notification_channels`
	var rows *sql.Rows
	var err error
	if orgID != "" {
		rows, err = s.db.Query(query+" WHERE org_id = ? ORDER BY created_at ASC", orgID)
	} else {
		rows, err = s.db.Query(query + " ORDER BY created_at ASC")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []model.NotificationChannel
	for rows.Next() {
		var ch model.NotificationChannel
		if err := rows.Scan(&ch.ID, &ch.OrgID, &ch.Name, &ch.Type, &ch.Config, &ch.Enabled, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

func (s *Store) UpdateNotificationChannel(id string, enabled bool) error {
	_, err := s.db.Exec(`UPDATE notification_channels SET enabled=? WHERE id=?`, enabled, id)
	return err
}

func (s *Store) DeleteNotificationChannel(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec(`DELETE FROM notification_rules WHERE channel_id=?`, id)
	tx.Exec(`DELETE FROM notification_channels WHERE id=?`, id)
	return tx.Commit()
}

// --- Notification Rules ---

func (s *Store) CreateNotificationRule(rule *model.NotificationRule) error {
	_, err := s.db.Exec(`
		INSERT INTO notification_rules (id, channel_id, event, project_id, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(channel_id, event, project_id) DO NOTHING`,
		rule.ID, rule.ChannelID, rule.Event, rule.ProjectID, rule.CreatedAt,
	)
	return err
}

func (s *Store) ListNotificationRules(channelID string) ([]model.NotificationRule, error) {
	rows, err := s.db.Query(`
		SELECT id, channel_id, event, project_id, created_at
		FROM notification_rules WHERE channel_id = ?`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []model.NotificationRule
	for rows.Next() {
		var r model.NotificationRule
		if err := rows.Scan(&r.ID, &r.ChannelID, &r.Event, &r.ProjectID, &r.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *Store) GetNotificationRulesForEvent(event, projectID string) ([]model.NotificationChannel, error) {
	rows, err := s.db.Query(`
		SELECT nc.id, nc.org_id, nc.name, nc.type, nc.config, nc.enabled, nc.created_at
		FROM notification_channels nc
		INNER JOIN notification_rules nr ON nc.id = nr.channel_id
		WHERE nc.enabled = 1 AND nr.event = ? AND (nr.project_id = '' OR nr.project_id = ?)`, event, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []model.NotificationChannel
	for rows.Next() {
		var ch model.NotificationChannel
		if err := rows.Scan(&ch.ID, &ch.OrgID, &ch.Name, &ch.Type, &ch.Config, &ch.Enabled, &ch.CreatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, ch)
	}
	return channels, nil
}

func (s *Store) DeleteNotificationRule(id string) error {
	_, err := s.db.Exec(`DELETE FROM notification_rules WHERE id=?`, id)
	return err
}

// --- Deployment count for quotas ---

func (s *Store) GetMonthlyDeploymentCount(orgID string) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM deployments d
		INNER JOIN projects p ON d.project_id = p.id
		WHERE p.org_id = ? AND d.created_at >= datetime('now', '-30 days')`, orgID).Scan(&count)
	return count, err
}
