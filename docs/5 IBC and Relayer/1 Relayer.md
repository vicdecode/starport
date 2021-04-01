# Relayer

A built-in IBC relayer in Starport lets you connect blockchains that run on your local computer and remote blockchains. The Starport relayer uses [IBC Go relayer](https://github.com/cosmos/relayer).

## `starport relayer configure`

The `configure` command configures a connection between two blockchains. You are prompted for the required RPC endpoints and optional faucet endpoints. Accounts used by the relayer are created on both blockchains and faucets are used, if available, to automatically fetch tokens. If the relayer fails to receive tokens from a faucet, you must manually send tokens to addresses. By default, a connection for token transfers is set up for the `ibc-transfer` module.

The optional `--advanced` flag lets you configure port and version for the custom IBC module.

By default, relayer configuration is stored in `$HOME/.relayer/`.

## Example

All values can be passed with flags.

```
starport relayer configure --advanced --source-rpc "http://0.0.0.0:26657" --source-faucet "http://0.0.0.0:4500" --source-port "blog" --source-version "blog-1" --target-rpc "http://0.0.0.0:26659" --target-faucet "http://0.0.0.0:4501" --target-port "blog" --target-version "blog-1"
```

## `starport relayer connect`

Connects configured blockchains and watches for IBC packets to relay.
