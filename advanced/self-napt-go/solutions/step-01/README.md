# Step 01 Solution

この解答では `PortAllocator` を完成させています。

実装ポイント:
- `Acquire`: 範囲内を循環しながら未使用ポートを探索
- 枯渇時: `port range exhausted` エラー
- `Release`: 使用中集合から解放

実行:

```bash
cd advanced/self-napt-go/solutions/step-01
go test ./...
```
