package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"shop_srvs/user_srv/proto"
)

var userClient proto.UserClient
var conn *grpc.ClientConn

func init() {
	// Here to get grpc userClient
	var (
		err error
	)
	conn, err = grpc.Dial(":8088", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	userClient = proto.NewUserClient(conn)
}

func TestGetUserList() {
	userListRsp, err := userClient.GetUserList(
		context.Background(),
		&proto.PageInfo{
			Pn:    1,
			PSize: 2,
		})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Total %d users found\n", userListRsp.Total)
	for _, user := range userListRsp.Data {
		log.Println(user)
		check, err := userClient.CheckPassWord(context.Background(), &proto.PasswordCheckInfo{
			Password:          "HardCodePassWord",
			EncryptedPassword: user.Password,
		})
		if err != nil {
			log.Fatalln(err)
		}
		if !check.Success {
			log.Printf("Password is error for user: %s\n", user.NickName)
		}
	}
}

func main() {
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(conn)
	TestGetUserList()
}
