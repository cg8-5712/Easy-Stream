package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"easy-stream/internal/config"
)

var (
	initConfig = flag.Bool("init-config", false, "Generate default config.yaml file")
	verify     = flag.Bool("verify", false, "Verify config.yaml file")
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

	return false
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
	fmt.Println("  --help           Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  easy-stream --init-config    # Create config.yaml with default settings")
	fmt.Println("  easy-stream --verify         # Verify config and test service connections")
	fmt.Println("  easy-stream                  # Start the server")
	fmt.Println()
	fmt.Println("GitHub:")
	fmt.Println("  https://github.com/cg8-5712/Easy-Stream")
	fmt.Println()
	fmt.Println("Documentation:")
	fmt.Println("  For more information, please visit the GitHub repository")
}
