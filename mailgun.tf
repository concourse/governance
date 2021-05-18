resource "mailgun_route" "routes" {
  for_each = local.team_mail_recipients

  priority = "0"

  # sneaky way of making it easy to import routes back into whichever
  # resource created them
  description = "mailgun_route.routes[\"${each.key}\"]"

  expression = "match_recipient(\"${each.key}@concourse-ci.org\")"

  actions = flatten([
    [for email in each.value : "forward(\"${email}\")"],
    ["stop()"]
  ])
}
