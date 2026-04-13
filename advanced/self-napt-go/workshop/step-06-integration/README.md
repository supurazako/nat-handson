# Step 06: Integration

目的:
- ここまで作った部品をつないで、`docs/instructions.md` Step1 の成功判定を満たす

課題:
- `main.go` で table/allocator/forwarder/GC を組み立てる
- JSONログで `create_mapping`, `translate_outbound`, `translate_inbound`, `mapping_expired` を出す

実行例（起動）:

```bash
cd advanced/self-napt-go
docker compose -f docker-compose.self.yml up -d --build
docker compose -f docker-compose.self.yml logs -f naptd
```

完了条件（必須）:

1. `client -> server` の `curl http://172.31.10.2` が成功
2. `conntrack -L` に `src=192.168.20.2` と `dst=172.31.10.2 dport=80`
3. `MASQUERADE` カウンタ増加
4. `FORWARD` 2ルールカウンタ増加
5. `docs/instructions.md` Step1-7 の成功時条件と一致

確認コマンド:

```bash
docker compose -f docker-compose.self.yml exec client sh -c "apk add --no-cache curl >/dev/null 2>&1 || true; curl http://172.31.10.2"
docker compose -f docker-compose.self.yml exec naptd conntrack -L
docker compose -f docker-compose.self.yml exec naptd iptables -t nat -L -n -v
docker compose -f docker-compose.self.yml exec naptd iptables -L -n -v
```

補足:
- `naptd` のJSONログはデバッグ用です。完了判定は上記5項目です。

答え合わせ:
- `../../solutions/step-06` を参照
