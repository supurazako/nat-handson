# Step 05: Timeout / GC

目的:
- NAT mapping を状態別に期限切れ削除する

課題:
- `nat/sweep.go` の `Sweep` を実装

実行:

```bash
cd advanced/self-napt-go/workshop/step-05-timeout
go test ./...
```

成功条件:
- `SYN_SENT` は短いタイムアウトで削除される
- `ESTABLISHED` は長いタイムアウトで削除される
