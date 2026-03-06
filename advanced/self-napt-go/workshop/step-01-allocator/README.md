# Step 01: PortAllocator

目的:
- NAPTで使う変換後ポートの採番器を実装する

課題:
- `nat/allocator.go` の `Acquire` / `Release` を実装

実行:

```bash
cd advanced/self-napt-go/workshop/step-01-allocator
go test ./...
```

成功条件:
- 枯渇・再利用を含むテストが通る

ヒント:
- `min..max` を循環する線形探索で十分
