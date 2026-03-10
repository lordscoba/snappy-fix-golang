package config

type AmazonS3Bucket struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Bucket          string
	PublicBase      string
}
