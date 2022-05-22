package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"shop_srvs/user_srv/handler"
	"shop_srvs/user_srv/proto"
)

func main() {
	IP := flag.String("ip", "0.0.0.0", "IP地址")
	Port := flag.String("port", "8088", "监听端口")
	flag.Parse()

	server := grpc.NewServer()
	proto.RegisterUserServer(server, &handler.UserServer{})
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", *IP, *Port))
	if err != nil {
		log.Fatalln(err)
	}
	log.Fatalln(server.Serve(lis))
}
