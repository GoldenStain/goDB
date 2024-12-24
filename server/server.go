package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
	"github.com/GoldenStain/goDB/services"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config 用来读取配置文件中的数据
type Config struct {
	DB struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
	} `json:"db"`
}

// 读取配置文件
func loadConfig(file string) (*Config, error) {
	var config Config
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// 初始化数据库连接
func initDB(config *Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai",
		config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func ConnectDB() *gorm.DB {
	// 加载配置文件
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("无法读取main.go位置！")
	}
	config, err := loadConfig(fmt.Sprintf("%s\\server\\config.json", dir))
	if err != nil {
		log.Fatalf("无法加载配置文件: %v", err)
	}

	// 初始化数据库连接
	db, err := initDB(config)
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}

	fmt.Println("数据库连接成功:", db)

	// 自动迁移
	err = models.AutoMigrate(db)
	if err != nil {
		log.Fatalf("无法迁移数据库: %v", err)
	}
	return db
}

func registerRpcServices(gServer *grpc.Server, db *gorm.DB) {
	// 书籍
	bookService := services.NewBookServiceServer(db)
	pb.RegisterBookServiceServer(gServer, bookService)

	// 缺书登记
	stockRequestService := services.NewStockRequestServiceServer(db)
	pb.RegisterStockRequestServiceServer(gServer, stockRequestService)

	// 采购单
	purchaseOrderService := services.NewPurchaseOrderServiceServer(db)
	pb.RegisterPurchaseOrderServiceServer(gServer, purchaseOrderService)

	// 客户
	customerService := services.NewCustomerServiceServer(db)
	pb.RegisterCustomerServiceServer(gServer, customerService)

	// 销售单/订单
	customerOrderService := services.NewCustomerOrderServiceServer(db)
	pb.RegisterCustomerOrderServiceServer(gServer, customerOrderService)
}

func StartServer(db *gorm.DB) {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	registerRpcServices(grpcServer, db)

	log.Printf("server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
