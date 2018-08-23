#Terraform Configuration

variable "APP_NAME" {
  description = "Name of the application"
}

variable "SERVERS" {
  description = "Number of servers in an availability zone"
  default = "1"
}

variable "MAX_ZONES" {
  description = "Maximum number of availability zones to use"
  default = "4"
}

#See https://docs.aws.amazon.com/general/latest/gr/rande.html#ec2_region
#eu-west-3 does not contain CentOS images
variable "REGION" {
  description = "AWS Regions"
  default = "us-east-1"
}

variable "SSH_PRIVATE_FILE" {
  description = "SSH private key file to be used to connect to the nodes"
  type = "string"
}

variable "SSH_PUBLIC_FILE" {
  description = "SSH public key file to be used on the nodes"
  type = "string"
}

variable "CERTIFICATE_ARN" {
  description = "Load-balancer certificate AWS ARN"
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

provider "aws" {
  region = "${var.REGION}"
}

module "nodes" {
  source           = "infra"
  name             = "${var.APP_NAME}"
  image_name       = "${var.image}"
  instance_type    = "${var.instance_type}"
  ssh_public_file  = "${var.SSH_PUBLIC_FILE}"
  ssh_private_file = "${var.SSH_PRIVATE_FILE}"
  certificate_arn  = "${var.CERTIFICATE_ARN}"
  SERVERS          = "${var.SERVERS}"
  max_zones        = "${var.MAX_ZONES}"
}

output "public_ips" {
  value = "${module.nodes.public_ips}",
}

output "lb_name" {
  value = "${module.nodes.lb_name}"
}

