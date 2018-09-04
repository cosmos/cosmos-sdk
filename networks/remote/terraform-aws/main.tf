#Terraform Configuration

#See https://docs.aws.amazon.com/general/latest/gr/rande.html#ec2_region
#eu-west-3 does not contain CentOS images
#us-east-1 usually contains other infrastructure and creating keys and security groups might conflict with that
variable "REGIONS" {
  description = "AWS Regions"
  type = "list"
  default = ["us-east-2", "us-west-1", "us-west-2", "ap-south-1", "ap-northeast-2", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ca-central-1", "eu-central-1", "eu-west-1", "eu-west-2", "sa-east-1"]
}

variable "TESTNET_NAME" {
  description = "Name of the testnet"
  default = "remotenet"
}

variable "REGION_LIMIT" {
  description = "Number of regions to populate"
  default = "1"
}

variable "SERVERS" {
  description = "Number of servers in an availability zone"
  default = "1"
}

variable "SSH_PRIVATE_FILE" {
  description = "SSH private key file to be used to connect to the nodes"
  type = "string"
}

variable "SSH_PUBLIC_FILE" {
  description = "SSH public key file to be used on the nodes"
  type = "string"
}


# ap-southeast-1 and ap-southeast-2 does not contain the newer CentOS 1704 image
variable "image" {
  description = "AWS image name"
  default = "CentOS Linux 7 x86_64 HVM EBS 1703_01"
}

variable "instance_type" {
  description = "AWS instance type"
  default = "t2.large"
}

module "nodes-0" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,0)}"
  multiplier       = "0"
  execute          = "${var.REGION_LIMIT > 0}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-1" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,1)}"
  multiplier       = "1"
  execute          = "${var.REGION_LIMIT > 1}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-2" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,2)}"
  multiplier       = "2"
  execute          = "${var.REGION_LIMIT > 2}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-3" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,3)}"
  multiplier       = "3"
  execute          = "${var.REGION_LIMIT > 3}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-4" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,4)}"
  multiplier       = "4"
  execute          = "${var.REGION_LIMIT > 4}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-5" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,5)}"
  multiplier       = "5"
  execute          = "${var.REGION_LIMIT > 5}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-6" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,6)}"
  multiplier       = "6"
  execute          = "${var.REGION_LIMIT > 6}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-7" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,7)}"
  multiplier       = "7"
  execute          = "${var.REGION_LIMIT > 7}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-8" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,8)}"
  multiplier       = "8"
  execute          = "${var.REGION_LIMIT > 8}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-9" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,9)}"
  multiplier       = "9"
  execute          = "${var.REGION_LIMIT > 9}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-10" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,10)}"
  multiplier       = "10"
  execute          = "${var.REGION_LIMIT > 10}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-11" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,11)}"
  multiplier       = "11"
  execute          = "${var.REGION_LIMIT > 11}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-12" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,12)}"
  multiplier       = "12"
  execute          = "${var.REGION_LIMIT > 12}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

module "nodes-13" {
  source           = "nodes"
  name             = "${var.TESTNET_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  region           = "${element(var.REGIONS,13)}"
  multiplier       = "13"
  execute          = "${var.REGION_LIMIT > 13}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  SERVERS          = "${var.SERVERS}"
}

output "public_ips" {
  value = "${concat(
		module.nodes-0.public_ips,
		module.nodes-1.public_ips,
		module.nodes-2.public_ips,
		module.nodes-3.public_ips,
		module.nodes-4.public_ips,
		module.nodes-5.public_ips,
		module.nodes-6.public_ips,
		module.nodes-7.public_ips,
		module.nodes-8.public_ips,
		module.nodes-9.public_ips,
		module.nodes-10.public_ips,
		module.nodes-11.public_ips,
		module.nodes-12.public_ips,
		module.nodes-13.public_ips
		)}",
}

