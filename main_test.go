package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-eks/sdk/v2/go/eks"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

type mocksShared int

func (mocksShared) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	// fmt.Println(args.TypeToken)
	outputs := args.Inputs.Mappable()
	if args.TypeToken == "eks:index:Cluster" {
		outputs["core"] = map[string]interface{}{
			//"cluster":      map[string]interface{}{"name": "fake"},
			"oidcProvider": map[string]string{"url": "xx"},
		}
	}
	return args.Name + "_id", resource.NewPropertyMapFromMap(outputs), nil
}

func (mocksShared) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	outputs := map[string]interface{}{}
	return resource.NewPropertyMapFromMap(outputs), nil
}

func TestEKS(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		cluster, err := eks.NewCluster(ctx, "example-cluster-go-1", nil)
		if err != nil {
			return err
		}
		assert.NoError(t, err)

		_, err = iam.NewRole(ctx, "test", &iam.RoleArgs{
			AssumeRolePolicy: cluster.Core.OidcProvider().Url().ApplyT(
				func(oidcURL string) (string, error) {
					out, err := json.Marshal(map[string]interface{}{
						"Version": "2012-10-17",
						"Statement": []map[string]interface{}{
							{
								"Action":    "sts:AssumeRoleWithWebIdentity",
								"Effect":    "Allow",
								"Principal": map[string]string{"Federated": fmt.Sprintf("arn:aws:iam::fake:oidc-provider/%s", oidcURL)},
								"Condition": map[string]interface{}{
									"StringEquals": map[string]string{
										fmt.Sprintf("%s:aud", oidcURL): "sts.amazonaws.com",
									},
								},
							},
						},
					})
					return string(out), err
				}).(pulumi.StringOutput),
		})
		assert.NoError(t, err)

		return nil
	}, pulumi.WithMocks("project", "stack", mocksShared(0)))

	assert.NoError(t, err)
}
