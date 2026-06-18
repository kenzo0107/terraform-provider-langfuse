package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestProjectDataSource_Read(t *testing.T) {
	mock := newMockLangfuseServer(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(mock.URL()) + `
resource "langfuse_project" "test" {
  name = "datasource-test-project"
}

data "langfuse_project" "test" {
  id = langfuse_project.test.id
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.langfuse_project.test", "name", "datasource-test-project"),
					resource.TestCheckResourceAttrSet("data.langfuse_project.test", "id"),
					func(s *terraform.State) error {
						resID := s.RootModule().Resources["langfuse_project.test"].Primary.ID
						dsID := s.RootModule().Resources["data.langfuse_project.test"].Primary.ID
						if resID != dsID {
							return fmt.Errorf("resource id %q != data source id %q", resID, dsID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestProjectDataSource_NotFound(t *testing.T) {
	mock := newMockLangfuseServer(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(mock.URL()) + `
data "langfuse_project" "missing" {
  id = "does-not-exist"
}
`,
				ExpectError: regexp.MustCompile(`Project Not Found`),
			},
		},
	})
}
