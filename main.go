package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/minio/minio-go"
)

func main() {
	endpoint := os.Getenv("MINIO_URL") + ":" + os.Getenv("MINIO_PORT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := false

	dirs := []string{"tv", "movies"}
	bucketLocation := "rapture"

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}

	// Make the complete bucket
	err = minioClient.MakeBucket("complete", bucketLocation)
	if err != nil {
		// Check to see if we already own this bucket
		exists, err := minioClient.BucketExists("complete")
		if err == nil && exists {
			log.Printf("We already own %s\n", "complete")
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", "complete")
	}

	for _, dir := range dirs {
		file := dir + "/.keep"
		localFile := os.Getenv("PWD") + "/.keep"
		err = ioutil.WriteFile(localFile, []byte("Placeholder"), 0644)
		if err != nil {
			log.Fatalln(err)
		}
		_, err := minioClient.FPutObject("complete", file, localFile, minio.PutObjectOptions{})
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Successfully added .keep to complete bucket directories.")
		os.Remove(localFile)
	}

	// Make the transcode bucket
	err = minioClient.MakeBucket("transcode", bucketLocation)
	if err != nil {
		// Check to see if we already own this bucket
		exists, err := minioClient.BucketExists("transcode")
		if err == nil && exists {
			log.Printf("We already own %s\n", "transcode")
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", "transcode")
	}

	// Set notification configs
	sqsArn := minio.NewArn("minio", "sqs", "rapture", "1", "webhook")
	sqsConfig := minio.NewNotificationConfig(sqsArn)
	sqsConfig.AddEvents(minio.ObjectCreatedAll)

	// Set notification configs
	bucketNotification := minio.BucketNotification{}
	bucketNotification.AddQueue(sqsConfig)

	// Add notifications to the transcode bucket
	err = minioClient.SetBucketNotification("transcode", bucketNotification)
	if err != nil {
		log.Fatalln("Error: " + err.Error())
	}
	log.Println("Successfully added bucket notification")

	// Read the files in the dir slice
	for _, dir := range dirs {
		bucketDirPrefix := strings.Replace(dir, "/", "", 1)
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			log.Fatal(err)
		}

		// Set new spinner
		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)

		// Upload each file to the trasncode bucket
		for _, f := range files {

			filePath := dir + "/" + f.Name()
			objectName := bucketDirPrefix + "/" + f.Name()

			s.Suffix = " Uploading ==> " + f.Name()
			s.Start()

			// Upload the file to the bucket
			_, err := minioClient.FPutObject("transcode", objectName, filePath, minio.PutObjectOptions{})
			if err != nil {
				log.Fatalln(err)
			}

			s.FinalMSG = "Successfully uploaded ==> " + fmt.Sprintf("%s\n", f.Name())
			os.Remove(filePath)
			s.Stop()
		}
	}
}
