# Step 03 Solution

この解答では outbound 変換関数を実装しています。

実装ポイント:
- 入力パケットをコピーして返す
- `Allocator` から変換後ポートを取得
- 学習用として末尾2byteに変換後ポートを書き込む

実行:

```bash
cd advanced/self-napt-go/solutions/step-03
go test ./...
```
