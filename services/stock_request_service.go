package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type StockRequestServiceServer struct {
	pb.UnimplementedStockRequestServiceServer
	db *gorm.DB
}

// NewStockRequestServiceServer 用于创建 StockRequestServiceServer
func NewStockRequestServiceServer(db *gorm.DB) *StockRequestServiceServer {
	return &StockRequestServiceServer{
		db: db,
	}
}

// CreateStockRequest 创建缺书登记
func (s *StockRequestServiceServer) CreateStockRequest(ctx context.Context, req *pb.CreateStockRequestRequest) (*pb.CreateStockRequestResponse, error) {
	// 验证用户输入
	if req.GetBookNo() == "" {
		return &pb.CreateStockRequestResponse{
			Success:  false,
			Feedback: "BookNo is required",
		}, nil
	}

	if req.GetTitle() == "" {
		return &pb.CreateStockRequestResponse{
			Success:  false,
			Feedback: "Title is required",
		}, nil
	}

	if req.GetPublisher() == "" {
		return &pb.CreateStockRequestResponse{
			Success:  false,
			Feedback: "Publisher is required",
		}, nil
	}

	if req.GetQuantity() <= 0 {
		return &pb.CreateStockRequestResponse{
			Success:  false,
			Feedback: "Quantity must be greater than 0",
		}, nil
	}

	if req.GetRequestDate() == "" {
		return &pb.CreateStockRequestResponse{
			Success:  false,
			Feedback: "Request date is required",
		}, nil
	}

	// 构建新的 StockRequest 对象
	stockRequest := &models.StockRequest{
		BookNo:      req.GetBookNo(),
		Title:       req.GetTitle(),
		Publisher:   req.GetPublisher(),
		Supplier:    req.GetSupplier(),
		Author:      req.GetAuthor(),
		Quantity:    req.GetQuantity(),
		RequestDate: req.GetRequestDate(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 写入数据库
	if err := s.db.Create(&stockRequest).Error; err != nil {
		return &pb.CreateStockRequestResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to create stock request: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.CreateStockRequestResponse{
		Success:  true,
		Feedback: "Stock request created successfully",
	}, nil
}

// UpdateStockRequest 更新缺书登记
func (s *StockRequestServiceServer) UpdateStockRequest(ctx context.Context, req *pb.UpdateStockRequestRequest) (*pb.UpdateStockRequestResponse, error) {
	var stockRequest models.StockRequest
	var err error

	// 如果有ID，则利用ID查询
	if req.GetId() != 0 {
		err = s.db.First(&stockRequest, req.GetId()).Error
	} else {
		// 如果没有ID，则利用其他字段查询
		query := s.db.Model(&models.StockRequest{})
		if req.GetBookNo() != "" {
			query = query.Where("book_no = ?", req.GetBookNo())
		}
		if req.GetTitle() != "" {
			query = query.Where("title = ?", req.GetTitle())
		}
		if req.GetPublisher() != "" {
			query = query.Where("publisher = ?", req.GetPublisher())
		}
		if req.GetSupplier() != "" {
			query = query.Where("supplier = ?", req.GetSupplier())
		}
		if req.GetAuthor() != "" {
			query = query.Where("author = ?", req.GetAuthor())
		}
		if req.GetQuantity() != 0 {
			query = query.Where("quantity = ?", req.GetQuantity())
		}
		if req.GetRequestDate() != "" {
			query = query.Where("request_date = ?", req.GetRequestDate())
		}

		err = query.First(&stockRequest).Error
	}

	// 如果查询失败，返回错误
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.UpdateStockRequestResponse{
				Success:  false,
				Feedback: "Stock request not found",
			}, nil
		}
		return &pb.UpdateStockRequestResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to query stock request: %v", err),
		}, nil
	}

	// 更新字段
	stockRequest.Finished = req.GetFinished() || stockRequest.Finished

	// 保存更新
	if err := s.db.Save(&stockRequest).Error; err != nil {
		return &pb.UpdateStockRequestResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to update stock request: %v", err),
		}, nil
	}

	emailFeedback := "Stock request updated successfully"

	// 生成假电子邮件
	if stockRequest.Finished {
		emailFeedback = s.generateFakeEmail(stockRequest)
	}

	// 返回成功的响应
	return &pb.UpdateStockRequestResponse{
		Success:  true,
		Feedback: emailFeedback,
	}, nil
}

// GetStockRequest 获取缺书登记
func (s *StockRequestServiceServer) GetStockRequest(ctx context.Context, req *pb.GetStockRequestRequest) (*pb.GetStockRequestResponse, error) {
	// 验证请求参数
	if req.GetStart() < 0 || req.GetStop() < req.GetStart() {
		return &pb.GetStockRequestResponse{
			Success:  false,
			Feedback: "Invalid range: start must be >= 0 and stop must be >= start",
		}, nil
	}

	// 查询缺书登记
	var stockRequests []*models.StockRequest
	if err := s.db.Offset(int(req.GetStart())).Limit(int(req.GetStop() - req.GetStart() + 1)).Find(&stockRequests).Error; err != nil {
		return &pb.GetStockRequestResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to query stock requests: %v", err),
		}, nil
	}

	// 构建响应
	var pbStockRequests []*pb.StockRequest
	for _, stockRequest := range stockRequests {
		pbStockRequest := &pb.StockRequest{
			Id:          stockRequest.ID,
			BookNo:      stockRequest.BookNo,
			Title:       stockRequest.Title,
			Publisher:   stockRequest.Publisher,
			Supplier:    stockRequest.Supplier,
			Author:      stockRequest.Author,
			Quantity:    stockRequest.Quantity,
			RequestDate: stockRequest.RequestDate,
			Finished:    stockRequest.Finished,
			CreatedAt:   timestamppb.New(stockRequest.CreatedAt),
			UpdatedAt:   timestamppb.New(stockRequest.UpdatedAt),
		}

		pbStockRequests = append(pbStockRequests, pbStockRequest)
	}

	// 返回响应
	return &pb.GetStockRequestResponse{
		Success:       true,
		Feedback:      "Stock requests retrieved successfully",
		StockRequests: pbStockRequests,
	}, nil
}

// DeleteStockRequest 删除缺书登记
func (s *StockRequestServiceServer) DeleteStockRequest(ctx context.Context, req *pb.DeleteStockRequestRequest) (*pb.DeleteStockRequestResponse, error) {
	// 查找缺书登记
	var stockRequest models.StockRequest
	if err := s.db.First(&stockRequest, req.GetStockRequestId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.DeleteStockRequestResponse{
				Success:  false,
				Feedback: "Stock request not found",
			}, nil
		}
		return &pb.DeleteStockRequestResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to query stock request: %v", err),
		}, nil
	}

	// 删除缺书登记
	if err := s.db.Delete(&stockRequest).Error; err != nil {
		return &pb.DeleteStockRequestResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to delete stock request: %v", err),
		}, nil
	}

	feedback := "Stock request deleted successfully"

	// 生成假电子邮件
	if !stockRequest.Finished {
		// 如果缺书登记没有完成，生成假电子邮件
		feedback = s.generateFakeEmail(stockRequest)
	}

	// 返回成功的响应
	return &pb.DeleteStockRequestResponse{
		Success:  true,
		Feedback: feedback,
	}, nil
}

// generateFakeEmail 生成假电子邮件
func (s *StockRequestServiceServer) generateFakeEmail(stockRequest models.StockRequest) string {
	// 创建目录
	dir := ".\\fake_emails"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		fmt.Printf("Failed to create directory: %v\n", err)
		return "Failed to create directory"
	}

	// 查找相关客户订单
	var customerOrders []models.CustomerOrder
	if err := s.db.Where("book_no = ?", stockRequest.BookNo).Find(&customerOrders).Error; err != nil {
		fmt.Printf("Failed to query customer orders: %v\n", err)
		return "Failed to query customer orders"
	}

	if len(customerOrders) == 0 {
		return fmt.Sprintf("没有客户创建了书目%s的相关订单，无须发送邮件", stockRequest.Title)
	}

	// 生成假电子邮件
	for _, order := range customerOrders {
		var customer models.Customer
		if err := s.db.Where("online_id = ?", order.CustomerOnlineID).First(&customer).Error; err != nil {
			fmt.Printf("Failed to query customer: %v\n", err)
			continue
		}

		emailContent := fmt.Sprintf("客户%s您好，您要的%s书目已经上新", customer.Name, stockRequest.Title)
		today := time.Now().Format("2006-01-02")
		emailFile := filepath.Join(dir, fmt.Sprintf("email_%d_%s.txt", order.ID, today))
		if err := os.WriteFile(emailFile, []byte(emailContent), 0644); err != nil {
			fmt.Printf("Failed to write email file: %v\n", err)
		}
	}

	return "已暂存电子邮件通知客户"
}
