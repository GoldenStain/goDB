package services

import (
	"context"
	"fmt"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
	"gorm.io/gorm"
)

type CustomerOrderServiceServer struct {
	pb.UnimplementedCustomerOrderServiceServer
	db *gorm.DB
}

// NewCustomerOrderServiceServer 用于创建 CustomerOrderServiceServer
func NewCustomerOrderServiceServer(db *gorm.DB) *CustomerOrderServiceServer {
	return &CustomerOrderServiceServer{
		db: db,
	}
}

// CreateCustomerOrder 创建客户订单
func (s *CustomerOrderServiceServer) CreateCustomerOrder(ctx context.Context, req *pb.CreateCustomerOrderRequest) (*pb.CreateCustomerOrderResponse, error) {
	// 验证用户输入
	if req.GetOrderDate() == "" {
		return &pb.CreateCustomerOrderResponse{
			Success:  false,
			Feedback: "Order date is required",
		}, nil
	}

	if req.GetCustomerOnlineId() == "" {
		return &pb.CreateCustomerOrderResponse{
			Success:  false,
			Feedback: "Customer ID or CustomerOnelineId is required",
		}, nil
	}

	if req.GetBookNo() == "" {
		return &pb.CreateCustomerOrderResponse{
			Success:  false,
			Feedback: "Book No is required",
		}, nil
	}

	if req.GetBookCount() <= 0 {
		return &pb.CreateCustomerOrderResponse{
			Success:  false,
			Feedback: "Book count must be greater than 0",
		}, nil
	}

	if req.GetPrice() <= 0 {
		return &pb.CreateCustomerOrderResponse{
			Success:  false,
			Feedback: "Price must be greater than 0",
		}, nil
	}

	if req.GetAddress() == "" {
		return &pb.CreateCustomerOrderResponse{
			Success:  false,
			Feedback: "Address is required",
		}, nil
	}

	if req.GetStatus() == "" {
		return &pb.CreateCustomerOrderResponse{
			Success:  false,
			Feedback: "Status is required",
		}, nil
	}

	// 查找客户
	var customer models.Customer
	if err := s.db.Where("online_id = ?", req.GetCustomerOnlineId()).First(&customer).Error; err != nil {
		return &pb.CreateCustomerOrderResponse{
			Success:  false,
			Feedback: "Customer not found",
		}, nil
	}

	// 检查客户余额是否足够
	if customer.AccountBalance < req.GetPrice() {
		return &pb.CreateCustomerOrderResponse{
			Success:  false,
			Feedback: "Insufficient account balance",
		}, nil
	}

	// 扣除客户余额
	customer.AccountBalance -= req.GetPrice()
	if err := s.db.Save(&customer).Error; err != nil {
		return &pb.CreateCustomerOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to update customer balance: %v", err),
		}, nil
	}

	// 构建新的 CustomerOrder 对象
	customerOrder := &models.CustomerOrder{
		OrderDate:        req.GetOrderDate(),
		CustomerOnlineID: req.GetCustomerOnlineId(),
		BookNo:           req.GetBookNo(),
		BookCount:        req.GetBookCount(),
		Price:            req.GetPrice(),
		Address:          req.GetAddress(),
		Status:           req.GetStatus(),
	}

	// 写入数据库
	if err := s.db.Create(&customerOrder).Error; err != nil {
		return &pb.CreateCustomerOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to create customer order: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.CreateCustomerOrderResponse{
		Success:  true,
		Feedback: "Customer order created successfully",
	}, nil
}

// GetCustomerOrder 获取客户订单
func (s *CustomerOrderServiceServer) GetCustomerOrder(ctx context.Context, req *pb.GetCustomerOrderRequest) (*pb.GetCustomerOrderResponse, error) {
	// 验证请求参数
	if req.GetStart() < 0 || req.GetStop() < req.GetStart() {
		return &pb.GetCustomerOrderResponse{
			Success:  false,
			Feedback: "Invalid range: start must be >= 0 and stop must be >= start",
		}, nil
	}

	// 查询客户订单
	var customerOrders []*models.CustomerOrder
	if err := s.db.Offset(int(req.GetStart())).Limit(int(req.GetStop() - req.GetStart() + 1)).Find(&customerOrders).Error; err != nil {
		return &pb.GetCustomerOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to query customer orders: %v", err),
		}, nil
	}

	// 构建响应
	var pbCustomerOrders []*pb.CustomerOrder
	for _, customerOrder := range customerOrders {
		pbCustomerOrder := &pb.CustomerOrder{
			Id:               customerOrder.ID,
			OrderDate:        customerOrder.OrderDate,
			CustomerOnlineId: customerOrder.CustomerOnlineID,
			BookNo:           customerOrder.BookNo,
			BookCount:        customerOrder.BookCount,
			Price:            customerOrder.Price,
			Address:          customerOrder.Address,
			Status:           customerOrder.Status,
		}

		pbCustomerOrders = append(pbCustomerOrders, pbCustomerOrder)
	}

	// 返回响应
	return &pb.GetCustomerOrderResponse{
		Success:        true,
		Feedback:       "Customer orders retrieved successfully",
		CustomerOrders: pbCustomerOrders,
	}, nil
}

// UpdateCustomerOrder 更新客户订单
func (s *CustomerOrderServiceServer) UpdateCustomerOrder(ctx context.Context, req *pb.UpdateCustomerOrderRequest) (*pb.UpdateCustomerOrderResponse, error) {
	var customerOrder models.CustomerOrder

	// 查询客户订单
	if err := s.db.First(&customerOrder, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.UpdateCustomerOrderResponse{
				Success:  false,
				Feedback: "Customer order not found",
			}, nil
		}
		return &pb.UpdateCustomerOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query customer order: %v", err),
		}, nil
	}

	// 更新字段
	if req.GetOrderDate() != "" {
		customerOrder.OrderDate = req.GetOrderDate()
	}
	if req.GetCustomerOnlineId() != "" {
		// 这个字段不能更改
		return &pb.UpdateCustomerOrderResponse{
			Success:  false,
			Feedback: "Customer ID or CustomerOnelineId cannot be changed",
		}, nil
	}
	if req.GetBookNo() != "" {
		customerOrder.BookNo = req.GetBookNo()
	}
	if req.GetBookCount() != 0 {
		customerOrder.BookCount = req.GetBookCount()
	}
	if req.GetPrice() != 0 {
		customerOrder.Price = req.GetPrice()
	}
	if req.GetAddress() != "" {
		customerOrder.Address = req.GetAddress()
	}
	if req.GetStatus() != "" {
		customerOrder.Status = req.GetStatus()
	}

	// 保存更新
	if err := s.db.Save(&customerOrder).Error; err != nil {
		return &pb.UpdateCustomerOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to update customer order: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.UpdateCustomerOrderResponse{
		Success:  true,
		Feedback: "Customer order updated successfully",
	}, nil
}

// DeleteCustomerOrder 删除客户订单
func (s *CustomerOrderServiceServer) DeleteCustomerOrder(ctx context.Context, req *pb.DeleteCustomerOrderRequest) (*pb.DeleteCustomerOrderResponse, error) {
	var customerOrder models.CustomerOrder

	// 查询客户订单
	if err := s.db.First(&customerOrder, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.DeleteCustomerOrderResponse{
				Success:  false,
				Feedback: "Customer order not found",
			}, nil
		}
		return &pb.DeleteCustomerOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query customer order: %v", err),
		}, nil
	}

	// 删除客户订单
	if err := s.db.Delete(&customerOrder).Error; err != nil {
		return &pb.DeleteCustomerOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to delete customer order: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.DeleteCustomerOrderResponse{
		Success:  true,
		Feedback: "Customer order deleted successfully",
	}, nil
}
