package handler

import (
	"fmt"
	"testing"

	"github.com/JohnsonYuanTW/NCAEats/models"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockRestaurantRepository struct {
	mock.Mock
}

func (m *MockRestaurantRepository) Init() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRestaurantRepository) CreateRestaurant(restaurant *models.Restaurant) error {
	args := m.Called(restaurant)
	return args.Error(0)
}

func (m *MockRestaurantRepository) GetAllRestaurants() ([]*models.Restaurant, error) {
	args := m.Called()
	return args.Get(0).([]*models.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) GetRestaurantByName(name string) (*models.Restaurant, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Restaurant), args.Error(1)
}

func (m *MockRestaurantRepository) DeleteRestaurant(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Init() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOrderRepository) CreateOrder(order *models.Order) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockOrderRepository) GetActiveOrders() ([]*models.Order, error) {
	args := m.Called()
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepository) GetActiveOrdersOfOwnerID(ownerID string) ([]*models.Order, error) {
	args := m.Called(ownerID)
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepository) CountActiveOrdersOfOwnerID(ownerID string) (int64, error) {
	args := m.Called(ownerID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepository) SaveOrderReport(orderID uint, report string) error {
	args := m.Called(orderID, report)
	return args.Error(0)
}

func (m *MockOrderRepository) GenerateUniqueReportID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOrderRepository) GetOrderReportByOrderID(orderID uint) (string, error) {
	args := m.Called(orderID)
	return args.String(0), args.Error(1)
}

func (m *MockOrderRepository) GetOrderReportIDByOrderID(orderID uint) (string, error) {
	args := m.Called(orderID)
	return args.String(0), args.Error(1)
}

func (m *MockOrderRepository) GetOrderIDByReportID(reportID string) (uint, error) {
	args := m.Called(reportID)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockOrderRepository) DeleteOrderByOrderID(orderID uint) error {
	args := m.Called(orderID)
	return args.Error(0)
}

type MockOrderDetailRepository struct {
	mock.Mock
}

func (m *MockOrderDetailRepository) Init() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOrderDetailRepository) CreateOrderDetail(detail *models.OrderDetail) error {
	args := m.Called(detail)
	return args.Error(0)
}

func (m *MockOrderDetailRepository) GetActiveOrderDetailsByOrderID(orderID uint) ([]*models.OrderDetail, error) {
	args := m.Called(orderID)
	return args.Get(0).([]*models.OrderDetail), args.Error(1)
}

func (m *MockOrderDetailRepository) DeleteOrderDetailsByOrderID(orderID uint) error {
	args := m.Called(orderID)
	return args.Error(0)
}

type MockMenuItemRepository struct {
	mock.Mock
}

func (m *MockMenuItemRepository) Init() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMenuItemRepository) CreateMenuItem(item *models.MenuItem) error {
	args := m.Called(item)
	return args.Error(0)
}

func (m *MockMenuItemRepository) GetMenuItemsByRestaurantName(name string) ([]*models.MenuItem, error) {
	args := m.Called(name)
	return args.Get(0).([]*models.MenuItem), args.Error(1)
}

func (m *MockMenuItemRepository) GetMenuItemByDetails(detail1, detail2 string) (*models.MenuItem, error) {
	args := m.Called(detail1, detail2)
	return args.Get(0).(*models.MenuItem), args.Error(1)
}

type MockTemplateHandler struct {
	mock.Mock
}

func (m *MockTemplateHandler) Init() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTemplateHandler) GetTemplate(templateName string) (string, error) {
	args := m.Called(templateName)
	return args.String(0), args.Error(1)
}

func (m *MockTemplateHandler) generateFlexContainer(templateName string, args ...interface{}) (linebot.FlexContainer, error) {
	mockArgs := m.Called(templateName, args)
	return mockArgs.Get(0).(linebot.FlexContainer), mockArgs.Error(1)
}

func (m *MockTemplateHandler) generateBoxComponent(templateName string, args ...interface{}) (linebot.BoxComponent, error) {
	mockArgs := m.Called(templateName, args)
	return mockArgs.Get(0).(linebot.BoxComponent), mockArgs.Error(1)
}

func TestHandleNewOrder(t *testing.T) {
	var (
		appHandler          AppHandler
		mockRestaurantRepo  MockRestaurantRepository
		mockOrderRepo       MockOrderRepository
		mockMenuItemRepo    MockMenuItemRepository
		mockTemplateHandler MockTemplateHandler
	)

	appHandler.RestaurantRepo = &mockRestaurantRepo
	appHandler.MenuItemRepo = &mockMenuItemRepo
	appHandler.OrderRepo = &mockOrderRepo
	appHandler.Templates = &mockTemplateHandler

	t.Run("should return error on invalid input", func(t *testing.T) {
		_, err := appHandler.handleNewOrder([]string{}, "123")
		assert.Equal(t, ErrInputError, err)
	})

	t.Run("should handle not found restaurant", func(t *testing.T) {
		mockRestaurantRepo.On("GetRestaurantByName", "unknownRestaurant").Return(nil, gorm.ErrRecordNotFound)
		_, err := appHandler.handleNewOrder([]string{"unknownRestaurant"}, "123")
		assert.Equal(t, ErrRestaurantNotFound, err)
	})

	// ... Other tests for handleNewOrder ...

	mockRestaurantRepo.AssertExpectations(t)
	mockMenuItemRepo.AssertExpectations(t)
	mockOrderRepo.AssertExpectations(t)
	mockTemplateHandler.AssertExpectations(t)
}

func TestFetchRestaurant(t *testing.T) {
	// Setting up our AppHandler with its dependencies.
	var (
		appHandler         AppHandler
		mockRestaurantRepo MockRestaurantRepository
		logger             = logrus.New()
	)

	appHandler.RestaurantRepo = &mockRestaurantRepo
	appHandler.Logger = logger

	t.Run("should handle not found restaurant", func(t *testing.T) {
		// Define the behavior for our mock when "GetRestaurantByName" is called.
		mockRestaurantRepo.On("GetRestaurantByName", "unknownRestaurant").Return(nil, gorm.ErrRecordNotFound)

		_, err := appHandler.fetchRestaurant("unknownRestaurant")
		assert.Equal(t, ErrRestaurantNotFound, err)

		// This ensures that our mock's expected calls are equivalent to its actual calls.
		mockRestaurantRepo.AssertExpectations(t)
	})

	t.Run("should handle other errors while fetching restaurant", func(t *testing.T) {
		// Simulating a random DB error.
		mockRestaurantRepo.On("GetRestaurantByName", "someErrorRestaurant").Return(nil, fmt.Errorf("some db error"))

		_, err := appHandler.fetchRestaurant("someErrorRestaurant")
		assert.Equal(t, ErrSystemError, err)

		mockRestaurantRepo.AssertExpectations(t)
	})

	t.Run("should fetch restaurant successfully", func(t *testing.T) {
		// Creating a mock restaurant to be returned by our mock repo.
		expectedRestaurant := &models.Restaurant{
			Name: "validRestaurant",
		}

		mockRestaurantRepo.On("GetRestaurantByName", "validRestaurant").Return(expectedRestaurant, nil)

		restaurant, err := appHandler.fetchRestaurant("validRestaurant")
		assert.NoError(t, err)
		assert.Equal(t, expectedRestaurant, restaurant)

		mockRestaurantRepo.AssertExpectations(t)
	})
}

func TestFetchMenuItems(t *testing.T) {
	var appHandler AppHandler
	var mockMenuItemRepo MockMenuItemRepository
	logger := logrus.New()

	appHandler.MenuItemRepo = &mockMenuItemRepo
	appHandler.Logger = logger

	tests := []struct {
		name                string
		input               string
		mockReturnMenuItems []*models.MenuItem
		mockReturnErr       error
		expectedErr         error
	}{
		{
			name:          "handle errors while fetching menu items",
			input:         "someErrorRestaurant",
			mockReturnErr: fmt.Errorf("some db error"),
			expectedErr:   ErrSystemError,
		},
		{
			name:  "fetch menu items successfully",
			input: "validRestaurant",
			mockReturnMenuItems: []*models.MenuItem{
				{Name: "Item1", Price: 10},
				{Name: "Item2", Price: 15},
			},
			expectedErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMenuItemRepo.On("GetMenuItemsByRestaurantName", tt.input).Return(tt.mockReturnMenuItems, tt.mockReturnErr)

			menuItems, err := appHandler.fetchMenuItems(tt.input)
			assert.Equal(t, tt.expectedErr, err)

			if tt.expectedErr == nil {
				assert.ElementsMatch(t, tt.mockReturnMenuItems, menuItems)
			}

			mockMenuItemRepo.AssertExpectations(t)
		})
	}
}

// ... And so on for other methods ...

// Mocked functions for order repository
// Add similar functions as above for the other mocked methods and repositories

// ... other mock functions ...

// func (m *MockTemplateHandler) generateFlexContainer(name string, args ...interface{}) (linebot.FlexContainer, error) {
// 	arguments := []interface{}{name}
// 	arguments = append(arguments, args...)
// 	args = m.Called(arguments...)
// 	if args.Get(0) == nil {
// 		return nil, args.Error(1)
// 	}
// 	return args.Get(0).(linebot.FlexContainer), args.Error(1)
// }
