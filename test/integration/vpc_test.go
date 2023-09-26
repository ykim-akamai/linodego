package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/linode/linodego"
)

func formatVPCError(err error, action string, vpcID *int) error {
	if err == nil {
		return nil
	}
	if vpcID == nil {
		return fmt.Errorf(
			"an error occurs when %v the VPC(s): %v",
			action,
			err,
		)
	}
	return fmt.Errorf(
		"an error occurs when %v the VPC %v: %v",
		action,
		*vpcID,
		err,
	)
}

func createVPC(t *testing.T, client *linodego.Client) (*linodego.VPC, func(), error) {
	t.Helper()
	createOpts := linodego.VPCCreateOptions{
		Label:  "go-test-vpc-" + getUniqueText(),
		Region: getRegionsWithCaps(t, client, []string{"VPCs"})[0],
	}
	vpc, err := client.CreateVPC(context.Background(), createOpts)
	if err != nil {
		t.Fatal(formatVPCError(err, "creating", nil))
	}

	teardown := func() {
		if err := client.DeleteVPC(context.Background(), vpc.ID); err != nil {
			t.Error(formatVPCError(err, "deleting", &vpc.ID))
		}
	}
	return vpc, teardown, err
}

func setupVPC(t *testing.T, fixturesYaml string) (
	*linodego.Client,
	*linodego.VPC,
	func(),
	error,
) {
	t.Helper()
	client, fixtureTeardown := createTestClient(t, fixturesYaml)

	vpc, vpcTeardown, err := createVPC(t, client)

	teardown := func() {
		vpcTeardown()
		fixtureTeardown()
	}
	return client, vpc, teardown, err
}

func vpcCheck(vpc *linodego.VPC, t *testing.T) {
	if vpc.ID == 0 {
		t.Errorf("expected a VPC ID, but got 0")
	}
	assertDateSet(t, vpc.Created)
	assertDateSet(t, vpc.Updated)
}

func vpcCreateOptionsCheck(
	opts *linodego.VPCCreateOptions,
	vpc *linodego.VPC,
	t *testing.T,
) {
	good := (opts.Description == vpc.Description &&
		opts.Label == vpc.Label &&
		opts.Region == vpc.Region &&
		len(opts.Subnets) == len(vpc.Subnets))

	for i := 0; i < minInt(len(opts.Subnets), len(vpc.Subnets)); i++ {
		good = good && (opts.Subnets[i].IPv4 == vpc.Subnets[i].IPv4 &&
			opts.Subnets[i].Label == vpc.Subnets[i].Label)
	}

	if !good {
		t.Error(
			"the VPC instance and the VPC creation options instance are mismatched",
		)
	}
}

func vpcUpdateOptionsCheck(
	opts *linodego.VPCUpdateOptions,
	vpc *linodego.VPC,
	t *testing.T,
) {
	if !(opts.Description == vpc.Description && opts.Label == vpc.Label) {
		t.Error("the VPC instance and VPC Update Options instance are mismatched")
	}
}

func TestVPC_Create(t *testing.T) {
	_, vpc, teardown, err := setupVPC(t, "fixtures/TestVPC_Create")
	defer teardown()
	if err != nil {
		t.Error(formatVPCError(err, "setting up", nil))
	}
	vpcCheck(vpc, t)
	opts := vpc.GetCreateOptions()
	vpcCreateOptionsCheck(&opts, vpc, t)
}

func TestVPC_Update(t *testing.T) {
	client, vpc, teardown, err := setupVPC(t, "fixtures/TestVPC_Update")
	defer teardown()
	if err != nil {
		t.Error(formatVPCError(err, "setting up", nil))
	}
	vpcCheck(vpc, t)

	opts := vpc.GetUpdateOptions()
	vpcUpdateOptionsCheck(&opts, vpc, t)

	updatedDescription := "updated description"
	updatedLabel := "updated-label"

	opts.Description = updatedDescription
	opts.Label = updatedLabel
	updatedVPC, err := client.UpdateVPC(context.Background(), vpc.ID, opts)
	if err != nil {
		t.Error(formatVPCError(err, "updating", &vpc.ID))
	}
	vpcUpdateOptionsCheck(&opts, updatedVPC, t)
}

func TestVPC_List(t *testing.T) {
	client, vpc, teardown, err := setupVPC(t, "fixtures/TestVPC_List")
	defer teardown()
	if err != nil {
		t.Error(formatVPCError(err, "setting up", nil))
	}
	vpcCheck(vpc, t)

	vpcs, err := client.ListVPC(context.Background(), nil)
	if err != nil {
		t.Error(formatVPCError(err, "listing", nil))
	}

	found := false
	for _, v := range vpcs {
		if v.ID == vpc.ID {
			found = true
		}
	}

	if !found {
		t.Errorf("vpc %v not found in list", vpc.ID)
	}
}