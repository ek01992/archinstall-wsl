package user

import (
	"fmt"
	"path/filepath"
	"strings"

	"archwsl-tui-configurator/internal/logx"
	"archwsl-tui-configurator/internal/tx"
)

func (s *Service) CreateUser(username, password string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username must not be empty")
	}
	if strings.ContainsAny(username, ":\n") || strings.Contains(password, "\n") {
		return fmt.Errorf("invalid credentials")
	}
	logx.Info("create user: begin", "user", username)

	if !s.id.UserExists(username) {
		if err := s.cmd.Run("useradd", "-m", username); err != nil {
			logx.Error("create user: useradd failed", "user", username, "err", err)
			return fmt.Errorf("useradd failed: %w", err)
		}
	}
	if err := s.cmd.RunWithStdin("chpasswd", fmt.Sprintf("%s:%s", username, password)); err != nil {
		logx.Error("create user: chpasswd failed", "user", username, "err", err)
		return fmt.Errorf("setting password failed: %w", err)
	}
	if err := s.cmd.Run("usermod", "-aG", "wheel", username); err != nil {
		logx.Warn("create user: usermod failed, trying gpasswd", "user", username, "err", err)
		if err2 := s.cmd.Run("gpasswd", "-a", username, "wheel"); err2 != nil {
			logx.Error("create user: add to wheel failed", "user", username, "usermodErr", err, "gpasswdErr", err2)
			return fmt.Errorf("adding to wheel failed (usermod: %v, gpasswd: %w)", err, err2)
		}
	}

	desired := "%wheel ALL=(ALL) NOPASSWD: ALL\n"
	dir := strings.TrimSpace(s.sudoersDir)
	if dir == "" { dir = "/etc/sudoers.d" }
	_ = s.fs.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, "010_wheel_nopasswd")
	current, _ := s.fs.ReadFile(path)
	if string(current) != desired {
		if err := s.sudo.Validate(desired); err != nil {
			logx.Error("create user: sudoers validate failed", "user", username, "err", err)
			return fmt.Errorf("sudoers validation failed: %w", err)
		}
		if err := s.fs.WriteFile(path, []byte(desired), 0o440); err != nil {
			logx.Error("create user: write sudoers failed", "user", username, "err", err)
			return fmt.Errorf("writing sudoers file failed: %w", err)
		}
	}
	return nil
}

func (s *Service) CreateUserTx(username, password string) (err error) {
	tr := tx.New()
	defer func() { if err != nil { _ = tr.Rollback() } }()

	userExisted := s.id.UserExists(username)
	if !userExisted {
		u := username
		tr.Defer(func() error { return s.cmd.Run("userdel", "-r", u) })
	}
	dir := strings.TrimSpace(s.sudoersDir)
	if dir == "" { dir = "/etc/sudoers.d" }
	sudoersFile := filepath.Join(dir, "010_wheel_nopasswd")
	if prev, perr := s.fs.ReadFile(sudoersFile); perr == nil {
		path := sudoersFile
		data := append([]byte(nil), prev...)
		tr.Defer(func() error { return s.fs.WriteFile(path, data, 0o440) })
	} else {
		path := sudoersFile
		tr.Defer(func() error { return s.cmd.Run("rm", "-f", path) })
	}

	if err = s.CreateUser(username, password); err != nil {
		return err
	}
	tr.Commit()
	return nil
}
