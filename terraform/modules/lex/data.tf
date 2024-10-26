locals {
  intent_files = fileset("${path.module}/../../../intents", "*.json")
  intents = { for file in local.intent_files : replace(basename(file), ".json", "") => jsondecode(file("${path.module}/../../../intents/${file}")) }
  slots_map = flatten([for intent_name, intent in local.intents : [for slot in lookup(intent, "slots", []) : merge(slot, { intent_name = intent_name })]])
}