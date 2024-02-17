package object // Special thanks via https://docs.digitalocean.com/products/spaces/resources/s3-sdk-examples/

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"

	c "github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/config"
	"github.com/Pharo-Non-Profit/nonprofitvault-backend/internal/provider/uuid"
)

type ObjectStorager interface {
	GenerateSSEC() (key, md5Hash string, err error)
	UploadContent(ctx context.Context, objectKey string, content []byte) error
	UploadContentFromMulipart(ctx context.Context, objectKey string, file multipart.File) error
	BucketExists(ctx context.Context, bucketName string) (bool, error)
	GetDownloadablePresignedURL(ctx context.Context, key string, duration time.Duration) (string, error)
	GetPresignedURL(ctx context.Context, key string, duration time.Duration) (string, error)
	DeleteByKeys(ctx context.Context, key []string) error
	Cut(ctx context.Context, sourceObjectKey string, destinationObjectKey string) error
	Copy(ctx context.Context, sourceObjectKey string, destinationObjectKey string) error
	GetBinaryData(ctx context.Context, objectKey string) (io.ReadCloser, error)
	DownloadToLocalfile(ctx context.Context, objectKey string, filePath string) (string, error)
	ListAllObjects(ctx context.Context) (*s3.ListObjectsOutput, error)
	FindMatchingObjectKey(s3Objects *s3.ListObjectsOutput, partialKey string) string
}

type objectStorager struct {
	S3Client              *s3.Client
	PresignClient         *s3.PresignClient
	UUID                  uuid.Provider
	Logger                *slog.Logger
	BucketName            string
	SSECustomerKey        string
	SSECustomerKeyMd5Hash string
}

// NewStorage connects to a specific S3 bucket instance and returns a connected
// instance structure.
func NewStorage(appConf *c.Conf, logger *slog.Logger, uuidp uuid.Provider) ObjectStorager {
	// DEVELOPERS NOTE:
	// How can I use the AWS SDK v2 for Go with DigitalOcean Spaces? via https://stackoverflow.com/a/74284205
	logger.Debug("object storage initializing...")

	// STEP 1: initialize the custom `endpoint` we will connect to.
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: appConf.AWS.Endpoint,
		}, nil
	})

	// STEP 2: Configure.
	sdkConfig, err := config.LoadDefaultConfig(
		context.TODO(), config.WithRegion(appConf.AWS.Region),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(appConf.AWS.AccessKey, appConf.AWS.SecretKey, "")),
	)
	if err != nil {
		log.Fatal(err) // We need to crash the program at start to satisfy google wire requirement of having no errors.
	}

	// STEP 3\: Load up s3 instance.
	s3Client := s3.NewFromConfig(sdkConfig)

	// For debugging purposes only.
	logger.Debug("object storage connected to remote service")

	// Create our storage handler.
	s3Storage := &objectStorager{
		S3Client:              s3Client,
		PresignClient:         s3.NewPresignClient(s3Client),
		Logger:                logger,
		UUID:                  uuidp,
		BucketName:            appConf.AWS.BucketName,
		SSECustomerKey:        appConf.AWS.SSECustomerKey,
		SSECustomerKeyMd5Hash: calculateMD5Hash(appConf.AWS.SSECustomerKey),
	}

	// STEP 4: Connect to the s3 bucket instance and confirm that bucket exists.
	doesExist, err := s3Storage.BucketExists(context.TODO(), appConf.AWS.BucketName)
	if err != nil {
		log.Fatal(err) // We need to crash the program at start to satisfy google wire requirement of having no errors.
	}
	if !doesExist {
		log.Fatal("bucket name does not exist") // We need to crash the program at start to satisfy google wire requirement of having no errors.
	}

	// For debugging purposes only.
	// x, y, _ := s3Storage.GenerateSSEC()
	// log.Println(x)
	// log.Println(y)

	// For debugging purposes only.
	logger.Debug("object storage ready")

	// Return our s3 storage handler.
	return s3Storage
}

// GenerateSSEC generates a random SSE-C key and its MD5 hash. Use this function
// to generate your key which you can use in this adapter.
func (s *objectStorager) GenerateSSEC() (key, md5Hash string, err error) {
	// Generate a random SSE-C key (32 bytes).
	keyBytes := make([]byte, 32)
	_, err = rand.Read(keyBytes)
	if err != nil {
		return "", "", err
	}

	// Base64 encode the SSE-C key.
	key = base64.StdEncoding.EncodeToString(keyBytes)

	// Calculate the MD5 hash of the SSE-C key.
	hashBytes := md5.Sum(keyBytes)
	md5Hash = fmt.Sprintf("%x", hashBytes)

	return key, md5Hash, nil
}

func (s *objectStorager) UploadContent(ctx context.Context, objectKey string, content []byte) error {
	params := &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
		Body:   bytes.NewReader(content),
	}

	// The following block of cdode will attach server side encryption if specified.
	if s.SSECustomerKey != "" {
		// Attach the server side encryption with customer key.
		params.ServerSideEncryption = types.ServerSideEncryptionAwsKms
		params.SSECustomerAlgorithm = aws.String("AES256") // SSE-C encryption algorithm
		params.SSECustomerKey = &s.SSECustomerKey
		params.SSECustomerKeyMD5 = &s.SSECustomerKeyMd5Hash
	}

	_, err := s.S3Client.PutObject(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (s *objectStorager) UploadContentFromMulipart(ctx context.Context, objectKey string, file multipart.File) error {
	// Create the S3 upload input parameters
	params := &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	}

	// The following block of code will attach server side encryption if specified.
	if s.SSECustomerKey != "" {
		// Attach the server side encryption with customer key.
		params.ServerSideEncryption = types.ServerSideEncryptionAwsKms
		params.SSECustomerAlgorithm = aws.String("AES256") // SSE-C encryption algorithm
		params.SSECustomerKey = &s.SSECustomerKey
		params.SSECustomerKeyMD5 = &s.SSECustomerKeyMd5Hash
	}

	// Perform the file upload to S3
	_, err := s.S3Client.PutObject(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (s *objectStorager) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	// Note: https://docs.aws.amazon.com/code-library/latest/ug/go_2_s3_code_examples.html#actions

	_, err := s.S3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	exists := true
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			switch apiError.(type) {
			case *types.NotFound:
				log.Printf("Bucket %v is available.\n", bucketName)
				exists = false
				err = nil
			default:
				log.Printf("Either you don't have access to bucket %v or another error occurred. "+
					"Here's what happened: %v\n", bucketName, err)
			}
		}
	}

	return exists, err
}

func (s *objectStorager) GetDownloadablePresignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	// DEVELOPERS NOTE:
	// AWS S3 Bucket — presigned URL APIs with Go (2022) via https://ronen-niv.medium.com/aws-s3-handling-presigned-urls-2718ab247d57

	params := &s3.GetObjectInput{
		Bucket:                     aws.String(s.BucketName),
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String("attachment"), // This field allows the file to download it directly from your browser
	}

	// The following block of cdode will attach server side encryption if specified.
	if s.SSECustomerKey != "" {
		params.SSECustomerAlgorithm = aws.String("AES256") // SSE-C encryption algorithm
		params.SSECustomerKey = &s.SSECustomerKey
		params.SSECustomerKeyMD5 = &s.SSECustomerKeyMd5Hash
	}

	presignedUrl, err := s.PresignClient.PresignGetObject(context.Background(),
		params,
		s3.WithPresignExpires(duration))
	if err != nil {
		return "", err
	}
	return presignedUrl.URL, nil
}

func (s *objectStorager) GetPresignedURL(ctx context.Context, objectKey string, duration time.Duration) (string, error) {
	// DEVELOPERS NOTE:
	// AWS S3 Bucket — presigned URL APIs with Go (2022) via https://ronen-niv.medium.com/aws-s3-handling-presigned-urls-2718ab247d57

	params := &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
	}

	// The following block of cdode will attach server side encryption if specified.
	if s.SSECustomerKey != "" {
		params.SSECustomerAlgorithm = aws.String("AES256") // SSE-C encryption algorithm
		params.SSECustomerKey = &s.SSECustomerKey
		params.SSECustomerKeyMD5 = &s.SSECustomerKeyMd5Hash
	}

	presignedUrl, err := s.PresignClient.PresignGetObject(context.Background(),
		params,
		s3.WithPresignExpires(duration))
	if err != nil {
		return "", err
	}
	return presignedUrl.URL, nil
}

func (s *objectStorager) DeleteByKeys(ctx context.Context, objectKeys []string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	var objectIds []types.ObjectIdentifier
	for _, key := range objectKeys {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(key)})
	}
	_, err := s.S3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.BucketName),
		Delete: &types.Delete{Objects: objectIds},
	})
	if err != nil {
		log.Printf("Couldn't delete objects from bucket %v. Here's why: %v\n", s.BucketName, err)
	}
	return err
}

func (s *objectStorager) Cut(ctx context.Context, sourceObjectKey string, destinationObjectKey string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second) // Increase timout so it runs longer then usual to handle this unique case.
	defer cancel()

	params := &s3.CopyObjectInput{
		Bucket:     aws.String(s.BucketName),
		CopySource: aws.String(s.BucketName + "/" + sourceObjectKey),
		Key:        aws.String(destinationObjectKey),
	}

	// The following block of cdode will attach server side encryption if specified.
	if s.SSECustomerKey != "" {
		params.SSECustomerAlgorithm = aws.String("AES256") // SSE-C encryption algorithm
		params.SSECustomerKey = &s.SSECustomerKey
		params.SSECustomerKeyMD5 = &s.SSECustomerKeyMd5Hash
	}

	_, copyErr := s.S3Client.CopyObject(ctx, params)
	if copyErr != nil {
		s.Logger.Error("Failed to copy object:", slog.Any("copyErr", copyErr))
		return copyErr
	}

	s.Logger.Debug("Object copied successfully.")

	// Delete the original object
	_, deleteErr := s.S3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(sourceObjectKey),
	})
	if deleteErr != nil {
		s.Logger.Error("Failed to delete original object:", slog.Any("deleteErr", deleteErr))
		return deleteErr
	}

	s.Logger.Debug("Original object deleted.")

	return nil
}

func (s *objectStorager) Copy(ctx context.Context, sourceObjectKey string, destinationObjectKey string) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second) // Increase timout so it runs longer then usual to handle this unique case.
	defer cancel()

	params := &s3.CopyObjectInput{
		Bucket:     aws.String(s.BucketName),
		CopySource: aws.String(s.BucketName + "/" + sourceObjectKey),
		Key:        aws.String(destinationObjectKey),
	}

	// The following block of cdode will attach server side encryption if specified.
	if s.SSECustomerKey != "" {
		params.SSECustomerAlgorithm = aws.String("AES256") // SSE-C encryption algorithm
		params.SSECustomerKey = &s.SSECustomerKey
		params.SSECustomerKeyMD5 = &s.SSECustomerKeyMd5Hash
	}

	_, copyErr := s.S3Client.CopyObject(ctx, params)
	if copyErr != nil {
		s.Logger.Error("Failed to copy object:", slog.Any("copyErr", copyErr))
		return copyErr
	}

	s.Logger.Debug("Object copied successfully.")

	return nil
}

// GetBinaryData function will return the binary data for the particular key.
func (s *objectStorager) GetBinaryData(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(objectKey),
	}

	// The following block of cdode will attach server side encryption if specified.
	if s.SSECustomerKey != "" {
		params.SSECustomerAlgorithm = aws.String("AES256") // SSE-C encryption algorithm
		params.SSECustomerKey = &s.SSECustomerKey
		params.SSECustomerKeyMD5 = &s.SSECustomerKeyMd5Hash
	}

	s3object, err := s.S3Client.GetObject(ctx, params)
	if err != nil {
		return nil, err
	}
	return s3object.Body, nil
}

func (s *objectStorager) DownloadToLocalfile(ctx context.Context, objectKey string, filePath string) (string, error) {
	responseBin, err := s.GetBinaryData(ctx, objectKey)
	if err != nil {
		return filePath, err
	}
	out, err := os.Create(filePath)
	if err != nil {
		return filePath, err
	}
	defer out.Close()

	_, err = io.Copy(out, responseBin)
	if err != nil {
		return "", err
	}
	return filePath, err
}

func (s *objectStorager) ListAllObjects(ctx context.Context) (*s3.ListObjectsOutput, error) {
	params := &s3.ListObjectsInput{
		Bucket: aws.String(s.BucketName),
	}

	objects, err := s.S3Client.ListObjects(ctx, params)
	if err != nil {
		return nil, err
	}

	return objects, nil
}

// Function will iterate over all the s3 objects to match the partial key with
// the actual key found in the S3 bucket.
func (s *objectStorager) FindMatchingObjectKey(s3Objects *s3.ListObjectsOutput, partialKey string) string {
	for _, obj := range s3Objects.Contents {

		match := strings.Contains(*obj.Key, partialKey)

		// If a match happens then it means we have found the ACTUAL KEY in the
		// s3 objects inside the bucket.
		if match == true {
			return *obj.Key
		}
	}
	return ""
}

// calculateMD5Hash function to calculate MD5 hash of a byte slice
func calculateMD5Hash(ssecKey string) string {
	rawKey, err := base64.StdEncoding.DecodeString(ssecKey)
	if err != nil {
		log.Fatalf("[ERROR] decoding ssecKey: %s\n", err)
	}
	hasher := md5.New()
	hasher.Write(rawKey)
	keyHashB64 := base64.StdEncoding.EncodeToString(hasher.Sum(nil))
	return keyHashB64
}
