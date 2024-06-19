package utility

import (
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	AWS_S3_REGION = "ap-south-1"
	AWS_S3_BUCKET = "vraj-s-bucket"
)

var sess = connectAWS()

func connectAWS() *session.Session {
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(AWS_S3_REGION),
		},
	)
	if err != nil {
		panic(err)
	}
	return sess
}

// UploadFile uploads a file to AWS S3.
func UploadFileToS3(file multipart.File, fileHeader *multipart.FileHeader, bucketName string) error {
	defer file.Close()
	uploader := s3manager.NewUploader(sess)

	uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(AWS_S3_BUCKET),
		Key:    aws.String(fileHeader.Filename),
		Body:   file,
	})

	// cfg, err := aws.NewConfig.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to load SDK config, %v", err)
	// }

	// client := s3.NewFromConfig(cfg)

	// file, err := fileHeader.Open()
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()

	// fileBytes := bytes.NewReader(nil)
	// if _, err := fileBytes.ReadFrom(file); err != nil {
	// 	return err
	// }

	// sess, err := session.NewSession(&aws.Config{
	// 	Region: aws.String("your-aws-region"),
	// })
	// if err != nil {
	// 	return err
	// }

	// svc := s3.New(sess)
	// _, err = svc.PutObject(&s3.PutObjectInput{
	// 	Bucket:      aws.String(bucketName),
	// 	Key:         aws.String(fileHeader.Filename),
	// 	Body:        fileBytes,
	// 	ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
	// })
	// if err != nil {
	// 	return err
	// }

	return nil
}

// // DeleteFile deletes a file from AWS S3 based on the provided key (path).
// func DeleteFile(c *fiber.Ctx, bucketName, key string) error {
// 	sess, err := session.NewSession(&aws.Config{
// 		Region: aws.String("your-aws-region"),
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	svc := s3.New(sess)
// 	_, err = svc.DeleteObject(&s3.DeleteObjectInput{
// 		Bucket: aws.String(bucketName),
// 		Key:    aws.String(key),
// 	})
// 	if err != nil {
// 		if aerr, ok := err.(awserr.Error); ok {
// 			switch aerr.Code() {
// 			case s3.ErrCodeNoSuchKey:
// 				fmt.Println(s3.ErrCodeNoSuchKey, aerr.Error())
// 			default:
// 				fmt.Println(aerr.Error())
// 			}
// 		} else {
// 			fmt.Println(err.Error())
// 		}
// 		return err
// 	}

// 	return nil
// }
