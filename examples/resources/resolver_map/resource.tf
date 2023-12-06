resource "resolver_map" "example" {
  keys        = ["a", "b", "c"]
  result_keys = ["a", "c"]
  values      = ["1", "2", "3"]
}
