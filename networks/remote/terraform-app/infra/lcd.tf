resource "aws_lb_listener" "lb_listener_lcd" {
  load_balancer_arn = "${aws_lb.lb.arn}"
  port              = "1317"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-TLS-1-2-Ext-2018-06"
  certificate_arn   = "${var.certificate_arn}"

  default_action {
    target_group_arn = "${aws_lb_target_group.lb_target_group_lcd.arn}"
    type             = "forward"
  }
}

resource "aws_lb_listener_rule" "listener_rule_lcd" {
  listener_arn = "${aws_lb_listener.lb_listener_lcd.arn}"
  priority     = "100"
  action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.lb_target_group_lcd.id}"
  }
  condition {
    field  = "path-pattern"
    values = ["/"]
  }
}

resource "aws_lb_target_group" "lb_target_group_lcd" {
  name     = "${var.name}lcd"
  port     = "1317"
  protocol = "HTTP"
  vpc_id   = "${aws_vpc.vpc.id}"
  tags {
    name = "${var.name}"
  }
  health_check {
    path                = "/node_version"
  }
}

