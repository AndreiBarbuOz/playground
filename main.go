package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	log "github.com/sirupsen/logrus"
)

func main() {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})

	// Create S3 service client
	svc := s3.New(sess)
	result, err := svc.ListBuckets(nil)
	if err != nil {
		log.Infof("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")
	var wg sync.WaitGroup

	for _, b := range result.Buckets {

		wg.Add(1)

		bucketName := aws.StringValue(b.Name)
		bucketCreation := b.CreationDate

		go func() {
			defer wg.Done()

			if bucketCreation.Before(time.Now().AddDate(0, 0, 4)) {
				fmt.Printf("* %s created on %s\n", bucketName, bucketCreation)

				location, err := svc.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: aws.String(bucketName)})
				if err != nil {
					log.Printf("Couldn't get bucket location %v. Here's why: %v\n", bucketName, err)
				}
				region := aws.String("us-east-1")
				if location.LocationConstraint != nil {
					region = location.LocationConstraint
				}

				log.Printf("Bucket %v is in region %v\n", bucketName, region)
				bucketSession, err := session.NewSession(&aws.Config{
					Region: region,
				})
				if err != nil {
					log.Printf("Couldn't create bucket session %v. Here's why: %v\n", bucketName, err)
				}

				bucketSvc := s3.New(bucketSession)

				iter := s3manager.NewDeleteListIterator(bucketSvc, &s3.ListObjectsInput{
					Bucket: aws.String(bucketName),
					Prefix: aws.String(""),
				})

				// use the iterator to delete the files.
				if err := s3manager.NewBatchDeleteWithClient(bucketSvc).Delete(context.Background(), iter); err != nil {
					log.Infof("failed to delete files under given directory: %v", err)
				}

				_, err = bucketSvc.DeleteBucket(&s3.DeleteBucketInput{Bucket: aws.String(bucketName)})
				if err != nil {
					log.Printf("Couldn't delete bucket %v. Here's why: %v\n", bucketName, err)
				}
			}
		}()
	}

	wg.Wait()
}
