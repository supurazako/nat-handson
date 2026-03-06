# Step 04 Solution

この解答では inbound 逆変換関数を実装しています。

実装ポイント:
- `lookup` 成功時: 変換済みパケットを返し `dropped=false`
- `lookup` 失敗時: `dropped=true`, `error=nil`
- 学習用として末尾2byteに復元先ポートを書き込む

実行:

```bash
cd advanced/self-napt-go/solutions/step-04
go test ./...
```
