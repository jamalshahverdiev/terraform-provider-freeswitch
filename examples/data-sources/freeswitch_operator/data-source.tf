data "freeswitch_operator" "reception" {
  subject = "a1b2c3d4-0000-1111-2222-333344445555"
}

output "reception_extension" {
  value = data.freeswitch_operator.reception.number
}
