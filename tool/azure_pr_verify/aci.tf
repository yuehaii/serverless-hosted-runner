terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.11.0"
    }
  }
}

provider "azurerm" {
  features {}
}

# resource "azurerm_resource_group" "example" {
#   name     = "example-resources"
#   location = "chinanorth3"
# }

resource "azurerm_container_group" "example" {
  name                = "example-continst"
  location            = "chinanorth3"
  resource_group_name = "sls-runner"
  ip_address_type     = "Private"
  # dns_name_label      = "aci-label"
  os_type             = "Linux"
  subnet_ids          = ["/subscriptions/b5937c02-df3d-4846-849b-6fe858a84d0e/resourceGroups/rg-auto-cn-north3-test/providers/Microsoft.Network/virtualNetworks/vnet-auto-cn-north3-test/subnets/sls-runner-acl-subnet"]
  container {
    name   = "hello-world"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    cpu    = "1"
    memory = "1"

    ports {
      port     = 443
      protocol = "TCP"
    }
  }

  tags = {
   environment = "testing"
  }
}