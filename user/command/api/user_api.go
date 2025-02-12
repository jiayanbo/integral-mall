package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"github.com/yakaa/grpcx"
	"github.com/yakaa/log4g"

	_ "github.com/go-sql-driver/mysql"

	"integral-mall/common/rpcxclient/integralrpcmodel"
	"integral-mall/user/command/api/config"
	"integral-mall/user/controller"
	"integral-mall/user/logic"
	"integral-mall/user/model"
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
	gin.DefaultWriter = log4g.InfoLog
	gin.DefaultErrorWriter = log4g.ErrorLog

	engine, err := xorm.NewEngine("mysql", conf.Mysql.DataSource)
	if err != nil {
		log.Println("1")
		log.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{Addr: conf.Redis.DataSource, Password: conf.Redis.Auth})
	log.Println("conf.IntegralRpc: ", conf.IntegralRpc)
	rpcxClient, err := grpcx.MustNewGrpcxClient(conf.IntegralRpc)
	if err != nil {
		log.Println("2")
		log.Fatal(err)
	}
	integralRpcModel := integralrpcmodel.NewIntegralRpcModel(
		rpcxClient,
	)
	
	userModel := model.NewUserModel(engine, client, conf.Mysql.Table.User)
	userLogic := logic.NewUserLogic(userModel, client, integralRpcModel)
	userController := controller.NewUserController(userLogic)

	r := gin.Default()

	userRouteGroup := r.Group("/user")
	{
		userRouteGroup.POST("/register", userController.Register)
		userRouteGroup.POST("/login", userController.Login)

	}
	log4g.Error(r.Run(conf.Port)) // listen and serve on 0.0.0.0:8080
}
