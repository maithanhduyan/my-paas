package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"

	"github.com/my-paas/server/model"
)

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func newID() string {
	return uuid.New().String()[:8]
}

func NewID() string {
	return newID()
}

// --- Projects ---

func (s *Store) CreateProject(input model.CreateProjectInput) (*model.Project, error) {
	p := &model.Project{
		ID:         newID(),
		Name:       input.Name,
		GitURL:     input.GitURL,
		Branch:     input.Branch,
		AutoDeploy: true,
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if p.Branch == "" {
		p.Branch = "main"
	}
	_, err := s.db.Exec(`
		INSERT INTO projects (id, name, git_url, branch, provider, framework, auto_deploy, status, cpu_limit, mem_limit, replicas, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.GitURL, p.Branch, p.Provider, p.Framework, p.AutoDeploy, p.Status, p.CPULimit, p.MemLimit, p.Replicas, p.CreatedBy, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Store) GetProject(id string) (*model.Project, error) {
	p := &model.Project{}
	err := s.db.QueryRow(`
		SELECT id, name, git_url, branch, provider, framework, auto_deploy, status, cpu_limit, mem_limit, replicas, created_by, created_at, updated_at
		FROM projects WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.GitURL, &p.Branch, &p.Provider, &p.Framework, &p.AutoDeploy, &p.Status, &p.CPULimit, &p.MemLimit, &p.Replicas, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return p, err
}

func (s *Store) ListProjects() ([]model.Project, error) {
	rows, err := s.db.Query(`
		SELECT id, name, git_url, branch, provider, framework, auto_deploy, status, cpu_limit, mem_limit, replicas, created_by, created_at, updated_at
		FROM projects ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.GitURL, &p.Branch, &p.Provider, &p.Framework, &p.AutoDeploy, &p.Status, &p.CPULimit, &p.MemLimit, &p.Replicas, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *Store) UpdateProject(id string, input model.UpdateProjectInput) (*model.Project, error) {
	p, err := s.GetProject(id)
	if err != nil || p == nil {
		return nil, fmt.Errorf("project not found")
	}
	if input.Name != nil {
		p.Name = *input.Name
	}
	if input.Branch != nil {
		p.Branch = *input.Branch
	}
	if input.AutoDeploy != nil {
		p.AutoDeploy = *input.AutoDeploy
	}
	if input.CPULimit != nil {
		p.CPULimit = *input.CPULimit
	}
	if input.MemLimit != nil {
		p.MemLimit = *input.MemLimit
	}
	if input.Replicas != nil {
		p.Replicas = *input.Replicas
	}
	p.UpdatedAt = time.Now()
	_, err = s.db.Exec(`
		UPDATE projects SET name=?, branch=?, auto_deploy=?, cpu_limit=?, mem_limit=?, replicas=?, updated_at=? WHERE id=?`,
		p.Name, p.Branch, p.AutoDeploy, p.CPULimit, p.MemLimit, p.Replicas, p.UpdatedAt, p.ID,
	)
	return p, err
}

func (s *Store) UpdateProjectDetection(id, provider, framework string) error {
	_, err := s.db.Exec(`UPDATE projects SET provider=?, framework=?, updated_at=? WHERE id=?`,
		provider, framework, time.Now(), id)
	return err
}

func (s *Store) UpdateProjectStatus(id, status string) error {
	_, err := s.db.Exec(`UPDATE projects SET status=?, updated_at=? WHERE id=?`,
		status, time.Now(), id)
	return err
}

func (s *Store) DeleteProject(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec(`DELETE FROM deployment_logs WHERE deployment_id IN (SELECT id FROM deployments WHERE project_id=?)`, id)
	tx.Exec(`DELETE FROM deployments WHERE project_id=?`, id)
	tx.Exec(`DELETE FROM environments WHERE project_id=?`, id)
	tx.Exec(`DELETE FROM domains WHERE project_id=?`, id)
	tx.Exec(`DELETE FROM service_links WHERE project_id=?`, id)
	tx.Exec(`DELETE FROM projects WHERE id=?`, id)

	return tx.Commit()
}

// --- Deployments ---

func (s *Store) CreateDeployment(projectID string, trigger model.DeployTrigger) (*model.Deployment, error) {
	now := time.Now()
	d := &model.Deployment{
		ID:        newID(),
		ProjectID: projectID,
		Status:    model.DeployQueued,
		Trigger:   trigger,
		CreatedAt: now,
	}
	_, err := s.db.Exec(`
		INSERT INTO deployments (id, project_id, commit_hash, commit_msg, status, image_tag, trigger, started_at, finished_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		d.ID, d.ProjectID, d.CommitHash, d.CommitMsg, d.Status, d.ImageTag, d.Trigger, d.StartedAt, d.FinishedAt, d.CreatedAt,
	)
	return d, err
}

func (s *Store) GetDeployment(id string) (*model.Deployment, error) {
	d := &model.Deployment{}
	err := s.db.QueryRow(`
		SELECT id, project_id, commit_hash, commit_msg, status, image_tag, trigger, started_at, finished_at, created_at
		FROM deployments WHERE id = ?`, id,
	).Scan(&d.ID, &d.ProjectID, &d.CommitHash, &d.CommitMsg, &d.Status, &d.ImageTag, &d.Trigger, &d.StartedAt, &d.FinishedAt, &d.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return d, err
}

func (s *Store) ListDeployments(projectID string) ([]model.Deployment, error) {
	rows, err := s.db.Query(`
		SELECT id, project_id, commit_hash, commit_msg, status, image_tag, trigger, started_at, finished_at, created_at
		FROM deployments WHERE project_id = ? ORDER BY created_at DESC LIMIT 50`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []model.Deployment
	for rows.Next() {
		var d model.Deployment
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.CommitHash, &d.CommitMsg, &d.Status, &d.ImageTag, &d.Trigger, &d.StartedAt, &d.FinishedAt, &d.CreatedAt); err != nil {
			return nil, err
		}
		deployments = append(deployments, d)
	}
	return deployments, nil
}

func (s *Store) UpdateDeploymentStatus(id string, status model.DeploymentStatus) error {
	q := `UPDATE deployments SET status=? WHERE id=?`
	if status == model.DeployCloning {
		now := time.Now()
		q = `UPDATE deployments SET status=?, started_at=? WHERE id=?`
		_, err := s.db.Exec(q, status, now, id)
		return err
	}
	if status == model.DeployHealthy || status == model.DeployFailed || status == model.DeployRolledBack || status == model.DeployCancelled {
		now := time.Now()
		q = `UPDATE deployments SET status=?, finished_at=? WHERE id=?`
		_, err := s.db.Exec(q, status, now, id)
		return err
	}
	_, err := s.db.Exec(q, status, id)
	return err
}

func (s *Store) UpdateDeploymentImage(id, imageTag string) error {
	_, err := s.db.Exec(`UPDATE deployments SET image_tag=? WHERE id=?`, imageTag, id)
	return err
}

func (s *Store) UpdateDeploymentCommit(id, hash, msg string) error {
	_, err := s.db.Exec(`UPDATE deployments SET commit_hash=?, commit_msg=? WHERE id=?`, hash, msg, id)
	return err
}

func (s *Store) GetLastHealthyDeployment(projectID string) (*model.Deployment, error) {
	d := &model.Deployment{}
	err := s.db.QueryRow(`
		SELECT id, project_id, commit_hash, commit_msg, status, image_tag, trigger, started_at, finished_at, created_at
		FROM deployments WHERE project_id = ? AND status = 'healthy' ORDER BY created_at DESC LIMIT 1`, projectID,
	).Scan(&d.ID, &d.ProjectID, &d.CommitHash, &d.CommitMsg, &d.Status, &d.ImageTag, &d.Trigger, &d.StartedAt, &d.FinishedAt, &d.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return d, err
}

// --- Deployment Logs ---

func (s *Store) AddDeploymentLog(deploymentID, step, level, message string) error {
	_, err := s.db.Exec(`
		INSERT INTO deployment_logs (id, deployment_id, step, level, message, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		newID(), deploymentID, step, level, message, time.Now(),
	)
	return err
}

func (s *Store) GetDeploymentLogs(deploymentID string) ([]model.DeploymentLog, error) {
	rows, err := s.db.Query(`
		SELECT id, deployment_id, step, level, message, created_at
		FROM deployment_logs WHERE deployment_id = ? ORDER BY created_at ASC`, deploymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.DeploymentLog
	for rows.Next() {
		var l model.DeploymentLog
		if err := rows.Scan(&l.ID, &l.DeploymentID, &l.Step, &l.Level, &l.Message, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// --- Environments ---

func (s *Store) SetEnvVars(projectID string, vars []model.EnvVarInput) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, v := range vars {
		_, err := tx.Exec(`
			INSERT INTO environments (id, project_id, key, value, is_secret, created_at)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(project_id, key) DO UPDATE SET value=excluded.value, is_secret=excluded.is_secret`,
			newID(), projectID, v.Key, v.Value, v.IsSecret, time.Now(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) GetEnvVars(projectID string) ([]model.EnvVar, error) {
	rows, err := s.db.Query(`
		SELECT id, project_id, key, value, is_secret, created_at
		FROM environments WHERE project_id = ? ORDER BY key ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vars []model.EnvVar
	for rows.Next() {
		var v model.EnvVar
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.Key, &v.Value, &v.IsSecret, &v.CreatedAt); err != nil {
			return nil, err
		}
		vars = append(vars, v)
	}
	return vars, nil
}

func (s *Store) DeleteEnvVar(projectID, key string) error {
	_, err := s.db.Exec(`DELETE FROM environments WHERE project_id=? AND key=?`, projectID, key)
	return err
}

// --- Services ---

func (s *Store) CreateService(input model.CreateServiceInput) (*model.Service, error) {
	image := input.Image
	if image == "" {
		defaultImage, _, _ := model.ServiceDefaults(input.Type)
		if defaultImage == "" {
			return nil, fmt.Errorf("unsupported service type: %s", input.Type)
		}
		image = defaultImage
	}

	svc := &model.Service{
		ID:        newID(),
		Name:      input.Name,
		Type:      input.Type,
		Image:     image,
		Status:    "stopped",
		CreatedAt: time.Now(),
	}
	_, err := s.db.Exec(`
		INSERT INTO services (id, name, type, image, status, container_id, config, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		svc.ID, svc.Name, svc.Type, svc.Image, svc.Status, svc.ContainerID, svc.Config, svc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (s *Store) GetService(id string) (*model.Service, error) {
	svc := &model.Service{}
	err := s.db.QueryRow(`
		SELECT id, name, type, image, status, container_id, config, created_at
		FROM services WHERE id = ?`, id,
	).Scan(&svc.ID, &svc.Name, &svc.Type, &svc.Image, &svc.Status, &svc.ContainerID, &svc.Config, &svc.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return svc, err
}

func (s *Store) ListServices() ([]model.Service, error) {
	rows, err := s.db.Query(`
		SELECT id, name, type, image, status, container_id, config, created_at
		FROM services ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []model.Service
	for rows.Next() {
		var svc model.Service
		if err := rows.Scan(&svc.ID, &svc.Name, &svc.Type, &svc.Image, &svc.Status, &svc.ContainerID, &svc.Config, &svc.CreatedAt); err != nil {
			return nil, err
		}
		services = append(services, svc)
	}
	return services, nil
}

func (s *Store) UpdateServiceStatus(id, status, containerID string) error {
	_, err := s.db.Exec(`UPDATE services SET status=?, container_id=? WHERE id=?`, status, containerID, id)
	return err
}

func (s *Store) DeleteService(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.Exec(`DELETE FROM service_links WHERE service_id=?`, id)
	tx.Exec(`DELETE FROM services WHERE id=?`, id)

	return tx.Commit()
}

// --- Service Links ---

func (s *Store) LinkService(projectID, serviceID, envPrefix string) (*model.ServiceLink, error) {
	link := &model.ServiceLink{
		ID:        newID(),
		ProjectID: projectID,
		ServiceID: serviceID,
		EnvPrefix: envPrefix,
		CreatedAt: time.Now(),
	}
	_, err := s.db.Exec(`
		INSERT INTO service_links (id, project_id, service_id, env_prefix, created_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(project_id, service_id) DO UPDATE SET env_prefix=excluded.env_prefix`,
		link.ID, link.ProjectID, link.ServiceID, link.EnvPrefix, link.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return link, nil
}

func (s *Store) UnlinkService(projectID, serviceID string) error {
	_, err := s.db.Exec(`DELETE FROM service_links WHERE project_id=? AND service_id=?`, projectID, serviceID)
	return err
}

func (s *Store) GetServiceLinks(projectID string) ([]model.ServiceLink, error) {
	rows, err := s.db.Query(`
		SELECT id, project_id, service_id, env_prefix, created_at
		FROM service_links WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.ServiceLink
	for rows.Next() {
		var l model.ServiceLink
		if err := rows.Scan(&l.ID, &l.ProjectID, &l.ServiceID, &l.EnvPrefix, &l.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, nil
}

func (s *Store) GetLinkedServices(projectID string) ([]model.Service, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.name, s.type, s.image, s.status, s.container_id, s.config, s.created_at
		FROM services s
		INNER JOIN service_links sl ON s.id = sl.service_id
		WHERE sl.project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []model.Service
	for rows.Next() {
		var svc model.Service
		if err := rows.Scan(&svc.ID, &svc.Name, &svc.Type, &svc.Image, &svc.Status, &svc.ContainerID, &svc.Config, &svc.CreatedAt); err != nil {
			return nil, err
		}
		services = append(services, svc)
	}
	return services, nil
}

// --- Domains ---

func (s *Store) CreateDomain(projectID string, input model.CreateDomainInput) (*model.Domain, error) {
	sslAuto := true
	if input.SSLAuto != nil {
		sslAuto = *input.SSLAuto
	}
	d := &model.Domain{
		ID:        newID(),
		ProjectID: projectID,
		Domain:    input.Domain,
		SSLAuto:   sslAuto,
		CreatedAt: time.Now(),
	}
	_, err := s.db.Exec(`
		INSERT INTO domains (id, project_id, domain, ssl_auto, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		d.ID, d.ProjectID, d.Domain, d.SSLAuto, d.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Store) GetDomains(projectID string) ([]model.Domain, error) {
	rows, err := s.db.Query(`
		SELECT id, project_id, domain, ssl_auto, created_at
		FROM domains WHERE project_id = ? ORDER BY created_at ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []model.Domain
	for rows.Next() {
		var d model.Domain
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Domain, &d.SSLAuto, &d.CreatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, nil
}

func (s *Store) DeleteDomain(id string) error {
	_, err := s.db.Exec(`DELETE FROM domains WHERE id=?`, id)
	return err
}

// --- Users ---

func (s *Store) CreateUser(username, passwordHash, role string) (*model.User, error) {
	if role == "" {
		role = "admin"
	}
	u := &model.User{
		ID:           newID(),
		Username:     username,
		PasswordHash: passwordHash,
		Role:         role,
		CreatedAt:    time.Now(),
	}
	_, err := s.db.Exec(`
		INSERT INTO users (id, username, password_hash, role, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		u.ID, u.Username, u.PasswordHash, u.Role, u.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) GetUserByUsername(username string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRow(`
		SELECT id, username, password_hash, role, created_at
		FROM users WHERE username = ?`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (s *Store) GetUser(id string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRow(`
		SELECT id, username, password_hash, role, created_at
		FROM users WHERE id = ?`, id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (s *Store) GetUserCount() (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

// --- Sessions ---

func (s *Store) CreateSession(userID, token string) (*model.Session, error) {
	sess := &model.Session{
		ID:        newID(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}
	_, err := s.db.Exec(`
		INSERT INTO sessions (id, user_id, token, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		sess.ID, sess.UserID, sess.Token, sess.ExpiresAt, sess.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *Store) GetSessionByToken(token string) (*model.Session, error) {
	sess := &model.Session{}
	err := s.db.QueryRow(`
		SELECT id, user_id, token, expires_at, created_at
		FROM sessions WHERE token = ?`, token,
	).Scan(&sess.ID, &sess.UserID, &sess.Token, &sess.ExpiresAt, &sess.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return sess, err
}

func (s *Store) DeleteSessionByToken(token string) error {
	_, err := s.db.Exec(`DELETE FROM sessions WHERE token=?`, token)
	return err
}

// --- Backups ---

func (s *Store) CreateBackup(b *model.Backup) error {
	_, err := s.db.Exec(`
		INSERT INTO backups (id, type, service_id, filename, size, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		b.ID, b.Type, b.ServiceID, b.Filename, b.Size, b.CreatedAt,
	)
	return err
}

func (s *Store) ListBackups() ([]model.Backup, error) {
	rows, err := s.db.Query(`
		SELECT id, type, service_id, filename, size, created_at
		FROM backups ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []model.Backup
	for rows.Next() {
		var b model.Backup
		if err := rows.Scan(&b.ID, &b.Type, &b.ServiceID, &b.Filename, &b.Size, &b.CreatedAt); err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, nil
}

func (s *Store) GetBackup(id string) (*model.Backup, error) {
	b := &model.Backup{}
	err := s.db.QueryRow(`
		SELECT id, type, service_id, filename, size, created_at
		FROM backups WHERE id = ?`, id,
	).Scan(&b.ID, &b.Type, &b.ServiceID, &b.Filename, &b.Size, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return b, err
}

func (s *Store) DeleteBackup(id string) error {
	_, err := s.db.Exec(`DELETE FROM backups WHERE id=?`, id)
	return err
}

// GetDB exposes the underlying *sql.DB for backup operations.
func (s *Store) GetDB() *sql.DB {
	return s.db
}

// --- Audit Logs ---

func (s *Store) AddAuditLog(userID, username, action, resource, resourceID, details string) error {
	_, err := s.db.Exec(`
		INSERT INTO audit_logs (id, user_id, username, action, resource, resource_id, details, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		newID(), userID, username, action, resource, resourceID, details, time.Now(),
	)
	return err
}

func (s *Store) ListAuditLogs(limit int) ([]model.AuditLog, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(`
		SELECT id, user_id, username, action, resource, resource_id, details, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Username, &l.Action, &l.Resource, &l.ResourceID, &l.Details, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// --- Invitations ---

func (s *Store) CreateInvitation(inv *model.Invitation) error {
	_, err := s.db.Exec(`
		INSERT INTO invitations (id, email, role, token, used, created_by, created_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		inv.ID, inv.Email, inv.Role, inv.Token, inv.Used, inv.CreatedBy, inv.CreatedAt, inv.ExpiresAt,
	)
	return err
}

func (s *Store) GetInvitationByToken(token string) (*model.Invitation, error) {
	inv := &model.Invitation{}
	err := s.db.QueryRow(`
		SELECT id, email, role, token, used, created_by, created_at, expires_at
		FROM invitations WHERE token = ?`, token,
	).Scan(&inv.ID, &inv.Email, &inv.Role, &inv.Token, &inv.Used, &inv.CreatedBy, &inv.CreatedAt, &inv.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return inv, err
}

func (s *Store) MarkInvitationUsed(id string) error {
	_, err := s.db.Exec(`UPDATE invitations SET used=1 WHERE id=?`, id)
	return err
}

func (s *Store) ListInvitations() ([]model.Invitation, error) {
	rows, err := s.db.Query(`
		SELECT id, email, role, token, used, created_by, created_at, expires_at
		FROM invitations ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invs []model.Invitation
	for rows.Next() {
		var inv model.Invitation
		if err := rows.Scan(&inv.ID, &inv.Email, &inv.Role, &inv.Token, &inv.Used, &inv.CreatedBy, &inv.CreatedAt, &inv.ExpiresAt); err != nil {
			return nil, err
		}
		invs = append(invs, inv)
	}
	return invs, nil
}

// --- Users extended ---

func (s *Store) ListUsers() ([]model.User, error) {
	rows, err := s.db.Query(`
		SELECT id, username, password_hash, role, created_at
		FROM users ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *Store) UpdateUserRole(id, role string) error {
	_, err := s.db.Exec(`UPDATE users SET role=? WHERE id=?`, role, id)
	return err
}

func (s *Store) DeleteUser(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	tx.Exec(`DELETE FROM sessions WHERE user_id=?`, id)
	tx.Exec(`DELETE FROM users WHERE id=?`, id)
	return tx.Commit()
}

// --- Volumes ---

func (s *Store) CreateVolume(projectID string, input model.CreateVolumeInput) (*model.Volume, error) {
	v := &model.Volume{
		ID:        newID(),
		Name:      input.Name,
		MountPath: input.MountPath,
		ProjectID: projectID,
		CreatedAt: time.Now(),
	}
	_, err := s.db.Exec(`
		INSERT INTO volumes (id, name, mount_path, project_id, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		v.ID, v.Name, v.MountPath, v.ProjectID, v.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (s *Store) ListVolumes(projectID string) ([]model.Volume, error) {
	rows, err := s.db.Query(`
		SELECT id, name, mount_path, project_id, created_at
		FROM volumes WHERE project_id = ? ORDER BY created_at ASC`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var volumes []model.Volume
	for rows.Next() {
		var v model.Volume
		if err := rows.Scan(&v.ID, &v.Name, &v.MountPath, &v.ProjectID, &v.CreatedAt); err != nil {
			return nil, err
		}
		volumes = append(volumes, v)
	}
	return volumes, nil
}

func (s *Store) DeleteVolume(id string) error {
	_, err := s.db.Exec(`DELETE FROM volumes WHERE id=?`, id)
	return err
}
