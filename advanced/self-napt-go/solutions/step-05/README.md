# Step 05 Solution

この解答では状態別タイムアウト判定 (`Sweep`) を実装しています。

実装ポイント:
- `StateSYNSent` は `synTimeout`
- `StateEstablished` は `establishedTimeout`
- 期限切れを `expired`、それ以外を `alive` に分離

実行:

```bash
cd advanced/self-napt-go/solutions/step-05
go test ./...
```
