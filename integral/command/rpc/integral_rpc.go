package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"integral-mall/common/utils"
	"integral-mall/integral/command/rpc/config"
	"integral-mall/integral/logic"
	"integral-mall/integral/model"
	"integral-mall/integral/protos"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/yakaa/grpcx"
	"github.com/yakaa/log4g"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "config/config.json", "use config")

func main() {
	flag.Parse()
	conf := new(config.Config)
	bs, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.Unmarshal(bs, conf); err != nil {
		log.Fatal(err)
	}
	log4g.Init(log4g.Config{Path: "logs"})
	engine, err := xorm.NewEngine("mysql", conf.Mysql.DataSource)
	if err != nil {
		log.Println("xorm err")
		log.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{Addr: conf.Redis.DataSource, Password: conf.Redis.Auth})
	integralModel := model.NewIntegralModel(engine, client, conf.Mysql.Table.Integral)
	rabbitMqServer, err := utils.NewRabbitMqServer(
		conf.RabbitMq.DataSource+conf.RabbitMq.VirtualHost,
		conf.RabbitMq.QueueName,
	)
	if err != nil {
		log.Println("rabbitMqServer err")
		log.Fatal(err)
	}
	userServerLogic := logic.NewIntegralLogic(
		rabbitMqServer,
		integralModel)

	rpcServer, err := grpcx.MustNewGrpcxServer(conf.RpcServerConfig, func(server *grpc.Server) {
		protos.RegisterIntegralRpcServer(server, userServerLogic)
	})
	if err != nil {
		log.Println("rpcServer err")
		log.Fatal(err)
	}
	userServerLogic.ConsumeMessage()
	defer userServerLogic.Close()
	log4g.InfoFormat("Integral rpc server has start ad %s ....", conf.RpcServerConfig.ServerAddress)
	log4g.Error(rpcServer.Run())
}
