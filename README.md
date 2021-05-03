# Simple tcp chat

## Wire protocol

Source at protocol `protocol` dir. First 2 bytes form a length of wire message, and next N bytes form an actual message.

## Integration tests

Intergration tests can be found at `test/intergration/client_server` dir