# Simple tcp chat

## Wire protocol

Source at protocol [protocol](protocol) dir. First 2 bytes form a length of wire message, and next N bytes form an
actual message.

## Application level protocol

Can be found in [marshal](marshal) dir. Simple test based protocol with optional tag to the specific user Examples:

`Simple message`

`@username,Private message`

## Integration tests

Intergration tests can be found in [integration](test/integration) dir