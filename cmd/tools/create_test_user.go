//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	phone := "17673796081"
	password := "hxf260011"

	// 生成密码哈希
	pwdHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("生成密码哈希失败:", err)
	}

	// 从环境变量获取数据库连接
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3307)/go_micro?parseTime=true&charset=utf8mb4&loc=Local"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	userNo := "U" + fmt.Sprintf("%d", time.Now().Unix())
	username := phone // 使用手机号作为用户名
	now := time.Now()

	// 插入用户 - 使用V2表结构
	query := `INSERT INTO users (user_no, username, password_hash, phone, status, register_source, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = db.Exec(query, userNo, username, string(pwdHash), phone, 1, "phone", now, now)
	if err != nil {
		log.Fatal("插入用户失败:", err)
	}

	fmt.Println("✓ 测试用户创建成功!")
	fmt.Printf("  用户编号: %s\n", userNo)
	fmt.Printf("  用户名(登录用): %s\n", username)
	fmt.Printf("  手机号: %s\n", phone)
	fmt.Printf("  密码: %s\n", password)
	fmt.Printf("  状态: 正常(1)\n")
	fmt.Printf("  注册来源: phone\n")
}
