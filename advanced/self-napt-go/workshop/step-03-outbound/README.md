# Step 03: Outbound Translation

目的:
- LAN->WAN の送信方向で src IP/port 変換を実装する

課題:
- `forwarder/forwarder.go` の `TranslateOutbound` を実装

実行:

```bash
cd advanced/self-napt-go/workshop/step-03-outbound
go test ./...
```

成功条件:
- 送信元ポートが変換後ポートに書き換わる
