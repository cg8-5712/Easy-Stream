package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"syscall"

	"easy-stream/internal/config"
	"easy-stream/internal/model"
	"easy-stream/internal/repository"
	"easy-stream/pkg/logger"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

var (
	initConfig = flag.Bool("init-config", false, "Generate default config.yaml file")
	verify     = flag.Bool("verify", false, "Verify config.yaml file")
	setAdmin   = flag.Bool("set-admin", false, "Set admin username and password")
	help       = flag.Bool("help", false, "Show help message")
)

// ParseFlags parses command line flags and returns true if the program should exit
func ParseFlags() bool {
	flag.Parse()

	// 显示帮助信息
	if *help {
		showHelp()
		return true
	}

	// 初始化配置文件
	if *initConfig {
		if err := config.InitConfig(); err != nil {
			log.Fatalf("Failed to initialize config: %v", err)
		}
		fmt.Println("✓ Config file created: config.yaml")
		fmt.Println("Please edit config.yaml and restart the server")
		return true
	}

	// 验证配置文件
	if *verify {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("✗ Config validation failed: %v", err)
		}
		if err := config.Validate(cfg); err != nil {
			log.Fatalf("✗ Config validation failed: %v", err)
		}
		fmt.Println("✓ Config file is valid")
		fmt.Println()

		// 测试服务连接
		fmt.Println("Testing service connections...")
		fmt.Println()

		statuses := config.VerifyConnections(cfg)
		allConnected := true

		for _, status := range statuses {
			fmt.Printf("%-15s %s", status.Name+":", status.Status)
			if status.Details != "" {
				fmt.Printf(" (%s)", status.Details)
			}
			fmt.Println()
			if status.Error != "" {
				fmt.Printf("  Error: %s\n", status.Error)
				allConnected = false
			}
		}

		fmt.Println()
		if allConnected {
			fmt.Println("✓ All services are connected")
		} else {
			fmt.Println("✗ Some services failed to connect")
			os.Exit(1)
		}

		return true
	}

	// 设置管理员账号密码
	if *setAdmin {
		if err := setAdminPassword(); err != nil {
			log.Fatalf("✗ Failed to set admin password: %v", err)
		}
		return true
	}

	return false
}

// setAdminPassword 设置管理员账号密码
func setAdminPassword() error {
	fmt.Println("Easy-Stream - Set Admin Password")
	fmt.Println()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 初始化日志（NewPostgresDB 会调用 AutoMigrate -> logger.Info）
	logger.Init(cfg.Log.Level)

	// 连接数据库
	db, err := repository.NewPostgresDB(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 读取用户名
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter admin username (default: admin): ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		username = "admin"
	}

	// 读取密码（隐藏输入）
	fmt.Print("Enter admin password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()
	password := string(passwordBytes)

	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters")
	}

	// 确认密码
	fmt.Print("Confirm admin password: ")
	confirmBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println()
	confirm := string(confirmBytes)

	if password != confirm {
		return fmt.Errorf("passwords do not match")
	}

	// 读取真实姓名
	fmt.Print("Enter real name: ")
	realName, _ := reader.ReadString('\n')
	realName = strings.TrimSpace(realName)
	if realName == "" {
		return fmt.Errorf("real name is required")
	}

	// 读取邮箱
	fmt.Print("Enter email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// 生成密码哈希
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 检查用户是否已存在
	var existingUser model.User
	result := db.Where("username = ?", username).First(&existingUser)

	if result.Error == nil {
		// 用户已存在，更新密码、真实姓名和邮箱
		updates := map[string]interface{}{
			"password_hash": string(hash),
			"real_name":     realName,
			"email":         email,
		}
		if err := db.Model(&existingUser).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to update admin user: %w", err)
		}
		fmt.Printf("✓ Admin user updated successfully: %s\n", username)
	} else {
		// 用户不存在，创建新用户
		user := &model.User{
			Username:     username,
			PasswordHash: string(hash),
			RealName:     &realName,
			Email:        &email,
		}

		if err := db.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		fmt.Printf("✓ Admin user created successfully: %s\n", username)
	}

	return nil
}

func showHelp() {
	fmt.Println("Easy-Stream - 流媒体管理系统")
	fmt.Println("A lightweight streaming media management system based on ZLMediaKit")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  easy-stream [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --init-config    Generate default config.yaml file with example configuration")
	fmt.Println("  --verify         Verify config.yaml file and test connections to all services")
	fmt.Println("                   (PostgreSQL, Redis, ZLMediaKit)")
	fmt.Println("  --set-admin      Set admin username and password interactively")
	fmt.Println("  --help           Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  easy-stream --init-config    # Create config.yaml with default settings")
	fmt.Println("  easy-stream --verify         # Verify config and test service connections")
	fmt.Println("  easy-stream --set-admin      # Set admin username and password")
	fmt.Println("  easy-stream                  # Start the server")
	fmt.Println()
	fmt.Println("GitHub:")
	fmt.Println("  https://github.com/cg8-5712/Easy-Stream")
	fmt.Println()
	fmt.Println("Documentation:")
	fmt.Println("  For more information, please visit the GitHub repository")
}
