#Terraform Configuration

variable "DO_API_TOKEN" {
  description = "DigitalOcean Access Token"
}

variable "TESTNET_NAME" {
  description = "Name of the testnet"
  default = "sentrynet"
}

variable "SSH_KEY_FILE" {
  description = "SSH public key file to be used on the nodes"
  type = "string"
}

variable "SERVERS" {
  description = "Number of nodes in testnet"
  default = "4"
}

provider "digitalocean" {
  token = "${var.DO_API_TOKEN}"
}

module "cluster" {
  source           = "./cluster"
  name             = "${var.TESTNET_NAME}"
  ssh_key          = "${var.SSH_KEY_FILE}"
  servers          = "${var.SERVERS}"
}


output "public_ips" {
  value = "${module.cluster.public_ips}"
}

