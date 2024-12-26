package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
	"gorm.io/gorm"
)

type PurchaseOrderServiceServer struct {
	pb.UnimplementedPurchaseOrderServiceServer
	db *gorm.DB
}

// NewPurchaseOrderServiceServer 用于创建 PurchaseOrderServiceServer
func NewPurchaseOrderServiceServer(db *gorm.DB) *PurchaseOrderServiceServer {
	return &PurchaseOrderServiceServer{
		db: db,
	}
}

// CreatePurchaseOrder 创建采购单
func (s *PurchaseOrderServiceServer) CreatePurchaseOrder(ctx context.Context, req *pb.CreatePurchaseOrderRequest) (*pb.CreatePurchaseOrderResponse, error) {
	// 验证用户输入
	if req.GetBookNo() == "" {
		return &pb.CreatePurchaseOrderResponse{
			Success:  false,
			Feedback: "BookNo is required",
		}, nil
	}

	if req.GetTitle() == "" {
		return &pb.CreatePurchaseOrderResponse{
			Success:  false,
			Feedback: "Title is required",
		}, nil
	}

	if req.GetPublisher() == "" {
		return &pb.CreatePurchaseOrderResponse{
			Success:  false,
			Feedback: "Publisher is required",
		}, nil
	}

	if req.GetSupplier() == "" {
		return &pb.CreatePurchaseOrderResponse{
			Success:  false,
			Feedback: "Supplier is required",
		}, nil
	}

	if req.GetAuthor() == "" {
		return &pb.CreatePurchaseOrderResponse{
			Success:  false,
			Feedback: "Author is required",
		}, nil
	}

	if req.GetQuantity() <= 0 {
		return &pb.CreatePurchaseOrderResponse{
			Success:  false,
			Feedback: "Quantity must be greater than 0",
		}, nil
	}

	if req.GetOrderDate() == "" {
		return &pb.CreatePurchaseOrderResponse{
			Success:  false,
			Feedback: "Order date is required",
		}, nil
	}

	// 构建新的 PurchaseOrder 对象
	purchaseOrder := &models.PurchaseOrder{
		BookNo:    req.GetBookNo(),
		Title:     req.GetTitle(),
		Publisher: req.GetPublisher(),
		Supplier:  req.GetSupplier(),
		Author:    req.GetAuthor(),
		Quantity:  req.GetQuantity(),
		OrderDate: req.GetOrderDate(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 写入数据库
	if err := s.db.Create(&purchaseOrder).Error; err != nil {
		return &pb.CreatePurchaseOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to create purchase order: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.CreatePurchaseOrderResponse{
		Success:  true,
		Feedback: "Purchase order created successfully",
	}, nil
}

// GetPurchaseOrder 获取采购单
func (s *PurchaseOrderServiceServer) GetPurchaseOrder(ctx context.Context, req *pb.GetPurchaseOrderRequest) (*pb.GetPurchaseOrderResponse, error) {
	// 验证请求参数
	if req.GetStart() < 0 || req.GetStop() < req.GetStart() {
		return &pb.GetPurchaseOrderResponse{
			Success:  false,
			Feedback: "Invalid range: start must be >= 0 and stop must be >= start",
		}, nil
	}

	// 查询采购单
	var purchaseOrders []*models.PurchaseOrder
	if err := s.db.Offset(int(req.GetStart())).Limit(int(req.GetStop() - req.GetStart() + 1)).Find(&purchaseOrders).Error; err != nil {
		return &pb.GetPurchaseOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to query purchase orders: %v", err),
		}, nil
	}

	// 构建响应
	var pbPurchaseOrders []*pb.PurchaseOrder
	for _, purchaseOrder := range purchaseOrders {
		pbPurchaseOrder := &pb.PurchaseOrder{
			Id:        int32(purchaseOrder.ID),
			BookNo:    purchaseOrder.BookNo,
			Title:     purchaseOrder.Title,
			Publisher: purchaseOrder.Publisher,
			Supplier:  purchaseOrder.Supplier,
			Author:    purchaseOrder.Author,
			Quantity:  purchaseOrder.Quantity,
			OrderDate: purchaseOrder.OrderDate,
		}

		pbPurchaseOrders = append(pbPurchaseOrders, pbPurchaseOrder)
	}

	// 返回响应
	return &pb.GetPurchaseOrderResponse{
		Success:        true,
		Feedback:       "Purchase orders retrieved successfully",
		PurchaseOrders: pbPurchaseOrders,
	}, nil
}

// UpdatePurchaseOrder 更新采购单
func (s *PurchaseOrderServiceServer) UpdatePurchaseOrder(ctx context.Context, req *pb.UpdatePurchaseOrderRequest) (*pb.UpdatePurchaseOrderResponse, error) {
	var purchaseOrder models.PurchaseOrder

	// 查询采购单
	if err := s.db.First(&purchaseOrder, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.UpdatePurchaseOrderResponse{
				Success:  false,
				Feedback: "Purchase order not found",
			}, nil
		}
		return &pb.UpdatePurchaseOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query purchase order: %v", err),
		}, nil
	}

	// 更新字段
	if req.GetBookNo() != purchaseOrder.BookNo && req.GetBookNo() != "" {
		// 这个字段不能更改
		return &pb.UpdatePurchaseOrderResponse{
			Success:  false,
			Feedback: "BookNo cannot be updated",
		}, nil
	}
	if req.GetTitle() != "" {
		purchaseOrder.Title = req.GetTitle()
	}
	if req.GetPublisher() != "" {
		purchaseOrder.Publisher = req.GetPublisher()
	}
	if req.GetSupplier() != "" {
		purchaseOrder.Supplier = req.GetSupplier()
	}
	if req.GetAuthor() != "" {
		purchaseOrder.Author = req.GetAuthor()
	}
	if req.GetQuantity() != 0 {
		purchaseOrder.Quantity = req.GetQuantity()
	}
	if req.GetOrderDate() != "" {
		purchaseOrder.OrderDate = req.GetOrderDate()
	}
	if req.GetFinished() {
		purchaseOrder.Finished = req.GetFinished()

		// 更新书籍库存数量
		var book models.Book
		if err := s.db.Where("book_no = ?", purchaseOrder.BookNo).First(&book).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// 如果书籍不存在，创建书籍记录
				book = models.Book{
					BookNo:        purchaseOrder.BookNo,
					Title:         purchaseOrder.Title,
					PublisherName: purchaseOrder.Publisher,
					Authors:       purchaseOrder.Author,
					StockQuantity: purchaseOrder.Quantity,
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				}
				if err := s.db.Create(&book).Error; err != nil {
					return &pb.UpdatePurchaseOrderResponse{
						Success:  false,
						Feedback: fmt.Sprintf("failed to create book: %v", err),
					}, nil
				}
			} else {
				return &pb.UpdatePurchaseOrderResponse{
					Success:  false,
					Feedback: fmt.Sprintf("failed to find book: %v", err),
				}, nil
			}
		} else {
			book.StockQuantity += purchaseOrder.Quantity
			if err := s.db.Save(&book).Error; err != nil {
				return &pb.UpdatePurchaseOrderResponse{
					Success:  false,
					Feedback: fmt.Sprintf("failed to update book stock quantity: %v", err),
				}, nil
			}
		}
	}
	purchaseOrder.UpdatedAt = time.Now()

	// 保存更新
	if err := s.db.Save(&purchaseOrder).Error; err != nil {
		return &pb.UpdatePurchaseOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to update purchase order: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.UpdatePurchaseOrderResponse{
		Success:  true,
		Feedback: "Purchase order updated successfully",
	}, nil
}

// DeletePurchaseOrder 删除采购单
func (s *PurchaseOrderServiceServer) DeletePurchaseOrder(ctx context.Context, req *pb.DeletePurchaseOrderRequest) (*pb.DeletePurchaseOrderResponse, error) {
	var purchaseOrder models.PurchaseOrder

	// 查询采购单
	if err := s.db.First(&purchaseOrder, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.DeletePurchaseOrderResponse{
				Success:  false,
				Feedback: "Purchase order not found",
			}, nil
		}
		return &pb.DeletePurchaseOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query purchase order: %v", err),
		}, nil
	}

	// 删除采购单
	if err := s.db.Delete(&purchaseOrder).Error; err != nil {
		return &pb.DeletePurchaseOrderResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to delete purchase order: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.DeletePurchaseOrderResponse{
		Success:  true,
		Feedback: "Purchase order deleted successfully",
	}, nil
}

// GeneratePurchaseOrdersFromStockRequests 根据未完成的缺书记录生成采购单
func (s *PurchaseOrderServiceServer) GeneratePurchaseOrdersFromStockRequests(ctx context.Context, req *pb.GeneratePurchaseOrdersRequest) (*pb.GeneratePurchaseOrdersResponse, error) {
	// 查询所有 Finished=false 的缺书记录
	var stockRequests []*models.StockRequest
	if err := s.db.Where("finished = ?", false).Find(&stockRequests).Error; err != nil {
		return &pb.GeneratePurchaseOrdersResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to query stock requests: %v", err),
		}, nil
	}
	var feedback string
	// 生成采购单
	for _, stockRequest := range stockRequests {
		purchaseOrder := &models.PurchaseOrder{
			BookNo:    stockRequest.BookNo,
			Title:     stockRequest.Title,
			Publisher: stockRequest.Publisher,
			Supplier:  stockRequest.Supplier,
			Author:    stockRequest.Author,
			Quantity:  stockRequest.Quantity,
			OrderDate: time.Now().Format("2006-01-02"),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 写入数据库
		if err := s.db.Create(&purchaseOrder).Error; err != nil {
			return &pb.GeneratePurchaseOrdersResponse{
				Success:  false,
				Feedback: fmt.Sprintf("Failed to create purchase order: %v", err),
			}, nil
		}

		// 更新缺书记录的 Finished 字段
		stockRequest.Finished = true
		if err := s.db.Save(&stockRequest).Error; err != nil {
			return &pb.GeneratePurchaseOrdersResponse{
				Success:  false,
				Feedback: fmt.Sprintf("Failed to update stock request: %v", err),
			}, nil
		}

		// 发送电子邮件通想要买相应书的客户
		feedback = s.generateFakeEmail(*stockRequest)
	}

	// 返回成功的响应
	return &pb.GeneratePurchaseOrdersResponse{
		Success:  true,
		Feedback: fmt.Sprintf("Purchase orders generated successfully\n%s", feedback),
	}, nil
}

// generateFakeEmail 生成假电子邮件
func (s *PurchaseOrderServiceServer) generateFakeEmail(stockRequest models.StockRequest) string {
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
