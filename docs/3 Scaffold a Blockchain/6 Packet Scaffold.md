# Packet Scaffold

An IBC packet is a data structure with sequence-related metadata and an opaque value field referred to as the packet data. The packet semantics are defined by the application layer, for example, token amount and denomination. Packets are sent through [IBC channels](https://docs.cosmos.network/master/ibc/overview.html) and can only be scaffolded in IBC modules.

`starport packet` command scaffolds IBC packets.

```
starport packet [packetName] [field1] [field2]
```

`--ack`

  Comma-separated list (no spaces) of fields that describe the acknowledgement fields.

When you scaffold a packet, the following files and directories are created and modified:

* `proto`: packet data and acknowledgement type and message type
* `x/module_name/keeper`: IBC hooks, gRPC message server
* `x/module_name/types`: message types, IBC events
* `x/module_name/client/cli`: CLI command to broadcast a transaction containing a message with a packet

## Example

```
starport packet buyOrder amountDenom amount:int priceDenom price:int --ack remainingAmount:int,purchase:int --module ibcdex
```
