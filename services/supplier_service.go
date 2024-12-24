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

type SupplierServiceServer struct {
	pb.UnimplementedSupplierServiceServer
	db *gorm.DB
}

// NewSupplierServiceServer 用于创建 SupplierServiceServer
func NewSupplierServiceServer(db *gorm.DB) *SupplierServiceServer {
	return &SupplierServiceServer{
		db: db,
	}
}

type SupplyBookServiceServer struct {
	pb.UnimplementedSupplyBookServiceServer
	db *gorm.DB
}

// NewSupplyBookServiceServer 用于创建 SupplyBookServiceServer
func NewSupplyBookServiceServer(db *gorm.DB) *SupplyBookServiceServer {
	return &SupplyBookServiceServer{
		db: db,
	}
}

// CreateSupplier 创建供应商
func (s *SupplierServiceServer) CreateSupplier(ctx context.Context, req *pb.CreateSupplierRequest) (*pb.CreateSupplierResponse, error) {
	// 验证用户输入
	if req.GetName() == "" {
		return &pb.CreateSupplierResponse{
			Success:  false,
			Feedback: "Name is required",
		}, nil
	}

	// 构建新的 Supplier 对象
	supplier := &models.Supplier{
		Name:       req.GetName(),
		BasicInfo:  req.GetBasicInfo(),
		SupplyInfo: req.GetSupplyInfo(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// 写入数据库
	if err := s.db.Create(&supplier).Error; err != nil {
		return &pb.CreateSupplierResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to create supplier: %v", err),
		}, nil
	}

	// 添加供应的书籍
	for _, book := range req.GetSupplyBooks() {
		bookModel := models.SupplyBook{
			BookNo:        book.GetBookNo(),
			Title:         book.GetTitle(),
			PublisherName: book.GetPublisherName(),
			Price:         book.GetPrice(),
			Quantity:      book.GetQuantity(),
			SupplierID:    supplier.ID,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := s.db.Create(&bookModel).Error; err != nil {
			return &pb.CreateSupplierResponse{
				Success:  false,
				Feedback: fmt.Sprintf("failed to create book: %v", err),
			}, nil
		}
	}

	// 返回成功的响应
	return &pb.CreateSupplierResponse{
		Success:  true,
		Feedback: "Supplier created successfully",
	}, nil
}

// GetSupplier 获取供应商
func (s *SupplierServiceServer) GetSupplier(ctx context.Context, req *pb.GetSupplierRequest) (*pb.GetSupplierResponse, error) {
	var suppliers []models.Supplier

	// 查询供应商
	if err := s.db.Preload("SupplyBooks").Offset(int(req.GetStart())).Limit(int(req.GetStop() - req.GetStart() + 1)).Find(&suppliers).Error; err != nil {
		return &pb.GetSupplierResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query suppliers: %v", err),
		}, nil
	}

	// 构建响应
	var pbSuppliers []*pb.Supplier
	for _, supplier := range suppliers {
		var pbBooks []*pb.SupplyBook
		for _, book := range supplier.SupplyBooks {
			pbBook := &pb.SupplyBook{
				Id:            book.ID,
				BookNo:        book.BookNo,
				Title:         book.Title,
				PublisherName: book.PublisherName,
				Price:         book.Price,
				Quantity:      book.Quantity,
				SupplierId:    book.SupplierID,
				CreatedAt:     timestamppb.New(book.CreatedAt),
				UpdatedAt:     timestamppb.New(book.UpdatedAt),
			}
			pbBooks = append(pbBooks, pbBook)
		}

		pbSupplier := &pb.Supplier{
			Id:          supplier.ID,
			Name:        supplier.Name,
			BasicInfo:   supplier.BasicInfo,
			SupplyInfo:  supplier.SupplyInfo,
			SupplyBooks: pbBooks,
			CreatedAt:   timestamppb.New(supplier.CreatedAt),
			UpdatedAt:   timestamppb.New(supplier.UpdatedAt),
		}
		pbSuppliers = append(pbSuppliers, pbSupplier)
	}

	// 返回成功的响应
	return &pb.GetSupplierResponse{
		Success:   true,
		Feedback:  "Suppliers retrieved successfully",
		Suppliers: pbSuppliers,
	}, nil
}

// UpdateSupplier 更新供应商
func (s *SupplierServiceServer) UpdateSupplier(ctx context.Context, req *pb.UpdateSupplierRequest) (*pb.UpdateSupplierResponse, error) {
	var supplier models.Supplier

	// 查询供应商
	if err := s.db.First(&supplier, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.UpdateSupplierResponse{
				Success:  false,
				Feedback: "Supplier not found",
			}, nil
		}
		return &pb.UpdateSupplierResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query supplier: %v", err),
		}, nil
	}

	// 更新字段
	if req.GetName() != "" {
		supplier.Name = req.GetName()
	}
	if req.GetBasicInfo() != "" {
		supplier.BasicInfo = req.GetBasicInfo()
	}
	if req.GetSupplyInfo() != "" {
		supplier.SupplyInfo = req.GetSupplyInfo()
	}
	supplier.UpdatedAt = time.Now()

	// 保存更新
	if err := s.db.Save(&supplier).Error; err != nil {
		return &pb.UpdateSupplierResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to update supplier: %v", err),
		}, nil
	}

	// 更新供应的书籍
	if len(req.GetSupplyBooks()) > 0 {
		// 删除旧的供应书籍
		if err := s.db.Where("supplier_id = ?", supplier.ID).Delete(&models.SupplyBook{}).Error; err != nil {
			return &pb.UpdateSupplierResponse{
				Success:  false,
				Feedback: fmt.Sprintf("failed to delete old supply books: %v", err),
			}, nil
		}

		// 添加新的供应书籍
		for _, book := range req.GetSupplyBooks() {
			bookModel := models.SupplyBook{
				BookNo:        book.GetBookNo(),
				Title:         book.GetTitle(),
				PublisherName: book.GetPublisherName(),
				Price:         book.GetPrice(),
				Quantity:      book.GetQuantity(),
				SupplierID:    supplier.ID,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			if err := s.db.Create(&bookModel).Error; err != nil {
				return &pb.UpdateSupplierResponse{
					Success:  false,
					Feedback: fmt.Sprintf("failed to create new supply book: %v", err),
				}, nil
			}
		}
	}

	// 返回成功的响应
	return &pb.UpdateSupplierResponse{
		Success:  true,
		Feedback: "Supplier updated successfully",
	}, nil
}

// DeleteSupplier 删除供应商
func (s *SupplierServiceServer) DeleteSupplier(ctx context.Context, req *pb.DeleteSupplierRequest) (*pb.DeleteSupplierResponse, error) {
	var supplier models.Supplier

	// 查询供应商
	if err := s.db.First(&supplier, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.DeleteSupplierResponse{
				Success:  false,
				Feedback: "Supplier not found",
			}, nil
		}
		return &pb.DeleteSupplierResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query supplier: %v", err),
		}, nil
	}

	// 删除供应商
	if err := s.db.Delete(&supplier).Error; err != nil {
		return &pb.DeleteSupplierResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to delete supplier: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.DeleteSupplierResponse{
		Success:  true,
		Feedback: "Supplier deleted successfully",
	}, nil
}

// CreateSupplyBook 创建供书记录
func (s *SupplyBookServiceServer) CreateSupplyBook(ctx context.Context, req *pb.CreateSupplyBookRequest) (*pb.CreateSupplyBookResponse, error) {
	// 验证用户输入
	if req.GetBookNo() == "" {
		return &pb.CreateSupplyBookResponse{
			Success:  false,
			Feedback: "Book No is required",
		}, nil
	}

	if req.GetTitle() == "" {
		return &pb.CreateSupplyBookResponse{
			Success:  false,
			Feedback: "Title is required",
		}, nil
	}

	if req.GetPublisherName() == "" {
		return &pb.CreateSupplyBookResponse{
			Success:  false,
			Feedback: "Publisher Name is required",
		}, nil
	}

	if req.GetPrice() <= 0 {
		return &pb.CreateSupplyBookResponse{
			Success:  false,
			Feedback: "Price must be greater than 0",
		}, nil
	}

	if req.GetQuantity() <= 0 {
		return &pb.CreateSupplyBookResponse{
			Success:  false,
			Feedback: "Quantity must be greater than 0",
		}, nil
	}

	if req.GetSupplierId() <= 0 {
		return &pb.CreateSupplyBookResponse{
			Success:  false,
			Feedback: "Supplier ID is required",
		}, nil
	}

	// 构建新的 SupplyBook 对象
	supplyBook := &models.SupplyBook{
		BookNo:        req.GetBookNo(),
		Title:         req.GetTitle(),
		PublisherName: req.GetPublisherName(),
		Price:         req.GetPrice(),
		Quantity:      req.GetQuantity(),
		SupplierID:    req.GetSupplierId(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// 写入数据库
	if err := s.db.Create(&supplyBook).Error; err != nil {
		return &pb.CreateSupplyBookResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to create supply book: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.CreateSupplyBookResponse{
		Success:  true,
		Feedback: "Supply book created successfully",
	}, nil
}

// GetSupplyBooksBySupplier 获取供应商的供书记录
func (s *SupplyBookServiceServer) GetSupplyBooksBySupplier(ctx context.Context, req *pb.GetSupplyBooksBySupplierRequest) (*pb.GetSupplyBooksBySupplierResponse, error) {
	var supplyBooks []models.SupplyBook

	// 查询供书记录
	if err := s.db.Where("supplier_id = ?", req.GetSupplierId()).Find(&supplyBooks).Error; err != nil {
		return &pb.GetSupplyBooksBySupplierResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query supply books: %v", err),
		}, nil
	}

	// 构建响应
	var pbSupplyBooks []*pb.SupplyBook
	for _, book := range supplyBooks {
		pbBook := &pb.SupplyBook{
			Id:            book.ID,
			BookNo:        book.BookNo,
			Title:         book.Title,
			PublisherName: book.PublisherName,
			Price:         book.Price,
			Quantity:      book.Quantity,
			SupplierId:    book.SupplierID,
			CreatedAt:     timestamppb.New(book.CreatedAt),
			UpdatedAt:     timestamppb.New(book.UpdatedAt),
		}
		pbSupplyBooks = append(pbSupplyBooks, pbBook)
	}

	// 返回成功的响应
	return &pb.GetSupplyBooksBySupplierResponse{
		Success:     true,
		Feedback:    "Supply books retrieved successfully",
		SupplyBooks: pbSupplyBooks,
	}, nil
}

// GetSupplyBookByID 获取供书记录通过ID
func (s *SupplyBookServiceServer) GetSupplyBookByID(ctx context.Context, req *pb.GetSupplyBookByIDRequest) (*pb.GetSupplyBookByIDResponse, error) {
	var supplyBook models.SupplyBook

	// 查询供书记录
	if err := s.db.First(&supplyBook, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.GetSupplyBookByIDResponse{
				Success:  false,
				Feedback: "Supply book not found",
			}, nil
		}
		return &pb.GetSupplyBookByIDResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query supply book: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.GetSupplyBookByIDResponse{
		Success:  true,
		Feedback: "Supply book retrieved successfully",
		SupplyBook: &pb.SupplyBook{
			Id:            supplyBook.ID,
			BookNo:        supplyBook.BookNo,
			Title:         supplyBook.Title,
			PublisherName: supplyBook.PublisherName,
			Price:         supplyBook.Price,
			Quantity:      supplyBook.Quantity,
			SupplierId:    supplyBook.SupplierID,
			CreatedAt:     timestamppb.New(supplyBook.CreatedAt),
			UpdatedAt:     timestamppb.New(supplyBook.UpdatedAt),
		},
	}, nil
}

// UpdateSupplyBook 更新供书记录
func (s *SupplyBookServiceServer) UpdateSupplyBook(ctx context.Context, req *pb.UpdateSupplyBookRequest) (*pb.UpdateSupplyBookResponse, error) {
	var supplyBook models.SupplyBook

	// 查询供书记录
	if err := s.db.First(&supplyBook, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.UpdateSupplyBookResponse{
				Success:  false,
				Feedback: "Supply book not found",
			}, nil
		}
		return &pb.UpdateSupplyBookResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query supply book: %v", err),
		}, nil
	}

	// 更新字段
	if req.GetBookNo() != "" {
		supplyBook.BookNo = req.GetBookNo()
	}
	if req.GetTitle() != "" {
		supplyBook.Title = req.GetTitle()
	}
	if req.GetPublisherName() != "" {
		supplyBook.PublisherName = req.GetPublisherName()
	}
	if req.GetPrice() != -1 {
		supplyBook.Price = req.GetPrice()
	}
	if req.GetQuantity() != -1 {
		supplyBook.Quantity = req.GetQuantity()
	}
	if req.GetSupplierId() != -1 {
		supplyBook.SupplierID = req.GetSupplierId()
	}
	supplyBook.UpdatedAt = time.Now()

	// 保存更新
	if err := s.db.Save(&supplyBook).Error; err != nil {
		return &pb.UpdateSupplyBookResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to update supply book: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.UpdateSupplyBookResponse{
		Success:  true,
		Feedback: "Supply book updated successfully",
	}, nil
}

// DeleteSupplyBook 删除供书记录
func (s *SupplyBookServiceServer) DeleteSupplyBook(ctx context.Context, req *pb.DeleteSupplyBookRequest) (*pb.DeleteSupplyBookResponse, error) {
	var supplyBook models.SupplyBook

	// 查询供书记录
	if err := s.db.First(&supplyBook, req.GetId()).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.DeleteSupplyBookResponse{
				Success:  false,
				Feedback: "Supply book not found",
			}, nil
		}
		return &pb.DeleteSupplyBookResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to query supply book: %v", err),
		}, nil
	}

	// 删除供书记录
	if err := s.db.Delete(&supplyBook).Error; err != nil {
		return &pb.DeleteSupplyBookResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to delete supply book: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.DeleteSupplyBookResponse{
		Success:  true,
		Feedback: "Supply book deleted successfully",
	}, nil
}
