//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"ecomm-micro/proto/userpb"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	phone := "17673796081"
	password := "hxf260011"

	// 连接 user service gRPC
	conn, err := grpc.Dial("localhost:9083", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("连接 user service 失败:", err)
	}
	defer conn.Close()

	client := userpb.NewUserServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 测试通过手机号查询用户
	fmt.Println("=== 测试 GetUserByPhone ===")
	upb, err := client.GetUserByPhone(ctx, &userpb.GetUserByPhoneRequest{Phone: phone})
	if err != nil {
		log.Fatal("GetUserByPhone 失败:", err)
	}

	fmt.Printf("用户ID: %s\n", upb.UserId)
	fmt.Printf("用户名: %s\n", upb.Username)
	fmt.Printf("手机号: %s\n", upb.Phone)
	fmt.Printf("角色: %s\n", upb.Role)
	fmt.Printf("状态: %d\n", upb.Status)
	fmt.Printf("密码哈希: %s\n", upb.PasswordHash)
	fmt.Printf("密码哈希长度: %d\n", len(upb.PasswordHash))

	// 验证密码
	fmt.Println("\n=== 验证密码 ===")
	err = bcrypt.CompareHashAndPassword([]byte(upb.PasswordHash), []byte(password))
	if err == nil {
		fmt.Println("✓ 密码验证成功!")
	} else {
		fmt.Printf("✗ 密码验证失败: %v\n", err)
	}
}
