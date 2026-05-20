package config

import "fmt"

func GetTenantBucketPolicy(bucketName string) string {
	return fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Deny",
			"Principal": {"AWS": ["*"]},
			"Action": ["s3:GetObject"],
			"Resource": ["arn:aws:s3:::%s/*"]
		}]
	}`, bucketName)
}
