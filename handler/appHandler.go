package handler

import (
	"github.com/JohnsonYuanTW/NCAEats/models"
	"github.com/line/line-bot-sdk-go/v7/linebot"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type DB struct {
	Connection *gorm.DB
}

type AppHandler struct {
	Logger          *logrus.Logger
	Env             map[string]string
	Bot             *linebot.Client
	MenuItemRepo    *models.MenuItemRepository
	OrderRepo       *models.OrderRepository
	OrderDetailRepo *models.OrderDetailRepository
	RestaurantRepo  *models.RestaurantRepository
}

func NewAppHandler(log *logrus.Logger, env map[string]string, bot *linebot.Client, db *gorm.DB) (*AppHandler, error) {
	// AppHandler Creation
	baseRepo := &models.BaseRepository{
		DB: db,
	}
	appHandler := &AppHandler{
		Logger: log,
		Env:    env,
		Bot:    bot,
		MenuItemRepo: &models.MenuItemRepository{
			BaseRepository: baseRepo,
		},
		OrderRepo: &models.OrderRepository{
			BaseRepository: baseRepo,
		},
		OrderDetailRepo: &models.OrderDetailRepository{
			BaseRepository: baseRepo,
		},
		RestaurantRepo: &models.RestaurantRepository{
			BaseRepository: baseRepo,
		},
	}

	if err := appHandler.initRepository(); err != nil {
		return nil, err
	}

	return appHandler, nil
}

func (a *AppHandler) initRepository() error {
	initRepos := []struct {
		Repo interface {
			Init() error
		}
		Name string
	}{
		{Repo: a.MenuItemRepo, Name: "MenuItem"},
		{Repo: a.OrderRepo, Name: "Order"},
		{Repo: a.OrderDetailRepo, Name: "OrderDetail"},
		{Repo: a.RestaurantRepo, Name: "Restaurant"},
	}

	for _, initRepo := range initRepos {
		if err := initRepo.Repo.Init(); err != nil {
			a.Logger.WithError(err).Fatalf("Failed to initialize %s", initRepo.Name)
			return err
		}
	}
	return nil
}
