/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package blob_storage

import (
	"github.com/devtron-labs/common-lib/utils"
	"testing"
)

func TestPutWithCommand_MINIO_Success(t *testing.T) {
	logger, _ := utils.NewSugardLogger()
	service := NewBlobStorageServiceImpl(logger)
	request := &BlobStorageRequest{
		SourceKey:      "/Users/nishant/Desktop/scope-var.yaml",
		DestinationKey: "scope-var.yaml",
		StorageType:    BLOB_STORAGE_S3,
		AwsS3BaseConfig: &AwsS3BaseConfig{
			Region:      "ap-south-1",
			AccessKey:   "",
			Passkey:     "",
			EndpointUrl: "",
			BucketName:  "nishant-test",
		},
	}

	err := service.PutWithCommand(request)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestGet_MINIO_Success(t *testing.T) {
	logger, _ := utils.NewSugardLogger()
	service := NewBlobStorageServiceImpl(logger)
	request := &BlobStorageRequest{
		StorageType: BLOB_STORAGE_S3,
		AwsS3BaseConfig: &AwsS3BaseConfig{
			Region:      "ap-south-1",
			AccessKey:   "",
			Passkey:     "",
			EndpointUrl: "",
			BucketName:  "nishant-test",
		},
		SourceKey:      "scope-var.yaml",
		DestinationKey: "/tmp/scope-var.yaml",
	}

	success, numBytes, err := service.Get(request)
	if !success || err != nil {
		t.Errorf("expected success, got %v, %v", success, err)
	}
	if numBytes == 0 {
		t.Errorf("expected non-zero bytes, got %d", numBytes)
	}
}

func TestPutWithCommand_S3_Success(t *testing.T) {
	logger, _ := utils.NewSugardLogger()
	service := NewBlobStorageServiceImpl(logger)
	request := &BlobStorageRequest{
		SourceKey:      "/Users/nishant/Desktop/scope-var.yaml",
		DestinationKey: "scope-var.yaml",
		StorageType:    BLOB_STORAGE_S3,
		AwsS3BaseConfig: &AwsS3BaseConfig{
			Region:     "ap-south-1",
			AccessKey:  "",
			Passkey:    "",
			BucketName: "",
		},
	}

	err := service.PutWithCommand(request)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestGet_S3_Success(t *testing.T) {
	logger, _ := utils.NewSugardLogger()
	service := NewBlobStorageServiceImpl(logger)
	request := &BlobStorageRequest{
		StorageType: BLOB_STORAGE_S3,
		AwsS3BaseConfig: &AwsS3BaseConfig{
			Region:     "ap-south-1",
			AccessKey:  "",
			Passkey:    "",
			BucketName: "",
		},
		SourceKey:      "scope-var.yaml",
		DestinationKey: "/tmp/scope-var-s3.yaml",
	}

	success, numBytes, err := service.Get(request)
	if !success || err != nil {
		t.Errorf("expected success, got %v, %v", success, err)
	}
	if numBytes == 0 {
		t.Errorf("expected non-zero bytes, got %d", numBytes)
	}
}
