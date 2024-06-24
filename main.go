package main

import (
    "context"
    "encoding/json"
    "html/template"
    "log"
    "net/http"
    "path/filepath"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Order struct {
    Name         string `bson:"name" json:"name"`
    PhoneNumber  string `bson:"phone_number" json:"phone_number"`
    Address      string `bson:"address" json:"address"`
    DeliveryTime string `bson:"delivery_time" json:"delivery_time"`
}

var client *mongo.Client

func main() {
    // Connect to MongoDB
    clientOptions := options.Client().ApplyURI("mongodb+srv://prachhhi:oprybBJBWko7zbjE@cluster0.r487mib.mongodb.net/?retryWrites=true&w=majority")
    var err error
    client, err = mongo.Connect(context.Background(), clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    defer func() {
        if err = client.Disconnect(context.Background()); err != nil {
            log.Fatal(err)
        }
    }()

    http.HandleFunc("/", handleFormSubmission)
    http.HandleFunc("/map", handleMapDisplay)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleFormSubmission(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodPost {
        var order Order
        order.Name = r.FormValue("name")
        order.PhoneNumber = r.FormValue("phone_number")
        order.Address = r.FormValue("address")
        order.DeliveryTime = r.FormValue("delivery_time")

        collection := client.Database("order_locator").Collection("orders")
        _, err := collection.InsertOne(context.Background(), order)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        http.Redirect(w, r, "/map", http.StatusSeeOther)
    } else {
        tmpl, _ := template.ParseFiles(filepath.Join("templates", "form.html"))
        tmpl.Execute(w, nil)
    }
}

func handleMapDisplay(w http.ResponseWriter, r *http.Request) {
    collection := client.Database("order_locator").Collection("orders")
    cursor, err := collection.Find(context.Background(), bson.M{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    var orders []Order
    if err = cursor.All(context.Background(), &orders); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    data, err := json.Marshal(orders)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    tmpl, _ := template.ParseFiles(filepath.Join("templates", "map.html"))
    tmpl.Execute(w, string(data))
}
