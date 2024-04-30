# Multi-Shot Non-Dedicated Sender Broadcast

A sample implementation of a multi-shot, non-dedicated sender broadcast protocol.
The implementation is based on [BCB](/pkg/bcb/bcbmodule.go) but extends it by adding a `broadcastSender` and a `broadcastID` to the messages.

The implementation might be faulty, which is actually somewhat desired as it was created to experiment with mir-native protocol testing tools.
