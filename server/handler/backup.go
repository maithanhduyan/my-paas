package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/my-paas/server/model"
	"github.com/my-paas/server/store"
)

const backupDir = "/data/backups"

func init() {
	os.MkdirAll(backupDir, 0o755)
}

func (h *Handler) ListBackups(c *fiber.Ctx) error {
	backups, err := h.Store.ListBackups()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if backups == nil {
		backups = []model.Backup{}
	}
	return c.JSON(backups)
}

func (h *Handler) CreateBackup(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	username, _ := c.Locals("username").(string)

	filename := fmt.Sprintf("system-%s.db", time.Now().Format("20060102-150405"))
	destPath := filepath.Join(backupDir, filename)

	// Use SQLite backup API via raw copy of the WAL-checkpointed database
	if err := backupSQLite(h.Store, destPath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "backup failed: " + err.Error()})
	}

	info, _ := os.Stat(destPath)
	var size int64
	if info != nil {
		size = info.Size()
	}

	b := &model.Backup{
		ID:        store.NewID(),
		Type:      "system",
		Filename:  filename,
		Size:      size,
		CreatedAt: time.Now(),
	}
	if err := h.Store.CreateBackup(b); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.Store.AddAuditLog(userID, username, "backup", "system", b.ID, "system backup created: "+filename)
	return c.Status(201).JSON(b)
}

func (h *Handler) DownloadBackup(c *fiber.Ctx) error {
	id := c.Params("id")
	b, err := h.Store.GetBackup(id)
	if err != nil || b == nil {
		return c.Status(404).JSON(fiber.Map{"error": "backup not found"})
	}

	path := filepath.Join(backupDir, b.Filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{"error": "backup file not found on disk"})
	}

	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, b.Filename))
	return c.SendFile(path)
}

func (h *Handler) DeleteBackup(c *fiber.Ctx) error {
	id := c.Params("id")
	userID, _ := c.Locals("user_id").(string)
	username, _ := c.Locals("username").(string)

	b, err := h.Store.GetBackup(id)
	if err != nil || b == nil {
		return c.Status(404).JSON(fiber.Map{"error": "backup not found"})
	}

	// Remove file from disk
	os.Remove(filepath.Join(backupDir, b.Filename))

	if err := h.Store.DeleteBackup(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	h.Store.AddAuditLog(userID, username, "delete", "backup", id, "deleted backup: "+b.Filename)
	return c.JSON(fiber.Map{"message": "backup deleted"})
}

func (h *Handler) RestoreBackup(c *fiber.Ctx) error {
	id := c.Params("id")

	b, err := h.Store.GetBackup(id)
	if err != nil || b == nil {
		return c.Status(404).JSON(fiber.Map{"error": "backup not found"})
	}

	if b.Type != "system" {
		return c.Status(400).JSON(fiber.Map{"error": "only system backups can be restored"})
	}

	srcPath := filepath.Join(backupDir, b.Filename)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{"error": "backup file not found on disk"})
	}

	// Copy backup file to the database location
	dbPath := os.Getenv("MYPAAS_DB")
	if dbPath == "" {
		dbPath = "/data/mypaas.db"
	}

	// Create a safety backup of current DB before restore
	safetyPath := dbPath + ".pre-restore"
	copyFile(dbPath, safetyPath)

	// Note: actual restore requires restart - we write a flag file
	restorePath := dbPath + ".restore"
	if err := copyFile(srcPath, restorePath); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to prepare restore: " + err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "restore prepared. Restart the server to apply.",
		"note":    "Current database backed up to " + safetyPath,
	})
}

// backupSQLite creates a consistent backup of the SQLite database.
func backupSQLite(s *store.Store, destPath string) error {
	db := s.GetDB()
	// Force WAL checkpoint to get a consistent snapshot
	_, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	if err != nil {
		return fmt.Errorf("wal checkpoint: %w", err)
	}

	// Get the DB path
	var dbPath string
	row := db.QueryRow("PRAGMA database_list")
	var seq int
	var name string
	if err := row.Scan(&seq, &name, &dbPath); err != nil {
		return fmt.Errorf("get db path: %w", err)
	}

	return copyFile(dbPath, destPath)
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
