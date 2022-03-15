package main

// for test only

import (
	"errors"
	"fmt"
	"github.com/IBM/ibm-cos-sdk-go/aws"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials"
	"github.com/IBM/ibm-cos-sdk-go/aws/credentials/ibmiam"
	"github.com/IBM/ibm-cos-sdk-go/aws/session"
	"github.com/IBM/ibm-cos-sdk-go/service/s3"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type IBMCos struct {
	ApiKey            string `yaml:"ibm-api-key"`
	ServiceInstanceID string `yaml:"ibm-service-instance-id"`
	AuthEndpoint      string `yaml:"ibm-auth-endpoint"`
	ServiceEndpoint   string `yaml:"ibm-service-endpoint"`

	ImageBucketName   string `yaml:"ibm-image-bucket-name"`
	ImageBucketRegion string `yaml:"ibm-image-bucket-region"`
	ImageObjectName   string `yaml:"ibm-image-object-name"`

	AccessKeyId     string `yaml:"ibm-access-key-id"`
	SecretAccessKey string `yaml:"ibm-secret-access-key"`
}

type ImageServer struct {
	Config IBMCos

	KeyCount int64
	Keys     []string
}

func NewImageServer() *ImageServer {
	c := &ImageServer{}

	return c
}

func (imageServer *ImageServer) Configure(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, &imageServer.Config)

	if err != nil {
		data, _ = yaml.Marshal(&imageServer.Config)
		_ = ioutil.WriteFile(filename, data, 0644)
		return err
	}

	return nil
}

func (imageServer *ImageServer) FetchFileKeys() error {
	// Create config
	ibmcos := &imageServer.Config
	cred := ibmiam.NewStaticCredentials(aws.NewConfig(), ibmcos.AuthEndpoint, ibmcos.ApiKey, ibmcos.ServiceInstanceID)
	conf := aws.NewConfig().
		WithRegion(ibmcos.ImageBucketRegion).
		WithEndpoint(ibmcos.ServiceEndpoint).
		WithCredentials(cred).
		WithS3ForcePathStyle(true)

	// Create client
	sess := session.Must(session.NewSession())
	client := s3.New(sess, conf)

	// List content of bucket
	l, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(ibmcos.ImageBucketName),
	})

	if err != nil {
		return err
	}

	imageServer.KeyCount = *l.KeyCount
	imageServer.Keys = make([]string, *l.KeyCount)
	for i := int64(0); i < *l.KeyCount; i++ {
		imageServer.Keys[i] = *l.Contents[i].Key
	}

	return nil
}

func (imageServer *ImageServer) GetRandomUrl(expire time.Duration) (string, http.Header, error) {
	ibmcos := &imageServer.Config
	if imageServer.Keys == nil {
		return "", nil, errors.New("key list is empty")
	}

	n := rand.Int63n(imageServer.KeyCount)

	value := credentials.Value{
		AccessKeyID:     ibmcos.AccessKeyId,
		SecretAccessKey: ibmcos.SecretAccessKey,
	}
	fmt.Println(value.HasKeys())
	conf := aws.NewConfig().
		WithRegion(ibmcos.ImageBucketRegion).
		WithEndpoint(ibmcos.ServiceEndpoint).
		WithCredentials(credentials.NewStaticCredentialsFromCreds(value)).
		WithS3ForcePathStyle(true)
	sess := session.Must(session.NewSession())
	client := s3.New(sess, conf)

	req, _ := client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(ibmcos.ImageBucketName),
		Key:    aws.String(imageServer.Keys[n]),
	})

	return req.PresignRequest(expire)
}
