# Step 02 Solution

この解答では NAT テーブルの基本操作を完成させています。

実装ポイント:
- `Upsert`: 新規作成または `LastSeen` 更新
- `GetByFlow`: flow キーで取得
- `GetByReverse`: reverse キーから逆引き
- `DeleteByFlow`: flow と reverse を同時削除

実行:

```bash
cd advanced/self-napt-go/solutions/step-02
go test ./...
```
