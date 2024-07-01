package utility

import (
	"fmt"
	"mime/multipart"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
)

var (
	awsSession *session.Session
	once       sync.Once
)

// GetAWSSession returns a singleton AWS session
func GetAWSSession() (*session.Session, error) {
	var err error
	once.Do(func() {
		awsSession, err = session.NewSession(&aws.Config{
			Region:      aws.String(os.Getenv("AWS_REGION")),
			Credentials: credentials.NewStaticCredentials(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), ""),
		})
	})
	return awsSession, err
}

func SendSMS(phoneNumber, message string) error {
	sess, err := GetAWSSession()
	if err != nil {
		return fmt.Errorf("!! failed to connect to AWS: %w", err)
	}

	svc := sns.New(sess)
	params := &sns.PublishInput{
		PhoneNumber: aws.String(phoneNumber),
		Message:     aws.String(message),
	}

	// sends a text message (SMS message) directly to a phone number.
	resp, err := svc.Publish(params)

	if err != nil {
		return fmt.Errorf("!! failed to send SMS: %w", err)
	}

	fmt.Println(resp)

	return nil
}

func UploadFileToS3(fileHeader *multipart.FileHeader, s3Path, bucketName string) (string, error) {
	sess, err := GetAWSSession()
	if err != nil {
		return "", fmt.Errorf("!! failed to connect to AWS: %w", err)
	}

	svc := s3.New(sess)

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("!! error opening file: %v", err)
	}
	defer file.Close()

	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(s3Path),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("!! error uploading file: %v", err)
	}

	s3URL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, s3Path)
	return s3URL, nil
}
