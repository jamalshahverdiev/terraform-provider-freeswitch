data "freeswitch_voicemail" "reception" {
  domain = "example.com"
  number = "1001"
}

output "reception_unread" {
  value = data.freeswitch_voicemail.reception.unread
}
