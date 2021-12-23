package app

import (
	"context"
	"fmt"
	"goweb/db"
	"goweb/models"
	"log"
	"net/http"
	"os"
	"path"
	"text/template"
	"time"

	"github.com/gorilla/schema"
	//"go.mongodb.org/mongo-driver/bson"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"

	//"github.com/aws/aws-sdk-go/aws/session"
	//"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

var (
	collection = db.ConnectDB()
	// simple html files for viewing
	tpl        = template.Must(template.ParseFiles("assests/index.html"))
	tpl2       = template.Must(template.ParseFiles("assests/upload.html"))
	tpl3       = template.Must(template.ParseFiles("assests/creds.html"))

	decoder    = schema.NewDecoder()

	awsS3Client *s3.Client

	access_key  string
	secret_key  string
	session_key string
)

func mapUrls() {
	router.HandleFunc("/", indexHandler).Methods("GET")
	router.HandleFunc("/api/temp", redirectHandler).Methods("GET") // temporary redirect
	router.HandleFunc("/api/start", getCred).Methods("GET")
	router.HandleFunc("/api/upload", uploadHandler).Methods("POST")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tpl.Execute(w, nil)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusFound)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var videoData models.VideoMetaData //videoData.Description = r.PostFormValue("description")
	err = decoder.Decode(&videoData, r.PostForm)
	if err != nil {
		return
	}
	videoData.Key = uuid.New().String()

	file, header, err := r.FormFile("videofile")
	if err != nil {
		log.Fatal(err)
		fmt.Println("error when trying to read file")
		return
	}
	defer file.Close()

	filename := header.Filename

	configS3() // return awsS3Client with credentials populated
	// TODO add multipart?
	uploader := manager.NewUploader(awsS3Client)
	_, err = uploader.Upload(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("AWS_S3_BUCKET")),
		Key:    aws.String(path.Join("input", videoData.Key, "/", filename)),
		Body:   file,
	})
	
	if err != nil {
		http.Error(w, "error while uploading", http.StatusInternalServerError)
		fmt.Println(err)
		return
	} else {
		_, err = collection.InsertOne(context.TODO(), videoData)
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
	}

	//TODO sqs here
	//sendToSQS(context.TODO(), videoData)

	tpl2.Execute(w, nil)
	//http.Redirect(w, r, "/api/temp", http.StatusFound)
}

func getCred(w http.ResponseWriter, r *http.Request) {

	creds := credentials.NewStaticCredentialsProvider(
		os.Getenv("_access_key"),
		os.Getenv("_secret_key"),
		"")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_S3_REGION")),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		fmt.Println("error while sts config")
		log.Fatal(err)
	}

	// using sts to get temp credentials => kcred
	stsClient := sts.NewFromConfig(cfg)
	provider := stscreds.NewAssumeRoleProvider(stsClient, os.Getenv("roleARN"))
	cfg.Credentials = aws.NewCredentialsCache(provider)

	kcreds, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		fmt.Println("error while sts cred set")
		log.Fatal(err)
	}

	fmt.Println("Token expires in ", time.Time((kcreds.Expires)).Local())

	access_key = kcreds.AccessKeyID
	secret_key = kcreds.SecretAccessKey
	session_key = kcreds.SessionToken

	tpl3.Execute(w, nil)
}

func configS3() {

	// give access to S3 using temp creds from getCreds()
	// or can use IAM creds => NOT RECOMMENDED 
	creds := credentials.NewStaticCredentialsProvider(access_key, secret_key, session_key)

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(creds),
		config.WithRegion(os.Getenv("AWS_S3_REGION")),
	)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	awsS3Client = s3.NewFromConfig(cfg)
}
