package deployment

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrTargetNotFound = errors.New("vps target not found")
	ErrEnvVarNotFound = errors.New("env var not found")
	ErrRunNotFound    = errors.New("deployment run not found")
)

type Service struct {
	db  *gorm.DB
	key []byte
}

func NewService(db *gorm.DB, encryptionKeyHex string) (*Service, error) {
	key, err := hex.DecodeString(encryptionKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key: must be a hex-encoded string: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid encryption key: decoded length must be 32 bytes, got %d", len(key))
	}
	return &Service{db: db, key: key}, nil
}

// ─── Env Vars ────────────────────────────────────────────────────────────────

type EnvVarInput struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

// EnvVarResult mirrors ProjectEnvVar but with the decrypted value for the owner.
type EnvVarResult struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"` // plaintext (hidden on client for secrets)
	IsSecret  bool      `json:"is_secret"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Service) SetEnvVar(projectID uuid.UUID, input *EnvVarInput) (*EnvVarResult, error) {
	encrypted, err := s.encrypt(input.Value)
	if err != nil {
		return nil, fmt.Errorf("encrypt: %w", err)
	}

	ev := &models.ProjectEnvVar{
		ProjectID: projectID,
		Key:       input.Key,
		Value:     encrypted,
		IsSecret:  input.IsSecret,
	}

	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "is_secret", "updated_at"}),
	}).Create(ev).Error; err != nil {
		return nil, fmt.Errorf("upsert env var: %w", err)
	}

	return s.toResult(ev, input.Value), nil
}

func (s *Service) ListEnvVars(projectID uuid.UUID) ([]*EnvVarResult, error) {
	var rows []models.ProjectEnvVar
	if err := s.db.Where("project_id = ?", projectID).Order("key ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list env vars: %w", err)
	}

	results := make([]*EnvVarResult, 0, len(rows))
	for i := range rows {
		plain, err := s.decrypt(rows[i].Value)
		if err != nil {
			plain = "***"
		}
		results = append(results, s.toResult(&rows[i], plain))
	}
	return results, nil
}

func (s *Service) DeleteEnvVar(projectID, id uuid.UUID) error {
	res := s.db.Where("id = ? AND project_id = ?", id, projectID).Delete(&models.ProjectEnvVar{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrEnvVarNotFound
	}
	return nil
}

func (s *Service) toResult(ev *models.ProjectEnvVar, plain string) *EnvVarResult {
	return &EnvVarResult{
		ID:        ev.ID,
		ProjectID: ev.ProjectID,
		Key:       ev.Key,
		Value:     plain,
		IsSecret:  ev.IsSecret,
		CreatedAt: ev.CreatedAt,
		UpdatedAt: ev.UpdatedAt,
	}
}

// ─── VPS Targets ─────────────────────────────────────────────────────────────

type TargetInput struct {
	Name          string `json:"name"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	AuthType      string `json:"auth_type"` // "key" | "password"
	SSHKey        string `json:"ssh_key"`   // PEM private key (plaintext from client)
	SSHPassword   string `json:"ssh_password"`
	DeployPath    string `json:"deploy_path"`
	PreDeployCmd  string `json:"pre_deploy_cmd"`
	DeployCmd     string `json:"deploy_cmd"`
	PostDeployCmd string `json:"post_deploy_cmd"`
}

func (s *Service) CreateTarget(projectID uuid.UUID, input *TargetInput) (*models.VPSTarget, error) {
	if input.Port == 0 {
		input.Port = 22
	}

	target := &models.VPSTarget{
		ProjectID:     projectID,
		Name:          input.Name,
		Host:          input.Host,
		Port:          input.Port,
		Username:      input.Username,
		AuthType:      input.AuthType,
		DeployPath:    input.DeployPath,
		PreDeployCmd:  input.PreDeployCmd,
		DeployCmd:     input.DeployCmd,
		PostDeployCmd: input.PostDeployCmd,
		Status:        "idle",
	}

	if input.SSHKey != "" {
		enc, err := s.encrypt(input.SSHKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt ssh key: %w", err)
		}
		target.SSHKey = enc
	}
	if input.SSHPassword != "" {
		enc, err := s.encrypt(input.SSHPassword)
		if err != nil {
			return nil, fmt.Errorf("encrypt ssh password: %w", err)
		}
		target.SSHPassword = enc
	}

	if err := s.db.Create(target).Error; err != nil {
		return nil, fmt.Errorf("create target: %w", err)
	}
	return target, nil
}

func (s *Service) ListTargets(projectID uuid.UUID) ([]models.VPSTarget, error) {
	var targets []models.VPSTarget
	if err := s.db.Where("project_id = ?", projectID).Order("created_at ASC").Find(&targets).Error; err != nil {
		return nil, fmt.Errorf("list targets: %w", err)
	}
	return targets, nil
}

func (s *Service) GetTarget(projectID, targetID uuid.UUID) (*models.VPSTarget, error) {
	var target models.VPSTarget
	if err := s.db.Where("id = ? AND project_id = ?", targetID, projectID).First(&target).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTargetNotFound
		}
		return nil, err
	}
	return &target, nil
}

func (s *Service) UpdateTarget(projectID, targetID uuid.UUID, input *TargetInput) (*models.VPSTarget, error) {
	target, err := s.GetTarget(projectID, targetID)
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"name":           input.Name,
		"host":           input.Host,
		"port":           input.Port,
		"username":       input.Username,
		"auth_type":      input.AuthType,
		"deploy_path":    input.DeployPath,
		"pre_deploy_cmd": input.PreDeployCmd,
		"deploy_cmd":     input.DeployCmd,
		"post_deploy_cmd": input.PostDeployCmd,
	}
	if input.Port == 0 {
		updates["port"] = 22
	}
	if input.SSHKey != "" {
		enc, err := s.encrypt(input.SSHKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt ssh key: %w", err)
		}
		updates["ssh_key"] = enc
	}
	if input.SSHPassword != "" {
		enc, err := s.encrypt(input.SSHPassword)
		if err != nil {
			return nil, fmt.Errorf("encrypt ssh password: %w", err)
		}
		updates["ssh_password"] = enc
	}

	if err := s.db.Model(target).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update target: %w", err)
	}
	return s.GetTarget(projectID, targetID)
}

func (s *Service) DeleteTarget(projectID, targetID uuid.UUID) error {
	res := s.db.Where("id = ? AND project_id = ?", targetID, projectID).Delete(&models.VPSTarget{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrTargetNotFound
	}
	return nil
}

// ─── Deployment Runs ─────────────────────────────────────────────────────────

func (s *Service) ListRuns(projectID uuid.UUID, targetID *uuid.UUID, limit int) ([]models.DeploymentRun, error) {
	q := s.db.Where("project_id = ?", projectID)
	if targetID != nil {
		q = q.Where("vps_target_id = ?", *targetID)
	}
	var runs []models.DeploymentRun
	if err := q.Order("created_at DESC").Limit(limit).Find(&runs).Error; err != nil {
		return nil, err
	}
	return runs, nil
}

func (s *Service) GetRun(runID uuid.UUID) (*models.DeploymentRun, error) {
	var run models.DeploymentRun
	if err := s.db.Preload("VPSTarget").First(&run, "id = ?", runID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRunNotFound
		}
		return nil, err
	}
	return &run, nil
}

// ─── Deploy ──────────────────────────────────────────────────────────────────

// Deploy SSHes into the VPS target and runs the configured commands.
// Runs asynchronously; the returned DeploymentRun reflects initial state.
func (s *Service) Deploy(projectID, targetID uuid.UUID, triggeredBy *uuid.UUID) (*models.DeploymentRun, error) {
	target, err := s.GetTarget(projectID, targetID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	run := &models.DeploymentRun{
		VPSTargetID: targetID,
		ProjectID:   projectID,
		Status:      "running",
		TriggeredBy: triggeredBy,
		StartedAt:   now,
	}
	if err := s.db.Create(run).Error; err != nil {
		return nil, fmt.Errorf("create run: %w", err)
	}

	// Mark target as deploying
	s.db.Model(target).Updates(map[string]interface{}{"status": "deploying"})

	go s.runDeploy(run, target, projectID)
	return run, nil
}

func (s *Service) runDeploy(run *models.DeploymentRun, target *models.VPSTarget, projectID uuid.UUID) {
	var logBuf bytes.Buffer
	writeLog := func(msg string) {
		logBuf.WriteString(msg + "\n")
	}

	finish := func(status string) {
		now := time.Now()
		s.db.Model(run).Updates(map[string]interface{}{
			"status":      status,
			"log_output":  logBuf.String(),
			"finished_at": &now,
		})
		s.db.Model(target).Updates(map[string]interface{}{
			"status":           status,
			"last_deployed_at": &now,
		})
	}

	writeLog(fmt.Sprintf("[%s] Connecting to %s@%s:%d ...", time.Now().Format(time.RFC3339), target.Username, target.Host, target.Port))

	client, err := s.sshConnect(target)
	if err != nil {
		writeLog("[ERROR] SSH connection failed: " + err.Error())
		finish("failed")
		return
	}
	defer client.Close()
	writeLog("[OK] Connected")

	// Fetch env vars for the project and build export block
	envVars, _ := s.ListEnvVars(projectID)
	var envExport strings.Builder
	for _, ev := range envVars {
		// Single-quote the value and escape any single quotes in it
		safe := strings.ReplaceAll(ev.Value, "'", "'\\''")
		envExport.WriteString(fmt.Sprintf("export %s='%s'\n", ev.Key, safe))
	}

	// Compose the full remote script
	var script strings.Builder
	script.WriteString("set -e\n")
	if envExport.Len() > 0 {
		script.WriteString(envExport.String())
	}
	script.WriteString(fmt.Sprintf("cd %s\n", target.DeployPath))
	if target.PreDeployCmd != "" {
		script.WriteString(target.PreDeployCmd + "\n")
	}
	script.WriteString(target.DeployCmd + "\n")
	if target.PostDeployCmd != "" {
		script.WriteString(target.PostDeployCmd + "\n")
	}

	writeLog(fmt.Sprintf("[%s] Running deployment script ...", time.Now().Format(time.RFC3339)))
	out, err := s.runScript(client, script.String())
	writeLog(out)
	if err != nil {
		writeLog("[ERROR] Deployment failed: " + err.Error())
		finish("failed")
		return
	}

	writeLog(fmt.Sprintf("[%s] Deployment successful", time.Now().Format(time.RFC3339)))
	finish("success")
}

func (s *Service) sshConnect(target *models.VPSTarget) (*ssh.Client, error) {
	var authMethods []ssh.AuthMethod

	if target.AuthType == "key" && target.SSHKey != "" {
		pemBytes, err := s.decrypt(target.SSHKey)
		if err != nil {
			return nil, fmt.Errorf("decrypt ssh key: %w", err)
		}
		signer, err := ssh.ParsePrivateKey([]byte(pemBytes))
		if err != nil {
			return nil, fmt.Errorf("parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	} else if target.AuthType == "password" && target.SSHPassword != "" {
		pass, err := s.decrypt(target.SSHPassword)
		if err != nil {
			return nil, fmt.Errorf("decrypt ssh password: %w", err)
		}
		authMethods = append(authMethods, ssh.Password(pass))
	}

	cfg := &ssh.ClientConfig{
		User: target.Username,
		Auth: authMethods,
		// Accept any host key — in production this should be pinned
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Timeout: 30 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", target.Host, target.Port)
	return ssh.Dial("tcp", addr, cfg)
}

func (s *Service) runScript(client *ssh.Client, script string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var out bytes.Buffer
	session.Stdout = &out
	session.Stderr = &out
	session.Stdin = strings.NewReader(script)

	err = session.Run("bash -s")
	return out.String(), err
}

// RunSSHCommand satisfies the monitoring.SSHRunner interface.
// It connects to the target and runs a one-shot bash command.
func (s *Service) RunSSHCommand(targetID uuid.UUID, cmd string) (string, error) {
	// We need project_id to call GetTarget, but monitoring only has targetID.
	// Query the target directly by ID.
	var target models.VPSTarget
	if err := s.db.First(&target, "id = ?", targetID).Error; err != nil {
		return "", fmt.Errorf("target not found: %w", err)
	}
	client, err := s.sshConnect(&target)
	if err != nil {
		return "", err
	}
	defer client.Close()
	return s.runScript(client, cmd)
}

// ─── AES-GCM helpers ─────────────────────────────────────────────────────────

func (s *Service) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ct), nil
}

func (s *Service) decrypt(ciphertextHex string) (string, error) {
	ct, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := aesGCM.NonceSize()
	if len(ct) < ns {
		return "", errors.New("ciphertext too short")
	}
	plain, err := aesGCM.Open(nil, ct[:ns], ct[ns:], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
