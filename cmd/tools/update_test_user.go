package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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

	// 查询现有用户
	var userNo, username string
	err = db.QueryRow("SELECT user_no, username FROM users WHERE phone = ?", phone).Scan(&userNo, &username)
	if err != nil {
		log.Fatal("查询用户失败:", err)
	}

	// 更新密码
	_, err = db.Exec("UPDATE users SET password_hash = ?, status = 1 WHERE phone = ?", string(pwdHash), phone)
	if err != nil {
		log.Fatal("更新密码失败:", err)
	}

	fmt.Println("✓ 测试用户密码已更新!")
	fmt.Printf("  用户编号: %s\n", userNo)
	fmt.Printf("  用户名(登录用): %s\n", username)
	fmt.Printf("  手机号: %s\n", phone)
	fmt.Printf("  新密码: %s\n", password)
	fmt.Printf("  状态: 已激活(1)\n")
}
