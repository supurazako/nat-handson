# self-napt-go

Go で User-space NAT を自作する上級ワークショップです。

## 進め方

`workshop/` 配下を順に進めます。

1. `step-00-base`
2. `step-01-allocator`
3. `step-02-table`
4. `step-03-outbound`
5. `step-04-inbound`
6. `step-05-timeout`
7. `step-06-integration`

各stepで `go test ./...` を実行し、TODOを埋めてテストを通します。
ただし最終完了は「テスト通過」だけではなく、`docs/instructions.md` Step1 の成功判定を満たすことです。

## 答え合わせ

`solutions/` を参照してください。

- `solutions/step-06` は統合実装（prototype相当）です。

## 統合実行（step-06）

```bash
docker compose -f docker-compose.self.yml up -d --build
docker compose -f docker-compose.self.yml logs -f naptd
```

完了判定（必須）:
- `client -> server` の `curl` 成功（`http://172.31.10.2`）
- `conntrack` で `192.168.20.2 -> 172.31.10.2:80` フロー確認
- `MASQUERADE` カウンタ増加
- `FORWARD` 2ルールのカウンタ増加

詳細は `../../docs/advanced-self-napt.md` を参照してください。
