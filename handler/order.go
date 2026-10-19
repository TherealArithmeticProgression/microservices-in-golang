package handler

import (
	"fmt"
	"net/http"
)

type Order struct {
}

func (o *Order) Create(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Creating an order.")
}
func (o *Order) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Listing out all orders: ")
}
func (o *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got an order by it's ID")
}
func (o *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Update an order by it's ID")
}
func (o *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Delete an order by it's ID.")
}
