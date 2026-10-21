// You don't need pointers (*[]int) unless you plan to modify the
// slice's length or capacity inside the function.
// slice contains pointer, length and capacity

// Atomic transactions = all of them will work together or none of them will.
package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"microservices-in-golang/model"

	"github.com/go-redis/redis"
)

type RedisRepo struct {
	Client *redis.Client
}

func orderIDKey(id uint64) string {
	return fmt.Sprintf("order: %d", id)
}

func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order) // load/dump (in python) = marshall/unmarshall (in go)
	if err != nil {
		return fmt.Errorf("Unable to encode order: %w", err)
	}
	key := orderIDKey(order.OrderID)
	txn := r.Client.TxPipeline()

	res := txn.SetNX(key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("Failed to set order: %w", err)
	}
	if err := txn.SAdd("orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("Failed to add to orders set: %w", err)
	}
	if _, err := txn.Exec(); err != nil {
		return fmt.Errorf("Failed to execute: %w", err)
	}
	return nil
}

var ErrNotExist = errors.New("The order doesn't exist") //use var instead of go-walrus
// outside function definitions
func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIDKey(id)
	value, err := r.Client.Get(key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("Order get: %w", err)
	}

	var order model.Order //use var when you need to initialize an attribute, instead of an instance
	err = json.Unmarshal([]byte(value), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("Failed to decode order json: %w", err)
	}
	return order, nil
}

func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)
	txn := r.Client.TxPipeline()
	err := txn.Del(key).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("Got order: %w", err)
	}
	return nil
}

func (r *RedisRepo) UpdateByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)
	err := r.Client.SetXX(key, "", 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("The ID doesn't exist: %w", err)
	}
	return nil
}

type findAllPage struct {
	size   uint
	Offset uint
}
type FindRes struct {
	Orders []model.Order
	Cursor uint64
}

func (r *RedisRepo) FindAll(ctx context.Context, page findAllPage) (FindRes, error) {
	res := r.Client.SScan("orders", uint64(page.Offset), "*", int64(page.size))
	keys, cursor, err := res.Result()

	if err != nil {
		return FindRes{}, fmt.Errorf("Failed to get order IDs: %w", err)
	}
	if len(keys) == 0 {
		return FindRes{
			Orders: []model.Order{},
		}, nil
	}
	//func(...) - signifies variadic arguments
	xs, err := r.Client.MGet(keys...).Result()
	if err != nil {
		return FindRes{}, fmt.Errorf("Failed to get orders: %w", err)
	}
	orders := make([]model.Order, len(xs))
	for i, x := range xs {
		x, ok := x.(string)
		if !ok {
			return FindRes{}, fmt.Errorf("Failed to decode order json: invalid value type")
		}
		var order model.Order
		err := json.Unmarshal([]byte(x), &order)
		if err != nil {
			return FindRes{}, fmt.Errorf("Failed to decode order json: %w", err)
		}
		orders[i] = order
	}
	return FindRes{
		Orders: orders,
		Cursor: cursor,
	}, nil
}
