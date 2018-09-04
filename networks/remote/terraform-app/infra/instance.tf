resource "aws_key_pair" "key" {
  key_name   = "${var.name}"
  public_key = "${file(var.ssh_public_file)}"
}

data "aws_ami" "linux" {
  most_recent = true
  filter {
    name   = "name"
    values = ["${var.image_name}"]
  }
}

resource "aws_instance" "node" {
#  depends_on = ["${element(aws_route_table_association.route_table_association.*,count.index)}"]
  count = "${var.SERVERS*min(length(data.aws_availability_zones.zones.names),var.max_zones)}"
  ami = "${data.aws_ami.linux.image_id}"
  instance_type = "${var.instance_type}"
  key_name = "${aws_key_pair.key.key_name}"
  associate_public_ip_address = true
  vpc_security_group_ids = [ "${aws_security_group.secgroup.id}" ]
  subnet_id = "${element(aws_subnet.subnet.*.id,count.index)}"
  availability_zone = "${element(data.aws_availability_zones.zones.names,count.index)}"

  tags {
    Environment = "${var.name}"
    Name = "${var.name}-${element(data.aws_availability_zones.zones.names,count.index)}"
  }

  volume_tags {
    Environment = "${var.name}"
    Name = "${var.name}-${element(data.aws_availability_zones.zones.names,count.index)}-VOLUME"
  }

  root_block_device {
    volume_size = 40
  }

  connection {
    user = "centos"
    private_key = "${file(var.ssh_private_file)}"
    timeout = "600s"
  }

  provisioner "file" {
    source = "files/terraform.sh"
    destination = "/tmp/terraform.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /tmp/terraform.sh",
      "sudo /tmp/terraform.sh ${var.name} ${count.index}",
    ]
  }

}

