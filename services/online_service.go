package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	pb "github.com/GoldenStain/goDB/bookstorepb"
	"github.com/GoldenStain/goDB/models"
	"github.com/agnivade/levenshtein"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// OnlineServiceServer 定义服务
type OnlineServiceServer struct {
	pb.UnimplementedOnlineServiceServer
	db             *gorm.DB
	matchThreshold int32
}

// NewOnlineServiceServer 用于创建 OnlineServiceServer
func NewOnlineServiceServer(db *gorm.DB, matchThreshole_in int32) *OnlineServiceServer {
	return &OnlineServiceServer{
		db:             db,
		matchThreshold: matchThreshole_in,
	}
}

// QueryCustomer 查询客户
func (s *OnlineServiceServer) QueryCustomer(ctx context.Context, req *pb.QueryCustomerRequest) (*pb.QueryCustomerResponse, error) {
	var customers []models.Customer
	input := req.GetInput()
	customerMap := make(map[int32]models.Customer)
	feedbackMap := make(map[int32]string)

	if input != "" {
		// 尝试匹配 online_id
		var onlineIDCustomers []models.Customer
		if err := s.db.Preload("CustomerOrders").Where("online_id LIKE ?", "%"+input+"%").Find(&onlineIDCustomers).Error; err == nil && len(onlineIDCustomers) > 0 {
			for _, customer := range onlineIDCustomers {
				customerMap[customer.ID] = customer
				feedbackMap[customer.ID] = "Matched by online_id"
			}
		}

		// 尝试匹配 name
		var nameCustomers []models.Customer
		if err := s.db.Preload("CustomerOrders").Where("name LIKE ?", "%"+input+"%").Find(&nameCustomers).Error; err == nil && len(nameCustomers) > 0 {
			for _, customer := range nameCustomers {
				customerMap[customer.ID] = customer
				feedbackMap[customer.ID] = "Matched by name"
			}
		}

		// 尝试匹配 address
		var addressCustomers []models.Customer
		if err := s.db.Preload("CustomerOrders").Where("address LIKE ?", "%"+input+"%").Find(&addressCustomers).Error; err == nil && len(addressCustomers) > 0 {
			for _, customer := range addressCustomers {
				customerMap[customer.ID] = customer
				feedbackMap[customer.ID] = "Matched by address"
			}
		}

		// 尝试将输入转换为整数并匹配 CustomerOrder ID
		if orderId, err := strconv.Atoi(input); err == nil {
			var customerOrder models.CustomerOrder
			if err := s.db.First(&customerOrder, orderId).Error; err == nil {
				var customer models.Customer
				if err := s.db.Preload("CustomerOrders").Where("online_id = ?", customerOrder.CustomerOnlineID).First(&customer).Error; err == nil {
					customerMap[customer.ID] = customer
					feedbackMap[customer.ID] = "Matched by CustomerOrder ID"
				}
			}
		}
	}

	if len(customerMap) == 0 {
		return &pb.QueryCustomerResponse{
			Success:  false,
			Feedback: "No customers found",
		}, nil
	}

	for _, customer := range customerMap {
		customers = append(customers, customer)
	}

	var pbCustomers []*pb.Customer
	for _, customer := range customers {
		var pbOrders []*pb.CustomerOrder
		for _, order := range customer.CustomerOrders {
			pbOrder := &pb.CustomerOrder{
				Id:               order.ID,
				OrderDate:        order.OrderDate,
				CustomerOnlineId: order.CustomerOnlineID,
				BookNo:           order.BookNo,
				BookCount:        order.BookCount,
				Price:            order.Price,
				Address:          order.Address,
				Status:           order.Status,
				CreatedAt:        timestamppb.New(order.CreatedAt),
				UpdatedAt:        timestamppb.New(order.UpdatedAt),
			}
			pbOrders = append(pbOrders, pbOrder)
		}
		pbCustomer := &pb.Customer{
			OnlineId:       customer.OnlineID,
			Name:           fmt.Sprintf("%s (%s)", customer.Name, feedbackMap[customer.ID]),
			Address:        customer.Address,
			CustomerOrders: pbOrders,
		}
		pbCustomers = append(pbCustomers, pbCustomer)
	}

	return &pb.QueryCustomerResponse{
		Success:   true,
		Feedback:  "Customers found",
		Customers: pbCustomers,
	}, nil
}

// QueryBook 查询书籍
func (s *OnlineServiceServer) QueryBook(ctx context.Context, req *pb.QueryBookRequest) (*pb.QueryBookResponse, error) {
	var books []models.Book
	input := req.GetInput()
	bookMap := make(map[int32]models.Book)
	feedbackMap := make(map[int32]string)
	matchScoreMap := make(map[int32]float64)

	if input != "" {
		// 尝试匹配 book_no
		var bookNoBooks []models.Book
		if err := s.db.Where("book_no LIKE ?", "%"+input+"%").Find(&bookNoBooks).Error; err == nil && len(bookNoBooks) > 0 {
			for _, book := range bookNoBooks {
				if matchScore := getMatchScore(input, book.BookNo); matchScore >= float64(s.matchThreshold) {
					bookMap[book.ID] = book
					feedbackMap[book.ID] = "Matched by book_no"
					matchScoreMap[book.ID] = matchScore
				}
			}
		}

		// 尝试匹配 title
		var titleBooks []models.Book
		if err := s.db.Where("title LIKE ?", "%"+input+"%").Find(&titleBooks).Error; err == nil && len(titleBooks) > 0 {
			for _, book := range titleBooks {
				if matchScore := getMatchScore(input, book.Title); matchScore >= float64(s.matchThreshold) {
					bookMap[book.ID] = book
					feedbackMap[book.ID] = "Matched by title"
					matchScoreMap[book.ID] = matchScore
				}
			}
		}

		// 尝试匹配 publisher_name
		var publisherNameBooks []models.Book
		if err := s.db.Where("publisher_name LIKE ?", "%"+input+"%").Find(&publisherNameBooks).Error; err == nil && len(publisherNameBooks) > 0 {
			for _, book := range publisherNameBooks {
				if matchScore := getMatchScore(input, book.PublisherName); matchScore >= float64(s.matchThreshold) {
					bookMap[book.ID] = book
					feedbackMap[book.ID] = "Matched by publisher_name"
					matchScoreMap[book.ID] = matchScore
				}
			}
		}

		// 尝试匹配 keywords
		var keywordsBooks []models.Book
		if err := s.db.Where("keywords LIKE ?", "%"+input+"%").Find(&keywordsBooks).Error; err == nil && len(keywordsBooks) > 0 {
			for _, book := range keywordsBooks {
				if matchScore := getKeywordsMatchScore(input, book.Keywords, s); matchScore >= float64(s.matchThreshold) {
					bookMap[book.ID] = book
					feedbackMap[book.ID] = "Matched by keywords"
					matchScoreMap[book.ID] = matchScore
				}
			}
		}

		// 尝试匹配 authors
		var authorsBooks []models.Book
		if err := s.db.Where("authors LIKE ?", "%"+input+"%").Find(&authorsBooks).Error; err == nil && len(authorsBooks) > 0 {
			for _, book := range authorsBooks {
				if matchScore := getAuthorsMatchScore(input, book.Authors, s); matchScore >= float64(s.matchThreshold) {
					bookMap[book.ID] = book
					feedbackMap[book.ID] = "Matched by authors"
					matchScoreMap[book.ID] = matchScore
				}
			}
		}
	}

	if len(bookMap) == 0 {
		return &pb.QueryBookResponse{
			Success:  false,
			Feedback: "No books found",
		}, nil
	}

	for _, book := range bookMap {
		books = append(books, book)
	}

	var pbBooks []*pb.Book
	for _, book := range books {
		pbBook := &pb.Book{
			Id:            book.ID,
			BookNo:        book.BookNo,
			Title:         fmt.Sprintf("%s (%s, %.2f%%)", book.Title, feedbackMap[book.ID], matchScoreMap[book.ID]),
			PublisherName: book.PublisherName,
			Price:         book.Price,
			Keywords:      book.Keywords,
			Authors:       book.Authors,
			StockQuantity: book.StockQuantity,
			CreatedAt:     timestamppb.New(book.CreatedAt),
			UpdatedAt:     timestamppb.New(book.UpdatedAt),
		}
		pbBooks = append(pbBooks, pbBook)
	}

	return &pb.QueryBookResponse{
		Success:  true,
		Feedback: "Books found",
		Books:    pbBooks,
	}, nil
}

func getMatchScore(query, target string) float64 {
	distance := levenshtein.ComputeDistance(query, target)
	maxLen := max(len(query), len(target))
	return (1 - float64(distance)/float64(maxLen)) * 100
}

func getKeywordsMatchScore(query, target string, s *OnlineServiceServer) float64 {
	queryKeywords := strings.Split(query, ",")
	targetKeywords := strings.Split(target, ",")

	var totalScore float64
	var matchedCount int

	for _, qk := range queryKeywords {
		for _, tk := range targetKeywords {
			if matchScore := getMatchScore(qk, tk); matchScore >= float64(s.matchThreshold) {
				totalScore += matchScore
				matchedCount++
				break
			}
		}
	}

	if matchedCount == 0 {
		return 0
	}

	return totalScore / float64(matchedCount)
}

func getAuthorsMatchScore(query, target string, s *OnlineServiceServer) float64 {
	queryAuthors := strings.Split(query, ",")
	targetAuthors := strings.Split(target, ",")

	var totalScore float64
	var matchedCount int

	for _, qa := range queryAuthors {
		for _, ta := range targetAuthors {
			if matchScore := getMatchScore(qa, ta); matchScore >= float64(s.matchThreshold) {
				totalScore += matchScore
				matchedCount++
				break
			}
		}
	}

	if matchedCount == 0 {
		return 0
	}

	return totalScore / float64(matchedCount)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
