# Advanced Self NAPT Troubleshoot

## 1. stepで `go test` が失敗する

まず想定どおりです。`workshop` は TODO を埋める前提です。

確認:
- 対象stepの README を読み、実装対象の関数だけ先に埋める
- `cd advanced/self-napt-go/workshop/<step-dir> && go test ./...`

## 2. どのstepから始めるべきか分からない

順序固定で進めてください。

1. `step-00-base`
2. `step-01-allocator`
3. `step-02-table`
4. `step-03-outbound`
5. `step-04-inbound`
6. `step-05-timeout`
7. `step-06-integration`

## 3. 答え合わせしたい

- `advanced/self-napt-go/solutions/` を参照
- 現在のstepに対応する解答を参照
  - step-01 -> `solutions/step-01`
  - step-02 -> `solutions/step-02`
  - step-03 -> `solutions/step-03`
  - step-04 -> `solutions/step-04`
  - step-05 -> `solutions/step-05`
- 統合実装は `solutions/step-06`

## 4. 統合実行で `naptd` ログが出ない

```bash
cd advanced/self-napt-go
docker compose -f docker-compose.self.yml ps
docker compose -f docker-compose.self.yml logs -f naptd
```

`naptd` が起動していない場合:
- `docker compose -f docker-compose.self.yml up -d --build` を再実行

## 5. `curl` は成功するが完了条件を満たせない

症状:
- `curl http://172.31.10.2` は成功する
- ただし `conntrack` / `MASQUERADE` / `FORWARD` の観察が一致しない

確認:

```bash
docker compose -f docker-compose.self.yml exec naptd conntrack -L
docker compose -f docker-compose.self.yml exec naptd iptables -t nat -L -n -v
docker compose -f docker-compose.self.yml exec naptd iptables -L -n -v
```

観点:
- `conntrack` に `src=192.168.20.2` と `dst=172.31.10.2 dport=80` があるか
- `MASQUERADE` の `pkts/bytes` が増えているか
- `FORWARD` 2ルール（LAN->WAN / WAN->LAN ESTABLISHED,RELATED）が増えているか

## 6. `/dev/net/tun` が使えない

症状:
- `no such file or directory: /dev/net/tun`
- `operation not permitted`

対処:
- Linux環境か確認
- Docker実行ユーザー権限を確認
- `docker-compose.self.yml` の `devices` と `cap_add` を確認

## 7. 既存トラックと混ざって混乱する

切り分け:
- 基本トラック: `docker-compose.yml` + `docs/instructions.md`
- 上級トラック: `advanced/self-napt-go/workshop` + `docs/advanced-self-napt.md`
