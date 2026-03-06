# Step 02: NAT Table

目的:
- FlowKey をキーに Mapping を保持し、返信方向の逆引きを実装する

課題:
- `nat/table.go` の `Upsert`, `GetByFlow`, `GetByReverse`, `DeleteByFlow` を実装

実行:

```bash
cd advanced/self-napt-go/workshop/step-02-table
go test ./...
```

成功条件:
- 往路登録と復路逆引きのテストが通る
