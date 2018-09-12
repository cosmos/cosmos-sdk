resource "aws_vpc" "vpc" {
  cidr_block = "10.0.0.0/16"

  tags {
    Name = "${var.name}"
  }

}

resource "aws_internet_gateway" "internet_gateway" {
  vpc_id = "${aws_vpc.vpc.id}"

  tags {
    Name = "${var.name}"
  }
}

resource "aws_route_table" "route_table" {
  vpc_id = "${aws_vpc.vpc.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.internet_gateway.id}"
  }

  tags {
    Name = "${var.name}"
  }
}

data "aws_availability_zones" "zones" {
  state = "available"
}

resource "aws_subnet" "subnet" {
  count                   = "${min(length(data.aws_availability_zones.zones.names),var.max_zones)}"
  vpc_id                  = "${aws_vpc.vpc.id}"
  availability_zone       = "${element(data.aws_availability_zones.zones.names,count.index)}"
  cidr_block              = "${cidrsubnet(aws_vpc.vpc.cidr_block, 8, count.index)}"
  map_public_ip_on_launch = "true"

  tags {
    Name = "${var.name}-${element(data.aws_availability_zones.zones.names,count.index)}"
  }
}

resource "aws_route_table_association" "route_table_association" {
  count = "${min(length(data.aws_availability_zones.zones.names),var.max_zones)}"
  subnet_id = "${element(aws_subnet.subnet.*.id,count.index)}"
  route_table_id = "${aws_route_table.route_table.id}"
}

resource "aws_security_group" "secgroup" {
  name = "${var.name}"
  vpc_id      = "${aws_vpc.vpc.id}"
  description = "Automated security group for application instances"
  tags {
    Name = "${var.name}"
  }

  ingress {
    from_port = 22
    to_port = 22
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 443
    to_port = 443
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 1317
    to_port = 1317
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 26656
    to_port = 26657
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port = 26660
    to_port = 26660
    protocol = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port = 0
    to_port = 0
    protocol = "-1"
    cidr_blocks = ["0.0.0.0/0"]

  }
}

