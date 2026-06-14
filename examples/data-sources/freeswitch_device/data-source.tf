data "freeswitch_device" "reception" {
  mac = "80:5e:c0:11:22:33"
}

output "reception_number" {
  value = data.freeswitch_device.reception.number
}
