package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	accessKey := os.Getenv("ACCESS_KEY")
	secretKey := os.Getenv("SECRET_KEY")
	bucketName := os.Getenv("BUCKET_NAME")
	region := os.Getenv("REGION")
	port := os.Getenv("PORT")
	apiKey := os.Getenv("API_KEY")

	if accessKey == "" || secretKey == "" || bucketName == "" || region == "" || port == "" || apiKey == "" {
		log.Fatal("Missing required envvars")
	}

	minioClient, err := minio.New(region+".digitaloceanspaces.com", &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: true,
	})
	if err != nil {
		log.Fatal("Error initializing DO Spaces client: %w", err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/delete", deleteHandler(minioClient, bucketName, apiKey)).Methods("POST")

	fmt.Printf("Server started on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func deleteHandler(minioClient *minio.Client, bucketName, apiKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// authenticate request
		providedApiKey := r.Header.Get("X-API-Key")
		if providedApiKey != apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// parse request body
		var data struct {
			ImagePath string `json:"image_path"`
		}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil || data.ImagePath == "" {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// delete image from Space
		ctx := context.Background()
		err = minioClient.RemoveObject(ctx, bucketName, data.ImagePath, minio.RemoveObjectOptions{})
		if err != nil {
			log.Printf("Error deleting object %s: %v", data.ImagePath, err)
			http.Error(w, "Error deleting image", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}
