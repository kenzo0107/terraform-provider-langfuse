package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestProjectResource_CreateAndRead(t *testing.T) {
	mock := newMockLangfuseServer(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(mock.URL()) + `
resource "langfuse_project" "test" {
  name = "my-project"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("langfuse_project.test", "name", "my-project"),
					resource.TestCheckResourceAttrSet("langfuse_project.test", "id"),
				),
			},
		},
	})
}

func TestProjectResource_Update(t *testing.T) {
	mock := newMockLangfuseServer(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(mock.URL()) + `
resource "langfuse_project" "test" {
  name = "original-name"
}
`,
				Check: resource.TestCheckResourceAttr("langfuse_project.test", "name", "original-name"),
			},
			{
				Config: providerConfig(mock.URL()) + `
resource "langfuse_project" "test" {
  name = "updated-name"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("langfuse_project.test", "name", "updated-name"),
					resource.TestCheckResourceAttrSet("langfuse_project.test", "id"),
				),
			},
		},
	})
}

func TestProjectResource_IDIsStableOnUpdate(t *testing.T) {
	mock := newMockLangfuseServer(t)

	var firstID string

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(mock.URL()) + `
resource "langfuse_project" "test" {
  name = "first-name"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["langfuse_project.test"]
						if rs == nil {
							return fmt.Errorf("resource not found")
						}
						firstID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: providerConfig(mock.URL()) + `
resource "langfuse_project" "test" {
  name = "renamed"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["langfuse_project.test"]
						if rs == nil {
							return fmt.Errorf("resource not found")
						}
						if rs.Primary.ID != firstID {
							return fmt.Errorf("ID changed after update: was %s, now %s", firstID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func TestProjectResource_ImportState(t *testing.T) {
	mock := newMockLangfuseServer(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(mock.URL()) + `
resource "langfuse_project" "test" {
  name = "importable-project"
}
`,
			},
			{
				ResourceName:      "langfuse_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestProjectResource_DeletedExternally(t *testing.T) {
	mock := newMockLangfuseServer(t)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: providerConfig(mock.URL()) + `
resource "langfuse_project" "test" {
  name = "ephemeral-project"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("langfuse_project.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["langfuse_project.test"]
						if rs == nil {
							return fmt.Errorf("resource not found in state")
						}
						mock.mu.Lock()
						defer mock.mu.Unlock()
						delete(mock.projects, rs.Primary.ID)
						return nil
					},
				),
				// After deleting externally, the next plan should recreate it.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestProjectResource_InvalidProvider(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "langfuse" {
  public_key = "pub"
  secret_key = "sec"
  host       = "http://127.0.0.1:1"
}
resource "langfuse_project" "test" {
  name = "test"
}
`,
				ExpectError: regexp.MustCompile(`Creating project`),
			},
		},
	})
}
