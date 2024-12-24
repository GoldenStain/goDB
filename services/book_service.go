package services

import (
	"context"
	"fmt"
	"github.com/GoldenStain/goDB/models"
	"time"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type BookServiceServer struct {
	pb.UnimplementedBookServiceServer
	db *gorm.DB
}

// NewBookServiceServer 用于创建 BookServiceServer
func NewBookServiceServer(db *gorm.DB) *BookServiceServer {
	return &BookServiceServer{
		db: db,
	}
}

// CreateBook 创建新书
func (s *BookServiceServer) CreateBook(ctx context.Context, req *pb.CreateBookRequest) (*pb.CreateBookResponse, error) {
	// 验证用户输入
	if req.GetBookNo() == "" {
		return &pb.CreateBookResponse{
			Success:  false,
			Feedback: "BookNo is required",
		}, nil
	}

	if req.GetTitle() == "" {
		return &pb.CreateBookResponse{
			Success:  false,
			Feedback: "Title is required",
		}, nil
	}

	if req.GetPublisherName() == "" {
		return &pb.CreateBookResponse{
			Success:  false,
			Feedback: "Publisher name is required or publisher_id must be provided",
		}, nil
	}

	if req.GetPrice() <= 0 {
		return &pb.CreateBookResponse{
			Success:  false,
			Feedback: "Price must be greater than 0",
		}, nil
	}

	if req.GetStockQuantity() <= 0 {
		return &pb.CreateBookResponse{
			Success:  false,
			Feedback: "Stock quantity must be greater than 0",
		}, nil
	}

	// 构建新的Book对象
	book := &models.Book{
		BookNo:        req.GetBookNo(),
		Title:         req.GetTitle(),
		PublisherName: req.GetPublisherName(),
		Price:         req.GetPrice(),
		StockQuantity: req.GetStockQuantity(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Authors:       req.GetAuthors(),
		Keywords:      req.GetKeywords(),
	}

	// 写入数据库
	if err := s.db.Create(&book).Error; err != nil {
		return &pb.CreateBookResponse{
			Success:  false,
			Feedback: fmt.Sprintf("failed to create book: %v", err),
		}, nil
	}

	// 返回成功的响应
	return &pb.CreateBookResponse{
		Success:  true,
		Feedback: "Book created successfully",
	}, nil
}

// GetBook 根据 ID 获取书籍
func (s *BookServiceServer) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.GetBookResponse, error) {
	// 验证请求参数
	if req.GetStart() < 0 || req.GetStop() < req.GetStart() {
		return &pb.GetBookResponse{
			Success:  false,
			Feedback: "Invalid range: start must be >= 0 and stop must be >= start",
		}, nil
	}

	// 查询书籍
	var books []*models.Book
	if err := s.db.Offset(int(req.GetStart())).Limit(int(req.GetStop() - req.GetStart() + 1)).Find(&books).Error; err != nil {
		return &pb.GetBookResponse{
			Success:  false,
			Feedback: fmt.Sprintf("Failed to query books: %v", err),
		}, nil
	}

	// 构建响应
	var pbBooks []*pb.Book
	for _, book := range books {
		pbBook := &pb.Book{
			Id:            book.ID,
			BookNo:        book.BookNo,
			Title:         book.Title,
			Price:         book.Price,
			StockQuantity: book.StockQuantity,
			PublisherName: book.PublisherName,
			Keywords:      book.Keywords,
			Authors:       book.Authors,
			CreatedAt:     timestamppb.New(book.CreatedAt),
			UpdatedAt:     timestamppb.New(book.UpdatedAt),
		}

		pbBooks = append(pbBooks, pbBook)
	}

	// 返回响应
	return &pb.GetBookResponse{
		Success:  true,
		Feedback: "Books retrieved successfully",
		Books:    pbBooks,
	}, nil
}

// UpdateBook 更新现有书籍
func (s *BookServiceServer) UpdateBook(ctx context.Context, req *pb.UpdateBookRequest) (*pb.UpdateBookResponse, error) {
	return nil, nil
}

// DeleteBook 删除书籍
func (s *BookServiceServer) DeleteBook(ctx context.Context, req *pb.DeleteBookRequest) (*pb.DeleteBookResponse, error) {
	return nil, nil
}
