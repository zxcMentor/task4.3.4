package main

import (
	"context"
	"encoding/json"
	"geotask/module/courier/models"
	models2 "geotask/module/order/models"
	"github.com/redis/go-redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSetCourierStatus(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	courierLocation := models.Point{
		Lat: 40.7128,
		Lng: -74.0060,
	}

	err := setCourierStatus(rdb, courierLocation)

	assert.NoError(t, err)

	ctx := context.Background()
	result, err := rdb.Get(ctx, redisKeyCourier).Result()

	assert.NoError(t, err)

	var storedLocation models.Point
	err = json.Unmarshal([]byte(result), &storedLocation)

	assert.NoError(t, err)

	assert.Equal(t, courierLocation, storedLocation)
}

func TestGetCourierStatus(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	expectedLocation := models.Point{
		Lat: 40.7128,
		Lng: -74.0060,
	}

	data, err := json.Marshal(expectedLocation)
	assert.NoError(t, err)
	err = rdb.Set(context.Background(), redisKeyCourier, data, 0).Err()
	assert.NoError(t, err)

	actualLocation, err := getCourierStatus(rdb)

	assert.NoError(t, err)

	assert.Equal(t, expectedLocation, actualLocation)
}

func TestGetOrdersInRadius(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	courierLocation := models.Point{
		Lat: 40.7128,
		Lng: -74.0060,
	}

	fakeOrders := []models2.Order{
		{
			ID:            1,
			Price:         25.99,
			DeliveryPrice: 5.0,
			Lng:           -74.0060,
			Lat:           40.7128,
			IsDelivered:   false,
			CreatedAt:     time.Now(),
		},
		{
			ID:            2,
			Price:         19.99,
			DeliveryPrice: 3.0,
			Lng:           -74.0065,
			Lat:           40.7228,
			IsDelivered:   false,
			CreatedAt:     time.Now(),
		},
	}

	data, err := json.Marshal(fakeOrders)
	assert.NoError(t, err)
	err = rdb.Set(context.Background(), redisKeyOrders, data, 0).Err()
	assert.NoError(t, err)

	radius := 5.0
	ordersInRadius, err := getOrdersInRadius(rdb, courierLocation, radius)

	assert.NoError(t, err)

	expectedOrders := []models2.Order{
		{
			ID:            1,
			Price:         25.99,
			DeliveryPrice: 5.0,
			Lng:           -74.0060,
			Lat:           40.7128,
			IsDelivered:   false,
			CreatedAt:     time.Now(),
		},
	}

	assert.ElementsMatch(t, expectedOrders, ordersInRadius)
}
