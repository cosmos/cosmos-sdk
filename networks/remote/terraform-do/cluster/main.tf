resource "digitalocean_tag" "cluster" {
  name = "${var.name}"
}

resource "digitalocean_ssh_key" "cluster" {
  name       = "${var.name}"
  public_key = "${file(var.ssh_public_file)}"
}

resource "digitalocean_droplet" "cluster" {
  name = "${var.name}-node${count.index}"
  image = "centos-7-x64"
  size = "${var.instance_size}"
  region = "${element(var.regions, count.index)}"
  ssh_keys = ["${digitalocean_ssh_key.cluster.id}"]
  count = "${var.servers}"
  tags = ["${digitalocean_tag.cluster.id}"]

  lifecycle = {
	prevent_destroy = false
  }

  connection {
    private_key = "${file(var.ssh_private_file)}"
  }

  provisioner "file" {
    source = "files/terraform.sh"
    destination = "/tmp/terraform.sh"
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /tmp/terraform.sh",
      "/tmp/terraform.sh ${var.name} ${count.index}",
    ]
  }

}

