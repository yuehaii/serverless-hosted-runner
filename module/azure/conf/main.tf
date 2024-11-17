terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.0"
      # version = "4.3.0"
      # version = "3.77.0"
    }
  }
}

# terraform {
#   required_version = ">=1.0"
#   required_providers {
#     azurerm = {
#       source  = "hashicorp/azurerm"
#       version = "~>3.0"
#     }
#     random = {
#       source  = "hashicorp/random"
#       version = "~>3.0"
#     }
#   }
#   backend "azurerm" {}
# }
