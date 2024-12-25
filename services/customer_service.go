package services

import (
	"context"
	"fmt"
	"time"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

const (
	threshold1 = 100
	threshold2 = 300
	threshold3 = 500
	threshold4 = 1000
	threshold5 = 3000
)

type CustomerServiceServer struct {
	pb.UnimplementedCustomerServiceServer
	db *gorm.DB
}

// NewCustomerServiceServer 用于创建 CustomerServiceServer
func NewCustomerServiceServer(db *gorm.DB) *CustomerServiceServer {
	return &CustomerServiceServer{
		db: db,
	}
}

// CreateCustomer 创建新客户
func (s *CustomerServiceServer) CreateCustomer(ctx context.Context, req *pb.CreateCustomerRequest) (*pb.CreateCustomerResponse, error) {
	// 验证用户输入
	if req.GetOnlineId() == "" {
		return &pb.CreateCustomerResponse{
			Success:  false,
			Feedback: "Online ID is required",
		}, nil
	}

	if req.GetPassword() == "" {
		return &pb.CreateCustomerResponse{
			Success:  false,
			Feedback: "Password is required",
		}, nil
	}

	if req.GetName() == "" {
		return &pb.CreateCustomerResponse{
			Success:  false,
			Feedback: "Name is required",
		}, nil
	}

	if req.GetAddress() == "" {
		return &pb.CreateCustomerResponse{
			Success:  false,
			Feedback: "Address is required",
		}, nil
	}

	// 构建新的 Customer 对象
	customer := &models.Customer{
		OnlineID:       req.GetOnlineId(),
		Password:       req.GetPassword(),
		Name:           req.GetName(),
		Address:        req.GetAddress(),
		AccountBalance: req.GetAccountBalance(),
		CreditLevel:    req.GetCreditLevel(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 写入数据库
	if err := s.db.Create(&customer).Error; err != nil {
		return &pb.CreateCustomerResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to create customer: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.CreateCustomerResponse{
		Success:  true,
		Feedback: "Customer created successfully",
	}, nil
}

// GetCustomer 获取客户详细信息
func (s *CustomerServiceServer) GetCustomer(ctx context.Context, req *pb.GetCustomerRequest) (*pb.GetCustomerResponse, error) {
	// 验证请求参数
	if req.GetStart() < 0 || req.GetStop() < req.GetStart() {
		return &pb.GetCustomerResponse{
			Success:  false,
			Feedback: "Invalid range: start must be >= 0 and stop must be >= start",
		}, nil
	}

	// 查询客户
	var customers []*models.Customer
	if err := s.db.Offset(int(req.GetStart())).Limit(int(req.GetStop() - req.GetStart() + 1)).Find(&customers).Error; err != nil {
		return &pb.GetCustomerResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to query customers: %v", err),
		}, nil
	}

	// 更新客户信用等级
	for _, customer := range customers {
		originalCreditLevel := customer.CreditLevel
		switch {
		case customer.AccountBalance >= threshold5:
			customer.CreditLevel = 5
		case customer.AccountBalance >= threshold4:
			customer.CreditLevel = 4
		case customer.AccountBalance >= threshold3:
			customer.CreditLevel = 3
		case customer.AccountBalance >= threshold2:
			customer.CreditLevel = 2
		case customer.AccountBalance >= threshold1:
			customer.CreditLevel = 1
		default:
			customer.CreditLevel = 0
		}

		// 如果信用等级有变化，更新数据库
		if customer.CreditLevel != originalCreditLevel {
			customer.UpdatedAt = time.Now()
			s.db.Save(&customer)
		}
	}

	// 构建响应
	var pbCustomers []*pb.Customer
	for _, customer := range customers {
		pbCustomer := &pb.Customer{
			Id:             customer.ID,
			OnlineId:       customer.OnlineID,
			Password:       customer.Password,
			Name:           customer.Name,
			Address:        customer.Address,
			AccountBalance: customer.AccountBalance,
			CreditLevel:    customer.CreditLevel,
			CreatedAt:      timestamppb.New(customer.CreatedAt),
			UpdatedAt:      timestamppb.New(customer.UpdatedAt),
		}

		pbCustomers = append(pbCustomers, pbCustomer)
	}

	// 返回响应
	return &pb.GetCustomerResponse{
		Success:   true,
		Feedback:  "Customers retrieved successfully",
		Customers: pbCustomers,
	}, nil
}

// UpdateCustomer 更新客户信息
func (s *CustomerServiceServer) UpdateCustomer(ctx context.Context, req *pb.UpdateCustomerRequest) (*pb.UpdateCustomerResponse, error) {
	var customer models.Customer

	// 查询客户
	if err := s.db.First(&customer, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.UpdateCustomerResponse{
				Success:  false,
				Feedback: "Customer not found",
			}, nil
		}
		return &pb.UpdateCustomerResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query customer: %v", err),
		}, nil
	}

	// 更新字段
	if req.GetOnlineId() != "" {
		customer.OnlineID = req.GetOnlineId()
	}
	if req.GetPassword() != "" {
		customer.Password = req.GetPassword()
	}
	if req.GetName() != "" {
		customer.Name = req.GetName()
	}
	if req.GetAddress() != "" {
		customer.Address = req.GetAddress()
	}
	if req.GetAccountBalance() != 0 {
		customer.AccountBalance = req.GetAccountBalance()
	}
	if req.GetCreditLevel() != 0 {
		customer.CreditLevel = req.GetCreditLevel()
	}
	customer.UpdatedAt = time.Now()

	// 保存更新
	if err := s.db.Save(&customer).Error; err != nil {
		return &pb.UpdateCustomerResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to update customer: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.UpdateCustomerResponse{
		Success:  true,
		Feedback: "Customer updated successfully",
	}, nil
}

// DeleteCustomer 删除客户记录
func (s *CustomerServiceServer) DeleteCustomer(ctx context.Context, req *pb.DeleteCustomerRequest) (*pb.DeleteCustomerResponse, error) {
	var customer models.Customer

	// 查询客户
	if err := s.db.First(&customer, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.DeleteCustomerResponse{
				Success:  false,
				Feedback: "Customer not found",
			}, nil
		}
		return &pb.DeleteCustomerResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query customer: %v", err),
		}, nil
	}

	// 删除客户记录
	if err := s.db.Delete(&customer).Error; err != nil {
		return &pb.DeleteCustomerResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to delete customer: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.DeleteCustomerResponse{
		Success:  true,
		Feedback: "Customer deleted successfully",
	}, nil
}
