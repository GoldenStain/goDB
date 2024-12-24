package services

import (
	"context"
	"fmt"
	"time"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
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
	stockRequest.Finished = true

	// 保存更新
	if err := s.db.Save(&stockRequest).Error; err != nil {
		return &pb.UpdateStockRequestResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to update stock request: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.UpdateStockRequestResponse{
		Success:  true,
		Feedback: "Stock request updated successfully",
	}, nil
}
