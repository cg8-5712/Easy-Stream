package repository

import (
	"database/sql"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// SeedData 在 debug 模式下插入测试数据
func SeedData(db *sql.DB) error {
	log.Println("Debug模式: 开始插入种子数据...")

	// 插入测试用户
	if err := seedUsers(db); err != nil {
		return err
	}

	// 插入测试直播流
	if err := seedStreams(db); err != nil {
		return err
	}

	log.Println("Debug模式: 种子数据插入完成")
	return nil
}

// seedUsers 插入测试用户
func seedUsers(db *sql.DB) error {
	users := []struct {
		username string
		password string
		realName string
		email    string
	}{
		{"admin", "admin123", "管理员", "admin@example.com"},
		{"operator", "operator123", "操作员", "operator@example.com"},
		{"viewer", "viewer123", "观众", "viewer@example.com"},
	}

	for _, u := range users {
		// 检查用户是否已存在
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", u.username).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		// 生成密码哈希
		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		// 插入用户
		_, err = db.Exec(`
			INSERT INTO users (username, password_hash, real_name, email, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $5)
		`, u.username, string(hash), u.realName, u.email, time.Now())
		if err != nil {
			return err
		}
		log.Printf("Debug模式: 创建用户 %s (密码: %s)", u.username, u.password)
	}

	return nil
}

// seedStreams 插入测试直播流
func seedStreams(db *sql.DB) error {
	// 获取 admin 用户 ID
	var adminID int64
	err := db.QueryRow("SELECT id FROM users WHERE username = 'admin'").Scan(&adminID)
	if err != nil {
		return err
	}

	streams := []struct {
		streamKey   string
		name        string
		description string
		visibility  string
		streamer    string
	}{
		{"test-stream-001", "测试直播间1", "这是一个公开的测试直播间", "public", "测试主播A"},
		{"test-stream-002", "测试直播间2", "这是一个私密的测试直播间", "private", "测试主播B"},
	}

	now := time.Now()
	scheduledStart := now.Add(-1 * time.Hour)
	scheduledEnd := now.Add(24 * time.Hour)

	for _, s := range streams {
		// 检查是否已存在
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM streams WHERE stream_key = $1)", s.streamKey).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		// 插入直播流
		_, err = db.Exec(`
			INSERT INTO streams (stream_key, name, description, visibility, status,
				streamer_name, scheduled_start_time, scheduled_end_time, created_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'idle', $5, $6, $7, $8, $9, $9)
		`, s.streamKey, s.name, s.description, s.visibility, s.streamer, scheduledStart, scheduledEnd, adminID, now)
		if err != nil {
			return err
		}
		log.Printf("Debug模式: 创建直播流 %s (%s)", s.name, s.streamKey)
	}

	return nil
}
