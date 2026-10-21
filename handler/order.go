package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"microservices-in-golang/model"
	"microservices-in-golang/repository/order"
	"net/http"
	"reflect"
	"strconv"
	"time"
	"uuid"

	"github.com/go-chi/chi/v5"
)

type Order struct {
	repo *order.RedisRepo
}

type findAllPage struct {
	Offset uint64
	Size   int
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		customerID uuid.UUID        `json:"customer_id,omitempty"`
		LineItems  []model.LineItem `json:"line_items,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	now := time.Now().UTC()
	order := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.customerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}
	err := o.repo.Insert(r.Context(), order)
	if err != nil {
		fmt.Println("Failed to insert: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(order)
	if err != nil {
		fmt.Println("Failed to marshal: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
	w.WriteHeader(http.StatusCreated)
}
func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}
	const decimal = 10
	const bitsize = 64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitsize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	const size = 50
	findAll := reflect.ValueOf(o.repo).MethodByName("FindAll")
	page := reflect.New(findAll.Type().In(1)).Elem()
	page.FieldByName("Offset").SetUint(cursor)
	page.FieldByName("Size").SetInt(int64(size))
	results := findAll.Call([]reflect.Value{reflect.ValueOf(r.Context()), page})
	var res struct {
		Orders []model.Order
		Cursor uint64
	}
	pageResult := results[0]
	res.Orders = pageResult.FieldByName("Orders").Interface().([]model.Order)
	res.Cursor = pageResult.FieldByName("Cursor").Uint()
	err, _ = results[1].Interface().(error)
	if err != nil {
		fmt.Printf("Failed to fetch all: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var response struct // <-- this is an anonymous struct
	{
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}
	response.Items = res.Orders
	response.Next = res.Cursor
	data, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Failed to marshal: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	id_param := chi.URLParam(r, "id")
	const base = 10
	const bitsize = 64
	orderId, err := strconv.ParseUint(id_param, base, bitsize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	o, err = o.repo.FindByID(r.Context(), orderId)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find by Id: ", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err := json.NewEncoder(w).Encode(o); err != nil {
		fmt.Println("Failed to marshal: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitsize = 64
	order_id, err := strconv.ParseUint(idParam, base, bitsize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	theOrder, err := a.repo.FindByID(r.context(), order_id)
	if errors.Is(err, ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find by id: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	const completed_status = "completed"
	const shipped_status = "shipped"
	now := time.Now().UTC()

	switch body.Status {
	case shipped_status:
		if theOrder.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		theOrder.shipped_at = &now
	case completed_status:
		if theOrder.completed_at != nil || theOrder.shipped_at == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		theOrder.completed_at = &now
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := a.repo.Update(r.context(), theOrder)
	if err != nil {
		fmt.Println("Failed to insert: %w", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(theOrder); err != nil {
		fmt.Println("Failed to marshal: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	const base = 10
	const bitsize = 64
	orderid, err := strconv.ParseUint(idParam, base, bitsize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = a.repo.DeleteByID(r.Context(), orderid)
	if errors.Is(err, order.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("Failed to find by ID: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
