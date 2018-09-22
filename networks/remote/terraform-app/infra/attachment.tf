# This is the reason why we can't separate nodes and load balancer creation into different modules.
# https://github.com/hashicorp/terraform/issues/10857
# In short: the list of instances coming from the nodes module is a generated variable
#  and it should be the input for the load-balancer generation. However when attaching the instances
#  to the load-balancer, aws_lb_target_group_attachment.count cannot be a generated value.

#Instance Attachment (autoscaling is the future)
resource "aws_lb_target_group_attachment" "lb_attach" {
  count = "${var.SERVERS*min(length(data.aws_availability_zones.zones.names),var.max_zones)}"
  target_group_arn = "${aws_lb_target_group.lb_target_group.arn}"
  target_id        = "${element(aws_instance.node.*.id,count.index)}"
  port             = 26657
}

resource "aws_lb_target_group_attachment" "lb_attach_lcd" {
  count = "${var.SERVERS*min(length(data.aws_availability_zones.zones.names),var.max_zones)}"
  target_group_arn = "${aws_lb_target_group.lb_target_group_lcd.arn}"
  target_id        = "${element(aws_instance.node.*.id,count.index)}"
  port             = 1317
}

