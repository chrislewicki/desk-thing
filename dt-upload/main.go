package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/image/bmp"
)

var minioClient *minio.Client
var bucketName string

func init() {
	// MinIO init
	bucketName := os.Getenv("BUCKET_NAME")
	endpoint := os.Getenv("ENDPOINT")
	accessKeyID := os.Getenv("ACCESS_KEY")
	secretAccessKey := os.Getenv("SECRET_KEY")
	useSSL := true

	var err error
	minioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		fmt.Println("Error initializing Spaces client:", err)
		os.Exit(1)
	}
}

func main() {
	important := flag.Bool("important", false, "Mark image as important")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Usage: dt-upload [--important] <image_path_or_directory")
		os.Exit(1)
	}

	for _, path := range args {
		err := processPath(path, *important)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", path, err)
		}
	}
}

func processPath(path string, important bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return processDirectory(path, important)
	} else {
		return processFile(path, important)
	}
}

func processDirectory(dirPath string, important bool) error {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if isImageFile(info.Name()) {
			return processFile(path, important)
		}
		return nil
	})
	return err
}

func processFile(filePath string, important bool) error {
	img, err := imaging.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	// Resize image to 320x240 for PyPortal (Pynt)
	// PyPortal Titano has res of 480x320
	resizedImg := imaging.Fill(img, 320, 240, imaging.Center, imaging.Lanczos)

	// Convert image to bmp
	outputPath := getOutputPath(filePath, important)
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}
	defer outputFile.Close()

	rgbaImg := image.NewRGBA(resizedImg.Bounds())
	draw.Draw(rgbaImg, resizedImg.Bounds(), resizedImg, image.Point{}, draw.Src)

	err = bmp.Encode(outputFile, rgbaImg)
	if err != nil {
		return fmt.Errorf("failed to encode bmp: %w", err)
	}

	err = uploadImage(outputPath, important)
	if err != nil {
		return fmt.Errorf("failed to upload image: %w", err)
	}

	os.Remove(outputPath)

	fmt.Printf("Processed and uploaded: %s\n", outputPath)
	return nil
}

func getOutputPath(inputPath string, important bool) string {
	filename := filepath.Base(inputPath)
	filename = strings.TrimSuffix(filename, filepath.Ext(filename)) + ".bmp"
	if important {
		return filepath.Join(os.TempDir(), "important_"+filename)
	}
	return filepath.Join(os.TempDir(), filename)
}

func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".bmp", "gif":
		return true
	default:
		return false
	}
}

func uploadImage(filePath string, important bool) error {
	objectName := filepath.Base(filePath)
	if important {
		objectName = "important/" + objectName
	} else {
		objectName = "images/" + objectName
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	_, err = minioClient.PutObject(
		context.Background(),
		bucketName,
		objectName,
		file,
		fileInfo.Size(),
		minio.PutObjectOptions{ContentType: "image/bmp"},
	)
	if err != nil {
		return err
	}

	return nil
}
