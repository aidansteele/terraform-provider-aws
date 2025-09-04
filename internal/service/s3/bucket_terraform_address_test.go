// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccS3Bucket_terraformAddressTag(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.test"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	// NOTE: Currently testing resource type name ("aws_s3_bucket") due to plugin framework limitations.
	// Future enhancement should test full resource address like "module.storage.aws_s3_bucket.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_terraformAddressTag(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.Project", "terraform-address-tag"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "3"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Project", "terraform-address-tag"),
					// Currently verifies resource type name; future: should verify full address
					resource.TestCheckResourceAttr(resourceName, "tags_all.terraform:address", "aws_s3_bucket"),
					testAccCheckBucketHasTerraformAddressTag(ctx, resourceName, "aws_s3_bucket"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrForceDestroy},
			},
			{
				Config: testAccBucketConfig_terraformAddressTagUpdated(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "production"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Environment", "production"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.terraform:address", "aws_s3_bucket"),
					testAccCheckBucketHasTerraformAddressTag(ctx, resourceName, "aws_s3_bucket"),
				),
			},
		},
	})
}

func TestAccS3Bucket_terraformAddressTagOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_s3_bucket.test"
	bucketName := sdkacctest.RandomWithPrefix("tf-test-bucket")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.S3ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBucketDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBucketConfig_terraformAddressTagAttemptOverride(bucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBucketExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.terraform:address", "user-provided-value"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.Environment", "test"),
					// The provider should override the user-provided value
					resource.TestCheckResourceAttr(resourceName, "tags_all.terraform:address", "aws_s3_bucket"),
					testAccCheckBucketHasTerraformAddressTag(ctx, resourceName, "aws_s3_bucket"),
				),
			},
		},
	})
}

func testAccCheckBucketHasTerraformAddressTag(ctx context.Context, n, expectedValue string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).S3Client(ctx)

		tags, err := tfs3.BucketListTags(ctx, conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error listing S3 bucket tags: %w", err)
		}

		actualValue, exists := tags[tftags.TerraformAddressTagKey]
		if !exists {
			return fmt.Errorf("terraform:address tag not found on S3 bucket")
		}

		if actualValue != expectedValue {
			return fmt.Errorf("terraform:address tag has value %q, expected %q", actualValue, expectedValue)
		}

		return nil
	}
}

func testAccBucketConfig_terraformAddressTag(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  tags = {
    Environment = "test"
    Project     = "terraform-address-tag"
  }
}
`, bucketName)
}

func testAccBucketConfig_terraformAddressTagUpdated(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  tags = {
    Environment = "production"
  }
}
`, bucketName)
}

func testAccBucketConfig_terraformAddressTagAttemptOverride(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true

  tags = {
    Environment       = "test"
    "terraform:address" = "user-provided-value"
  }
}
`, bucketName)
}