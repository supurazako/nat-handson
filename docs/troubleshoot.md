# Troubleshoot

通信できないときは、次を上から順に確認してください。

## 1. コンテナが起動しているか

```bash
docker compose ps
```

`nat-client1` / `nat-client2` / `nat-router` / `nat-server` が `Up` であること。

## 2. router の IP forward が有効か

```bash
docker compose exec router sysctl net.ipv4.ip_forward
```

`net.ipv4.ip_forward = 1` であること。

## 3. client1 / client2 の経路が router を向いているか

```bash
docker compose exec client1 ip route
docker compose exec client2 ip route
```

どちらも `default via 192.168.10.1` になっていること。

## 4. router の IF を取り違えていないか

```bash
docker compose exec router ip addr
```

`192.168.10.1/24` 側が LAN IF、`172.31.0.1/24` 側が WAN IF です。  
`eth0/eth1` 固定で判断しないこと。

## 5. NAT/FORWARD ルールが入っているか

```bash
docker compose exec router iptables -t nat -L -n -v
docker compose exec router iptables -L -n -v
```

`POSTROUTING MASQUERADE` と `FORWARD` 許可ルールが存在し、カウンタが増えること。
`pkts/bytes` が増えていれば、該当ルールに実トラフィックが当たっています。
読み方の補足は `docs/instructions.md` の「Step0: 詳細解説: iptables」を参照してください。
実行結果サンプルは `docs/instructions.md` の「Step1: 6. router で conntrack と NAT テーブル確認」を参照してください。

## 6. conntrack エントリが作られているか

```bash
docker compose exec router conntrack -L
```

`src=192.168.10.2` から `dst=172.31.0.2` へ向かうセッションが見えること。
加えて `dst=172.31.0.1`（または `reply_dst=172.31.0.1` 系）が見えれば、NAT後（WAN側）の追跡も確認できます。
`src/dst/sport/dport` は通信の向きとポートを示します。
読み方の補足は `docs/instructions.md` の「Step0: 詳細解説: conntrack」を参照してください。
実行結果サンプルは `docs/instructions.md` の「Step1: 6. router で conntrack と NAT テーブル確認」を参照してください。

未成功時/成功時の目安:
- 未成功時: `conntrack` に `dst=172.31.0.2` が出ない、`MASQUERADE` と `FORWARD` カウンタが `0 0`
- 成功時: `dst=172.31.0.2` に加え `dst=172.31.0.1`（または `reply_dst=...`）が見える、`MASQUERADE` と `FORWARD` カウンタが増える
- 詳細比較は `docs/instructions.md` の「Step1: 7. 成功判定（Before / After 比較）」を参照

## 7. 複数クライアント時の確認（Step2）

```bash
docker compose exec router conntrack -L | grep 'src=192.168.10.2'
docker compose exec router conntrack -L | grep 'src=192.168.10.3'
```

`client1`（`192.168.10.2`）と `client2`（`192.168.10.3`）の両方のエントリが見えること。
必要なら `docker compose exec router conntrack -L | grep '172.31.0.1'` でNAT後（WAN側）の表示も確認すること。
実行結果サンプルは `docs/instructions.md` の「Step2: 4. router で複数フローを観察」を参照してください。
仕組みの説明は `docs/instructions.md` の「Step2: 6. なぜ1つのグローバルIPで同時通信できるのか」を参照してください。

## 8. server 単体疎通を確認

```bash
docker compose exec router wget -qO- http://172.31.0.2 | head -n 1
```

router から server に届かなければ、WAN 側接続か server 側の起動状態を確認すること。

## 9. Docker network のサブネット重複を確認（Compose側を変更）

`docker compose up -d` で以下のようなエラーが出る場合があります。  
`invalid pool request: Pool overlaps with other one on this address space`

現在使われているサブネットを確認し、`docker-compose.yml` の `wan` 側を重複しない帯域へ変更します。

```bash
docker network inspect $(docker network ls -q) --format '{{.Name}} {{range .IPAM.Config}}{{.Subnet}} {{end}}'
```

`docker-compose.yml` では、次の値を同じプレフィックスで揃えて変更します。

- `networks.wan.ipam.config.subnet`
- `networks.wan.ipam.config.gateway`
- `services.router.networks.wan.ipv4_address`
- `services.server.networks.wan.ipv4_address`

例: `172.31.0.0/24` が重複する場合は `172.30.0.0/24` など別の `/24` に変更。

変更後、再作成します。

```bash
docker compose down
docker compose up -d
```
