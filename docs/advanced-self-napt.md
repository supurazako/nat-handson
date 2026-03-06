# Advanced Track: Self NAPT Workshop (Go)

このトラックは、iptables の設定観察ではなく、NAPT の中身を段階的に実装して理解するワークショップです。

## 1. 対象

- 「設定して通った」より「自分でコードを書いて動かしたい」人
- `FlowKey`, `NAT table`, `port allocator`, `reverse lookup` を自分で実装したい人

## 2. 前提

- Linux または Linux VM + Docker
- Go 1.22 以上
- `docker-compose.self.yml` を使える環境

## 3. 進行（Step制）

`advanced/self-napt-go/workshop` を順番に進めます。

1. `step-00-base`: 環境確認
2. `step-01-allocator`: 変換ポート採番
3. `step-02-table`: mapping と reverse lookup
4. `step-03-outbound`: 送信方向変換
5. `step-04-inbound`: 返信方向逆変換
6. `step-05-timeout`: 状態ごとの期限切れ
7. `step-06-integration`: 統合と Step1完全再現の確認

各stepで:

```bash
cd advanced/self-napt-go/workshop/<step-dir>
go test ./...
```

注:
- `step-01` 以降は最初はテストが失敗します。TODOを埋めて緑化する進行です。

## 4. 答え合わせ

- `advanced/self-napt-go/solutions/` を参照
- step別解答:
  - `solutions/step-01`: allocator
  - `solutions/step-02`: table
  - `solutions/step-03`: outbound
  - `solutions/step-04`: inbound
  - `solutions/step-05`: timeout
- `solutions/step-06` は統合実装です

## 5. 統合実行（step-06）

```bash
cd advanced/self-napt-go
docker compose -f docker-compose.self.yml up -d --build
docker compose -f docker-compose.self.yml logs -f naptd
```

補助的な観察ポイント:
- `startup`
- `create_mapping`
- `translate_outbound`
- `translate_inbound`
- `mapping_expired`

## 6. 完了条件（DoD: Step1完全再現）

上級トラックの完了は「ログが出た」ではなく、`docs/instructions.md` の Step1 判定を満たすことです。  
以下をすべて満たしたら完了です。

1. `client` から `curl http://172.31.10.2` が成功する
2. `naptd` の `conntrack -L` に `src=192.168.20.2` と `dst=172.31.10.2 dport=80` が見える
3. `iptables -t nat -L -n -v` で `MASQUERADE` の `pkts/bytes` が増える
4. `iptables -L -n -v` で `FORWARD` の LAN->WAN / WAN->LAN(ESTABLISHED,RELATED) カウンタが増える
5. Step1-7（Before / After 比較）の「成功時」条件に一致する

確認コマンド:

```bash
docker compose exec client sh -c "apk add --no-cache curl >/dev/null 2>&1 || true; curl http://172.31.10.2"
docker compose exec naptd conntrack -L
docker compose exec naptd iptables -t nat -L -n -v
docker compose exec naptd iptables -L -n -v
```

## 7. 現在のスコープ

- まずは「最小実装 + 観察」を目的にしています
- 実パケット配線の強化（TUN read/write の完全配線や複数クライアント制御）は次フェーズで拡張します

## 8. 既存トラックとの住み分け

- 基本トラック: [docs/instructions.md](instructions.md)
- 上級ワークショップ: `advanced/self-napt-go/workshop`
