//go:build ignore

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

	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3307)/go_micro?parseTime=true&charset=utf8mb4&loc=Local"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	var userNo, username, passwordHash string
	err = db.QueryRow("SELECT user_no, username, password_hash FROM users WHERE phone = ?", phone).Scan(&userNo, &username, &passwordHash)
	if err != nil {
		log.Fatal("查询用户失败:", err)
	}

	fmt.Println("数据库中的用户信息:")
	fmt.Printf("  用户编号: %s\n", userNo)
	fmt.Printf("  用户名: %s\n", username)
	fmt.Printf("  密码哈希: %s\n", passwordHash)
	fmt.Printf("  密码哈希长度: %d\n", len(passwordHash))

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err == nil {
		fmt.Println("\n✓ 密码验证成功!")
	} else {
		fmt.Printf("\n✗ 密码验证失败: %v\n", err)

		// 重新生成正确的密码哈希
		newHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		fmt.Printf("\n正确的密码哈希应该是:\n%s\n", string(newHash))
	}
}
