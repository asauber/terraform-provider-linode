package linode

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/chiefy/linodego"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccLinodeNodeBalancerConfigBasic(t *testing.T) {
	// t.Parallel()

	resName := "linode_nodebalancer_config.foofig"
	nodebalancerName := fmt.Sprintf("tf_test_%s", acctest.RandString(10))
	config := testAccCheckLinodeNodeBalancerConfigBasic(nodebalancerName)
	resource.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		CheckDestroy:              testAccCheckLinodeNodeBalancerConfigDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				//ImportState:       true,
				//ImportStateVerify: true,
				Config:       config,
				ResourceName: resName,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckLinodeNodeBalancerConfigExists,
					resource.TestCheckResourceAttr(resName, "port", "8080"),
					resource.TestCheckResourceAttr(resName, "protocol", string(linodego.ProtocolHTTP)),
					resource.TestCheckResourceAttr(resName, "check", string(linodego.CheckHTTP)),
					resource.TestCheckResourceAttr(resName, "check_path", "/"),

					resource.TestCheckResourceAttrSet(resName, "stickiness"),
					resource.TestCheckResourceAttrSet(resName, "check_attempts"),
					resource.TestCheckResourceAttrSet(resName, "check_timeout"),
				),
			},
		},
	})
}

func TestAccLinodeNodeBalancerConfigUpdate(t *testing.T) {
	t.Parallel()

	resName := "linode_nodebalancer_config.foofig"
	nodebalancerName := fmt.Sprintf("tf_test_%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLinodeNodeBalancerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckLinodeNodeBalancerConfigBasic(nodebalancerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeNodeBalancerConfigExists,
					resource.TestCheckResourceAttr(resName, "port", "8080"),
					resource.TestCheckResourceAttr(resName, "protocol", string(linodego.ProtocolHTTP)),
					resource.TestCheckResourceAttr(resName, "check", string(linodego.CheckHTTP)),
					resource.TestCheckResourceAttr(resName, "check_path", "/"),

					resource.TestCheckResourceAttrSet(resName, "stickiness"),
					resource.TestCheckResourceAttrSet(resName, "check_attempts"),
					resource.TestCheckResourceAttrSet(resName, "check_timeout"),
				),
			},
			{
				Config: testAccCheckLinodeNodeBalancerConfigUpdates(nodebalancerName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLinodeNodeBalancerConfigExists,
					resource.TestCheckResourceAttr(resName, "port", "8088"),
					resource.TestCheckResourceAttr(resName, "protocol", string(linodego.ProtocolHTTP)),
					resource.TestCheckResourceAttr(resName, "check", string(linodego.CheckHTTP)),
					resource.TestCheckResourceAttr(resName, "check_path", "/foo"),
					resource.TestCheckResourceAttr(resName, "check_attempts", "3"),
					resource.TestCheckResourceAttr(resName, "check_timeout", "30"),

					resource.TestCheckResourceAttr(resName, "stickiness", string(linodego.StickinessHTTPCookie)),
				),
			},
		},
	})
}

func testAccCheckLinodeNodeBalancerConfigExists(s *terraform.State) error {
	client := testAccProvider.Meta().(linodego.Client)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_nodebalancer_config" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		nodebalancerID, err := strconv.Atoi(rs.Primary.Attributes["nodebalancer_id"])

		_, err = client.GetNodeBalancerConfig(context.Background(), nodebalancerID, id)
		if err != nil {
			return fmt.Errorf("Error retrieving state of NodeBalancer Config %s: %s", rs.Primary.Attributes["label"], err)
		}
	}

	return nil
}

func testAccCheckLinodeNodeBalancerConfigDestroy(s *terraform.State) error {
	client, ok := testAccProvider.Meta().(linodego.Client)
	if !ok {
		return fmt.Errorf("Failed to get Linode client")
	}
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "linode_nodebalancer_config" {
			continue
		}

		id, err := strconv.Atoi(rs.Primary.ID)
		nodebalancerID, err := strconv.Atoi(rs.Primary.Attributes["nodebalancer_id"])

		if err != nil {
			return fmt.Errorf("Failed parsing %v to int", rs.Primary.ID)
		}
		if id == 0 {
			return fmt.Errorf("Would have considered %v as %d", rs.Primary.ID, id)

		}

		_, err = client.GetNodeBalancerConfig(context.Background(), nodebalancerID, id)

		if err == nil {
			return fmt.Errorf("NodeBalancer Config with id %d still exists", id)
		}

		if apiErr, ok := err.(linodego.Error); ok && apiErr.Code != 404 {
			return fmt.Errorf("Failed to request NodeBalancer Config with id %d", id)
		}
	}

	return nil
}
func testAccCheckLinodeNodeBalancerConfigBasic(nodebalancer string) string {
	return testAccCheckLinodeNodeBalancerBasic(nodebalancer) + `
resource "linode_nodebalancer_config" "foofig" {
	nodebalancer_id = "${linode_nodebalancer.foobar.id}"
	port = 8080
	protocol = "http"
	check = "http"
	check_path = "/"
}
`
}

func testAccCheckLinodeNodeBalancerConfigUpdates(nodebalancer string) string {
	return testAccCheckLinodeNodeBalancerBasic(nodebalancer) + `
resource "linode_nodebalancer_config" "foofig" {
	nodebalancer_id = "${linode_nodebalancer.foobar.id}"
	port = 8088
	protocol = "http"
	check = "http"
	check_path = "/foo"
	check_attempts = 3
	check_timeout = 30
	stickiness = "http_cookie"
	algorithm = "source"
}

`
}
