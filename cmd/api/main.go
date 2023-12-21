package main

import (
	"context"
	"encoding/json"
	"fmt"
	"geotask/cache"
	"geotask/module/courier/models"
	_ "geotask/module/courier/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	httpSwagger "github.com/swaggo/http-swagger"
	"myRedis/model"
	"net/http"
)

const (
	redisKeyCourier = "courier"
	redisKeyOrders  = "orders"
	CourierRadius   = 2500
)

// @title courier service
// @version 1.0
// @description courier service
// @host localhost:8080
// @BasePath /api/v1
//
//go:generate swagger generate spec -o ../public/swagger.json --scan-models
func main() {
	router := gin.Default()

	rdb := cache.NewRedisClient("localhost", ":6379")

	http.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/docs/doc.json"),
	))

	router.POST("/move-courier", func(c *gin.Context) {
		var courierLocation models.Point
		if err := c.BindJSON(&courierLocation); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := setCourierStatus(rdb, courierLocation)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Courier location updated"})
	})

	router.GET("/get-status", func(c *gin.Context) {
		courierLocation, err := getCourierStatus(rdb)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		orders, err := getOrdersInRadius(rdb, courierLocation, CourierRadius)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		response := gin.H{
			"courier": courierLocation,
			"orders":  orders,
		}

		c.JSON(http.StatusOK, response)
	})

	router.Run(":8080")
}

func setCourierStatus(rdb *redis.Client, courierLocation models.Point) error {
	ctx := context.Background()
	data, err := json.Marshal(courierLocation)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, redisKeyCourier, data, 0).Err()
}

// @Summary Получить информацию о курьере
// @Description Получить текущее местоположение курьера
// @ID getCourierStatus
// @Accept  json
// @Produce  json
// @Success 200 {object} models.Point
// @Router /courier/status [get]
func getCourierStatus(rdb *redis.Client) (models.Point, error) {
	ctx := context.Background()
	val, err := rdb.Get(ctx, redisKeyCourier).Result()
	if err != nil {
		return models.Point{}, err
	}

	var courierLocation models.Point
	err = json.Unmarshal([]byte(val), &courierLocation)
	if err != nil {
		return models.Point{}, err
	}

	return courierLocation, nil
}

func getOrdersInRadius(rdb *redis.Client, loc models.Point, radius float64) ([]model.Order, error) {
	ctx := context.Background()

	val, err := rdb.Get(ctx, redisKeyOrders).Result()
	if err != nil {
		return nil, err
	}

	var orders []model.Order
	err = json.Unmarshal([]byte(val), &orders)
	if err != nil {
		return nil, err
	}

	ordersInRadius, err := rdb.GeoRadius(ctx, redisKeyOrders, loc.Lat, loc.Lng, &redis.GeoRadiusQuery{
		Radius: radius,
		Unit:   "km",
	}).Result()
	if err != nil {
		fmt.Println("Ошибка при получении заказов в радиусе:", err)
		return nil, err
	}

	var ordersResult []model.Order

	for _, orderLocation := range ordersInRadius {
		var order model.Order

		err := json.Unmarshal([]byte(orderLocation.Name), &order)
		if err != nil {
			fmt.Println("Ошибка при декодировании JSON:", err)
			continue
		}

		ordersResult = append(ordersResult, order)
	}

	return ordersResult, nil
}
