# Step 04: Inbound Translation

目的:
- WAN->LAN の返信方向で逆変換を実装する

課題:
- `forwarder/forwarder.go` の `TranslateInbound` を実装

実行:

```bash
cd advanced/self-napt-go/workshop/step-04-inbound
go test ./...
```

成功条件:
- 追跡済みフローは dst が元クライアントに戻る
- 未追跡フローは drop 扱いになる
